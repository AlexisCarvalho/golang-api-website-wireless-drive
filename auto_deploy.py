#!/usr/bin/env python3
"""
Build and deploy the WirelessDrive (Go) server to an Android device via adb.

Steps:
  1. Compile the Go server for the device's ABI (NDK)
  2. (optional) Compile static ffmpeg/ffprobe for the same ABI
  3. Send everything to the device (the server binary will be named .new and replaced with mv)
  4. Restart the server and confirm it is running correctly
  5. Show the local IP of the device (and test the port, if known)

Examples:
  ./auto_deploy.py                    # interactive mode
  ./auto_deploy.py --no-ffmpeg        # no prompts
  ./auto_deploy.py -s emulator-5554   # select the device
  ./auto_deploy.py --ffmpeg or --ffmpeg-minimal # compiles ffmpeg 
  (The --ffmpeg-minimal binary uses less storage space, but it 
  may have difficulty generating thumbnails for some types of 
  images and videos.)
"""

import argparse
import os
import re
import shlex
import shutil
import socket
import subprocess
import sys
import time

# ==========================
# CONFIG
# ==========================

NDK = os.environ.get(
    "ANDROID_NDK",
    os.path.expanduser("~/android-ndk-r29")
)

BIN_NAME = "WirelessDrive"
TARGET_DIR = "/data/local/tmp/wirelessDrive"
TARGET_BIN = f"{TARGET_DIR}/{BIN_NAME}"
TARGET_LOG = f"{TARGET_DIR}/server.log"

BUILD_OUT = "server"

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
FFMPEG_DIR = os.path.join(SCRIPT_DIR, "FFmpeg")
FFMPEG_BUILD_ROOT = os.path.join(SCRIPT_DIR, "ffmpeg-build")
FFMPEG_REPO = "https://github.com/FFmpeg/FFmpeg.git"
# Tag used only for cloning. If the FFmpeg folder already exists, it is used as is.
FFMPEG_TAG = os.environ.get("FFMPEG_TAG", "n7.1")

# ==========================
# COLORS
# ==========================

RESET = "\033[0m"
GREEN = "\033[92m"
BLUE = "\033[94m"
RED = "\033[91m"
YELLOW = "\033[93m"


def log(tag, msg, color):
    print(f"{color}[{tag}]{RESET} {msg}")


def die(msg):
    log("ERROR", msg, RED)
    sys.exit(1)


def run(cmd, env=None, cwd=None, capture=False, check=True, quiet=False, stdin=None):
    if not quiet:
        log("RUN", shlex.join(cmd), BLUE)

    try:
        return subprocess.run(
            cmd,
            env=env,
            cwd=cwd,
            check=check,
            text=True,
            capture_output=capture,
            stdin=stdin,
        )
    except subprocess.CalledProcessError as e:
        if capture:
            for stream in (e.stdout, e.stderr):
                if stream and stream.strip():
                    print(stream.rstrip())
        raise


def ask_yes_no(prompt, default_no=True):
    if not sys.stdin.isatty():
        return not default_no

    suffix = "[y/N]" if default_no else "[Y/n]"
    answer = input(f"{prompt} {suffix}: ").strip().lower()

    if not answer:
        return not default_no

    return answer in ("y", "yes", "s", "sim")


# ==========================
# ADB / DEVICE
# ==========================

DEVICE = None


def adb(args, **kwargs):
    return run(["adb", "-s", DEVICE] + args, **kwargs)


def remote_executable(path):
    result = adb(["shell", "test", "-x", path],
                 capture=True, check=False, quiet=True)
    return result.returncode == 0


def get_devices():
    result = run(["adb", "devices"], capture=True, quiet=True)

    ready, blocked = [], []

    for line in result.stdout.splitlines()[1:]:
        parts = line.split()

        if len(parts) < 2:
            continue

        if parts[1] == "device":
            ready.append(parts[0])
        else:
            blocked.append((parts[0], parts[1]))

    return ready, blocked


def select_device(requested):
    ready, blocked = get_devices()

    for serial, state in blocked:
        log("WARN", f"Ignoring {serial} ({state})", YELLOW)

    requested = requested or os.environ.get("ANDROID_SERIAL")

    if requested:
        if requested not in ready:
            die(f"Device '{requested}' not found or not ready.")
        return requested

    if not ready:
        die("No Android device connected.")

    if len(ready) == 1:
        return ready[0]

    if not sys.stdin.isatty():
        die("Multiple devices detected; use --device SERIAL.")

    print()
    log("INFO", "Multiple devices detected:", YELLOW)

    for i, device in enumerate(ready, 1):
        print(f"  {i}. {device}")

    while True:
        try:
            choice = int(input("\nSelect device: "))

            if 1 <= choice <= len(ready):
                return ready[choice - 1]

            print("Invalid option.")
        except ValueError:
            print("Enter a valid number.")


# ==========================
# NDK / ARCH
# ==========================

def ndk_toolchain():
    if sys.platform.startswith("linux"):
        host = "linux-x86_64"
    elif sys.platform == "darwin":
        host = "darwin-x86_64"
    else:
        die(f"Unsupported host platform: {sys.platform}")

    toolchain = f"{NDK}/toolchains/llvm/prebuilt/{host}/bin"

    if not os.path.isdir(toolchain):
        die(f"NDK toolchain not found at {toolchain}\n"
            "       Set ANDROID_NDK to your NDK folder.")

    return toolchain


def build_arch_table(toolchain):
    return {
        "arm64": {
            "goarch": "arm64",
            "goarm": None,
            "cc": f"{toolchain}/aarch64-linux-android21-clang",
            "cxx": f"{toolchain}/aarch64-linux-android21-clang++",
            "ffmpeg_arch": "aarch64",
            "ffmpeg_cpu": "armv8-a",
        },
        "armeabi": {
            "goarch": "arm",
            "goarm": "7",
            "cc": f"{toolchain}/armv7a-linux-androideabi21-clang",
            "cxx": f"{toolchain}/armv7a-linux-androideabi21-clang++",
            "ffmpeg_arch": "arm",
            "ffmpeg_cpu": "armv7-a",
        },
        "x86_64": {
            "goarch": "amd64",
            "goarm": None,
            "cc": f"{toolchain}/x86_64-linux-android21-clang",
            "cxx": f"{toolchain}/x86_64-linux-android21-clang++",
            "ffmpeg_arch": "x86_64",
            "ffmpeg_cpu": "x86-64",
        },
        "x86": {
            "goarch": "386",
            "goarm": None,
            "cc": f"{toolchain}/i686-linux-android21-clang",
            "cxx": f"{toolchain}/i686-linux-android21-clang++",
            "ffmpeg_arch": "x86",
            "ffmpeg_cpu": "i686",
        },
    }


def pick_arch(table, abi):
    for key, info in table.items():
        if abi.startswith(key):
            return key, info

    die(f"Unsupported ABI: {abi}")


# ==========================
# GO BUILD
# ==========================

def build_server(arch_info):
    env = os.environ.copy()
    env["ANDROID_NDK"] = NDK
    env["GOOS"] = "android"
    env["CGO_ENABLED"] = "1"
    env["GOARCH"] = arch_info["goarch"]
    env["CC"] = arch_info["cc"]
    env["CXX"] = arch_info["cxx"]

    if arch_info["goarm"]:
        env["GOARM"] = arch_info["goarm"]
    else:
        env.pop("GOARM", None)

    log("INFO", f"GOARCH={env['GOARCH']}", GREEN)
    log("BUILD", "Compiling...", GREEN)

    if os.path.exists(BUILD_OUT):
        os.remove(BUILD_OUT)

    try:
        run([
            "go", "build", "-v",
            "-trimpath",
            "-ldflags", "-s -w",
            "-o", BUILD_OUT,
            "./cmd/server",
        ], env=env)
    except subprocess.CalledProcessError:
        die("Build failed. Deployment cancelled.")

    if not os.path.exists(BUILD_OUT):
        die("Executable not found.")

    log("BUILD", "Build successful.", GREEN)


# ==========================
# FFMPEG / FFPROBE (OPCIONAL)
# ==========================

# Components used by the server to generate thumbnails.
FFMPEG_COMPONENTS = [
    "--disable-shared",
    "--enable-static",
    "--disable-doc",
    "--disable-debug",

    "--enable-ffmpeg",
    "--enable-ffprobe",
    "--disable-ffplay",

    "--enable-avcodec",
    "--enable-avformat",
    "--enable-avutil",
    "--enable-swscale",

    "--enable-protocol=file",

    "--enable-demuxer=mov",
    "--enable-demuxer=matroska",
    "--enable-demuxer=avi",
    "--enable-demuxer=image2",
    "--enable-demuxer=webm_dash_manifest",
    "--enable-demuxer=gif",

    "--enable-muxer=image2",

    "--enable-parser=h264",
    "--enable-parser=hevc",
    "--enable-parser=vp8",
    "--enable-parser=vp9",
    "--enable-parser=av1",

    "--enable-decoder=h264",
    "--enable-decoder=hevc",
    "--enable-decoder=vp8",
    "--enable-decoder=vp9",
    "--enable-decoder=av1",

    "--enable-decoder=mjpeg",
    "--enable-decoder=png",
    "--enable-decoder=webp",
    "--enable-decoder=gif",

    "--enable-encoder=mjpeg",

    "--enable-filter=scale",
    "--enable-filter=format",
]

# Used only with --ffmpeg-minimal (comes AFTER --disable-everything).
FFMPEG_MINIMAL_EXTRA = [
    "--enable-avfilter",
    "--enable-swresample",

    "--enable-protocol=pipe",

    "--enable-muxer=mjpeg",
    "--enable-muxer=image2pipe",
    "--enable-muxer=null",

    "--enable-demuxer=image2pipe",
    "--enable-demuxer=mp3",
    "--enable-demuxer=flac",
    "--enable-demuxer=ogg",
    "--enable-demuxer=wav",
    "--enable-demuxer=aac",
    "--enable-demuxer=mpegts",
    "--enable-demuxer=flv",

    "--enable-filter=buffer",
    "--enable-filter=buffersink",
    "--enable-filter=abuffer",
    "--enable-filter=abuffersink",
    "--enable-filter=null",
    "--enable-filter=anull",

    "--enable-zlib",
    "--enable-small",
    "--disable-avdevice",
    "--disable-network",
    "--disable-autodetect",
]


def ffmpeg_configure_args(info, toolchain, minimal):
    args = [
        "--target-os=android",
        f"--arch={info['ffmpeg_arch']}",
        f"--cpu={info['ffmpeg_cpu']}",
        f"--cc={info['cc']}",
        f"--cxx={info['cxx']}",
        f"--ar={toolchain}/llvm-ar",
        f"--nm={toolchain}/llvm-nm",
        f"--ranlib={toolchain}/llvm-ranlib",
        f"--strip={toolchain}/llvm-strip",
        "--enable-cross-compile",
    ]

    if info["ffmpeg_arch"] == "x86":
        args.append("--disable-asm")

    if minimal:
        args.append("--disable-everything")

    args += FFMPEG_COMPONENTS

    if minimal:
        args += FFMPEG_MINIMAL_EXTRA

    return args


def ensure_ffmpeg_source():
    if os.path.isdir(FFMPEG_DIR):
        return True

    log("WARN", f"FFmpeg source folder not found at: {FFMPEG_DIR}", YELLOW)

    if shutil.which("git") is None:
        log("ERROR", "git not found in PATH. Cannot clone FFmpeg. Skipping.", RED)
        return False

    if not ask_yes_no(f"Clone FFmpeg ({FFMPEG_TAG}) from GitHub into that folder now?"):
        log("INFO", "Skipping FFmpeg/ffprobe deployment.", YELLOW)
        return False

    try:
        run([
            "git", "clone", "--depth", "1", "--branch", FFMPEG_TAG,
            FFMPEG_REPO, FFMPEG_DIR,
        ])
    except subprocess.CalledProcessError:
        log("ERROR", "Failed to clone FFmpeg. Skipping.", RED)
        return False

    return True


def build_ffmpeg(arch_key, info, toolchain, minimal, rebuild):
    """Compila ffmpeg/ffprobe fora da árvore de fontes, uma pasta por ABI.

    Retorna (ffmpeg_path, ffprobe_path) ou None se falhar.
    """

    build_dir = os.path.join(FFMPEG_BUILD_ROOT, arch_key)
    ffmpeg_bin = os.path.join(build_dir, "ffmpeg")
    ffprobe_bin = os.path.join(build_dir, "ffprobe")
    stamp_file = os.path.join(build_dir, ".build-stamp")

    args = ffmpeg_configure_args(info, toolchain, minimal)

    head = run(["git", "rev-parse", "HEAD"], cwd=FFMPEG_DIR,
               capture=True, check=False, quiet=True).stdout.strip() or "unknown"
    stamp = head + "\n" + " ".join(args)

    # If nothing has changed it reuses the previous build  (same version and same flags).
    if (not rebuild
            and os.path.exists(ffmpeg_bin)
            and os.path.exists(ffprobe_bin)
            and os.path.exists(stamp_file)):
        with open(stamp_file, encoding="utf-8") as f:
            if f.read() == stamp:
                log("BUILD", f"FFmpeg for {arch_key} is up to date; reusing build.", GREEN)
                return ffmpeg_bin, ffprobe_bin

    shutil.rmtree(build_dir, ignore_errors=True)
    os.makedirs(build_dir)

    if os.path.exists(os.path.join(FFMPEG_DIR, "ffbuild", "config.mak")):
        log("BUILD", "Cleaning previous in-tree build (make distclean)...", GREEN)
        run(["make", "distclean"], cwd=FFMPEG_DIR, check=False)

    log("BUILD",
        f"Configuring FFmpeg (arch={info['ffmpeg_arch']}, cpu={info['ffmpeg_cpu']}"
        f"{', minimal' if minimal else ''})...", GREEN)

    try:
        run([os.path.join(FFMPEG_DIR, "configure")] + args, cwd=build_dir)
    except subprocess.CalledProcessError:
        log("ERROR", "FFmpeg configure failed. Skipping FFmpeg/ffprobe deployment.", RED)
        return None

    nproc = str(os.cpu_count() or 4)

    log("BUILD", f"Building FFmpeg with -j{nproc} (this can take a while)...", GREEN)

    try:
        run(["make", f"-j{nproc}"], cwd=build_dir)
    except subprocess.CalledProcessError:
        log("ERROR", "FFmpeg build failed. Skipping FFmpeg/ffprobe deployment.", RED)
        return None

    if not os.path.exists(ffmpeg_bin) or not os.path.exists(ffprobe_bin):
        log("ERROR", "ffmpeg/ffprobe binaries not found after build.", RED)
        return None

    with open(stamp_file, "w", encoding="utf-8") as f:
        f.write(stamp)

    return ffmpeg_bin, ffprobe_bin


def maybe_build_ffmpeg(args, arch_key, arch_info, toolchain):
    """Decide se o ffmpeg entra neste deploy; devolve os binários ou None."""

    if args.ffmpeg is False:
        log("INFO", "Skipping FFmpeg/ffprobe deployment.", YELLOW)
        return None

    if args.ffmpeg is None:
        already = (remote_executable(f"{TARGET_DIR}/ffmpeg")
                   and remote_executable(f"{TARGET_DIR}/ffprobe"))

        print()
        log("INFO", "FFmpeg/ffprobe deployment is optional and separate from the server deploy.", YELLOW)
        log("INFO", "You should (re)run it whenever you switch to a device with a different CPU "
                    "architecture, or if it has never been done on the current device before.", YELLOW)
        log("INFO", "If skipped, native thumbnail generation won't be available on the device: the "
                    "app will have to rely on an external API for thumbnails, or won't generate "
                    "thumbnails for uploaded media at all.", YELLOW)

        if already:
            log("INFO", "ffmpeg/ffprobe are already present on this device.", GREEN)

        if not ask_yes_no("Build and deploy FFmpeg/ffprobe now?"):
            log("INFO", "Skipping FFmpeg/ffprobe deployment.", YELLOW)
            return None

    if shutil.which("make") is None:
        log("ERROR", "make not found in PATH. Cannot build FFmpeg.", RED)
        return None

    if not ensure_ffmpeg_source():
        return None

    return build_ffmpeg(
        arch_key, arch_info, ndk_toolchain(),
        minimal=args.ffmpeg_minimal,
        rebuild=args.ffmpeg_rebuild,
    )


# ==========================
# DEPLOY
# ==========================

def server_pids():
    result = adb(["shell", "pidof", BIN_NAME],
                 capture=True, check=False, quiet=True)
    return result.stdout.split()


def stop_server():
    pids = server_pids()

    if not pids:
        log("INFO", "No running server found.", YELLOW)
        return

    log("RUN", f"Stopping PID(s) {' '.join(pids)}", YELLOW)
    adb(["shell", "kill"] + pids, check=False)

    # kill is asynchronous: it waits for the process to terminate on its own
    for _ in range(10):
        time.sleep(0.5)
        if not server_pids():
            return

    log("WARN", "Server did not stop in time; sending SIGKILL.", YELLOW)
    adb(["shell", "kill", "-9"] + server_pids(), check=False)
    time.sleep(0.5)


def push_files(ffmpeg_bins):
    adb(["shell", "mkdir", "-p", TARGET_DIR])

# The server is initially built as `.new` and renamed later. 
# This approach prevents push conflicts with "Text file busy" errors, 
# which occur if the old binary is still running.
    log("PUSH", "Uploading executable...", GREEN)
    adb(["push", BUILD_OUT, f"{TARGET_BIN}.new"])
    adb(["shell", "chmod", "755", f"{TARGET_BIN}.new"])

    if os.path.exists(".env"):
        log("PUSH", "Uploading .env...", GREEN)
        adb(["push", ".env", f"{TARGET_DIR}/.env"])
        adb(["shell", "chmod", "600", f"{TARGET_DIR}/.env"])

    if ffmpeg_bins:
        for path in ffmpeg_bins:
            name = os.path.basename(path)
            log("PUSH", f"Uploading {name}...", GREEN)
            adb(["push", path, f"{TARGET_DIR}/{name}"])
            adb(["shell", "chmod", "755", f"{TARGET_DIR}/{name}"])


def start_server():
    log("RUN", "Starting server...", GREEN)

    adb([
        "shell",
        f"cd {TARGET_DIR} && (setsid nohup ./{BIN_NAME} >server.log 2>&1 </dev/null &)",
    ], stdin=subprocess.DEVNULL)

    time.sleep(1.5)

    if server_pids():
        log("INFO", f"Server is running (PID {' '.join(server_pids())}).", GREEN)
        return True

    log("ERROR", "Server is not running after start. Last log lines:", RED)
    adb(["shell", "tail", "-n", "20", TARGET_LOG], check=False)
    return False


# ==========================
# DEVICE LOCAL IP / PORT
# ==========================

def get_device_ip():
    """Best-effort lookup of the device's local network IP.

    Tries `ip route get` first (works regardless of interface name, e.g.
    wlan0 vs eth0), falling back to scanning `ip addr` for a private IPv4.
    """

    result = adb(["shell", "ip", "route", "get", "1.1.1.1"],
                 capture=True, check=False, quiet=True)
    match = re.search(r"src\s+(\d+\.\d+\.\d+\.\d+)", result.stdout)
    if match:
        return match.group(1)

    result = adb(["shell", "ip", "-f", "inet", "addr", "show"],
                 capture=True, check=False, quiet=True)
    for match in re.finditer(r"inet\s+(\d+\.\d+\.\d+\.\d+)/\d+", result.stdout):
        ip = match.group(1)
        if not ip.startswith("127."):
            return ip

    return None


def find_port(cli_port):
    if cli_port:
        return cli_port

    if os.path.exists(".env"):
        with open(".env", encoding="utf-8", errors="ignore") as f:
            for line in f:
                m = re.match(
                    r"\s*(?:export\s+)?(?:PORT|SERVER_PORT|HTTP_PORT)\s*=\s*['\"]?(\d+)",
                    line,
                )
                if m:
                    return int(m.group(1))

    return None


def port_open(ip, port, timeout=3.0):
    try:
        with socket.create_connection((ip, port), timeout=timeout):
            return True
    except OSError:
        return False


def report_address(cli_port):
    print()
    log("INFO", "Looking up device IP on the local network...", BLUE)

    ip = get_device_ip()

    if not ip:
        log("WARN", "Could not determine the device's local IP.", YELLOW)
        return

    log("INFO", f"Device local IP: {ip}", GREEN)

    port = find_port(cli_port)

    if not port:
        return

    log("INFO", f"Server URL: http://{ip}:{port}", GREEN)

    if port_open(ip, port):
        log("INFO", f"Port {port} is accepting connections.", GREEN)
    else:
        log("WARN", f"Could not connect to {ip}:{port} from this machine "
                    "(different network, firewall, or the server is still starting).", YELLOW)


# ==========================
# MAIN
# ==========================

def parse_args():
    p = argparse.ArgumentParser(description="Build and deploy WirelessDrive to an Android device.")

    p.add_argument("-s", "--device", help="adb serial (default: ANDROID_SERIAL or prompt)")
    p.add_argument("--port", type=int, help="server port, used only to print the URL/test the connection")
    p.add_argument("--no-restart", action="store_true",
                   help="upload the new binary but do not stop/start the server")

    g = p.add_mutually_exclusive_group()
    g.add_argument("--ffmpeg", dest="ffmpeg", action="store_true", default=None,
                   help="build and deploy ffmpeg/ffprobe without asking")
    g.add_argument("--no-ffmpeg", dest="ffmpeg", action="store_false",
                   help="skip ffmpeg/ffprobe without asking")

    p.add_argument("--ffmpeg-minimal", action="store_true",
                   help="use --disable-everything with an explicit component list (smaller binary; test it)")
    p.add_argument("--ffmpeg-rebuild", action="store_true",
                   help="ignore the cached ffmpeg build for this ABI")

    return p.parse_args()


def main():
    global DEVICE

    args = parse_args()

    for tool in ("go", "adb"):
        if shutil.which(tool) is None:
            die(f"{tool} not found in PATH")

    toolchain = ndk_toolchain()

    DEVICE = select_device(args.device)
    log("INFO", f"Using device: {DEVICE}", GREEN)

    try:
        abi = adb(["shell", "getprop", "ro.product.cpu.abi"],
                  capture=True).stdout.strip()
    except subprocess.CalledProcessError:
        die("Failed to communicate with the device.")

    log("INFO", f"Device ABI: {abi}", GREEN)

    arch_key, arch_info = pick_arch(build_arch_table(toolchain), abi)

    build_server(arch_info)

    ffmpeg_bins = maybe_build_ffmpeg(args, arch_key, arch_info, toolchain)

    push_files(ffmpeg_bins)

    if not args.no_restart:
        stop_server()

    adb(["shell", "mv", "-f", f"{TARGET_BIN}.new", TARGET_BIN])

    if args.no_restart:
        log("INFO", "Binary updated; server was not restarted (--no-restart).", YELLOW)
    elif not start_server():
        sys.exit(1)

    report_address(args.port)

    log("DONE", "Deployment finished successfully.", GREEN)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print()
        die("Interrupted.")
