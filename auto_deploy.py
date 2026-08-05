#!/usr/bin/env python3

import os
import shutil
import subprocess
import sys

# ==========================
# CONFIG
# ==========================

NDK = os.environ.get(
    "ANDROID_NDK",
    os.path.expanduser("~/android-ndk-r29")
)

TARGET_DIR = "/data/local/tmp/wirelessDrive"
TARGET_BIN = f"{TARGET_DIR}/WirelessDrive"

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
FFMPEG_DIR = os.path.join(SCRIPT_DIR, "FFmpeg")
FFMPEG_REPO = "https://github.com/FFmpeg/FFmpeg.git"

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


def run(cmd, env=None, cwd=None, capture=False, check=True):
    log("RUN", " ".join(cmd), BLUE)

    return subprocess.run(
        cmd,
        env=env,
        cwd=cwd,
        check=check,
        text=True,
        capture_output=capture,
    )


def ask_yes_no(prompt, default_no=True):
    suffix = "[y/N]" if default_no else "[Y/n]"

    answer = input(f"{prompt} {suffix}: ").strip().lower()

    if not answer:
        return not default_no

    return answer in ("y", "yes", "s", "sim")


# ==========================
# CHECK TOOLS
# ==========================

for tool in ("go", "adb"):
    if shutil.which(tool) is None:
        log("ERROR", f"{tool} not found in PATH", RED)
        sys.exit(1)


# ==========================
# DEVICE SELECTION
# ==========================

def get_devices():
    result = run(["adb", "devices"], capture=True)

    devices = []

    for line in result.stdout.splitlines()[1:]:
        line = line.strip()
        if not line:
            continue

        parts = line.split()

        if len(parts) >= 2 and parts[1] == "device":
            devices.append(parts[0])

    return devices


devices = get_devices()

if not devices:
    log("ERROR", "No Android device connected.", RED)
    sys.exit(1)

if len(devices) == 1:
    DEVICE = devices[0]
    log("INFO", f"Using device: {DEVICE}", GREEN)
else:
    print()

    log("INFO", "Multiple devices detected:", YELLOW)

    for i, device in enumerate(devices, 1):
        print(f"  {i}. {device}")

    while True:
        try:
            choice = int(input("\nSelect device: "))

            if 1 <= choice <= len(devices):
                DEVICE = devices[choice - 1]
                break

            print("Invalid option.")
        except ValueError:
            print("Enter a valid number.")

    log("INFO", f"Using device: {DEVICE}", GREEN)


def adb(args, capture=False):
    return run(
        ["adb", "-s", DEVICE] + args,
        capture=capture,
    )


# ==========================
# DETECT DEVICE ARCH
# ==========================

try:
    abi = adb(
        ["shell", "getprop", "ro.product.cpu.abi"],
        capture=True
    ).stdout.strip()
except subprocess.CalledProcessError:
    log("ERROR", "Failed to communicate with the device.", RED)
    sys.exit(1)

log("INFO", f"Device ABI: {abi}", GREEN)

toolchain = f"{NDK}/toolchains/llvm/prebuilt/linux-x86_64/bin"

# Per-arch settings shared between the Go build and the FFmpeg build.
# ffmpeg_arch / ffmpeg_cpu follow FFmpeg's ./configure --arch / --cpu flags.
ARCH_TABLE = {
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

arch_info = None

for prefix, info in ARCH_TABLE.items():
    if abi.startswith(prefix):
        arch_info = info
        break

if arch_info is None:
    log("ERROR", f"Unsupported ABI: {abi}", RED)
    sys.exit(1)

env = os.environ.copy()
env["ANDROID_NDK"] = NDK
env["GOOS"] = "android"
env["CGO_ENABLED"] = "1"
env["GOARCH"] = arch_info["goarch"]
env["CC"] = arch_info["cc"]

if arch_info["goarm"]:
    env["GOARM"] = arch_info["goarm"]
else:
    env.pop("GOARM", None)

log("INFO", f"GOARCH={env['GOARCH']}", GREEN)

# ==========================
# BUILD
# ==========================

log("BUILD", "Compiling...", GREEN)

try:
    run(["go", "build", "-v", "./cmd/server"], env=env)
except subprocess.CalledProcessError:
    log("ERROR", "Build failed. Deployment cancelled.", RED)
    sys.exit(1)

if not os.path.exists("server"):
    log("ERROR", "Executable not found.", RED)
    sys.exit(1)

log("BUILD", "Build successful.", GREEN)

# ==========================
# STOP OLD SERVER
# ==========================

result = adb(
    ["shell", "ps", "-A"],
    capture=True
)

for line in result.stdout.splitlines():
    if "WirelessDrive" in line:
        pid = line.split()[1]
        log("RUN", f"Stopping PID {pid}", YELLOW)
        adb(["shell", "kill", pid])

# ==========================
# CREATE DIRECTORY
# ==========================

adb([
    "shell",
    "mkdir",
    "-p",
    TARGET_DIR
])

# ==========================
# PUSH FILES
# ==========================

log("PUSH", "Uploading executable...", GREEN)

adb([
    "push",
    "server",
    TARGET_BIN
])

if os.path.exists(".env"):
    log("PUSH", "Uploading .env...", GREEN)
    adb([
        "push",
        ".env",
        f"{TARGET_DIR}/.env"
    ])

# ==========================
# START SERVER
# ==========================

log("RUN", "Starting server...", GREEN)

adb([
    "shell",
    "",  # If it doesn't run, try adding "sh"
    "",  # and "-c". For me it wasn't needed
    f"nohup {TARGET_BIN} >/dev/null 2>&1 &"
])

log("DONE", "Deployment finished successfully.", GREEN)


# ==========================
# FFMPEG / FFPROBE (OPTIONAL)
# ==========================

def build_ffmpeg(info):
    """Cross-compile a minimal static ffmpeg/ffprobe for the device's ABI."""

    ffmpeg_env = os.environ.copy()
    ffmpeg_cc = info["cc"]
    ffmpeg_cxx = info["cxx"]
    ffmpeg_env["AR"] = f"{toolchain}/llvm-ar"
    ffmpeg_env["LD"] = f"{toolchain}/ld.lld"
    ffmpeg_env["STRIP"] = f"{toolchain}/llvm-strip"

    # Architecture may differ from the last build (e.g. switched devices),
    # so wipe any previous build artifacts first.
    log("BUILD", "Running make distclean...", GREEN)
    run(["make", "distclean"], cwd=FFMPEG_DIR, check=False)

    configure_cmd = [
        "./configure",
        "--target-os=android",
        f"--arch={info['ffmpeg_arch']}",
        f"--cpu={info['ffmpeg_cpu']}",
        f"--cc={ffmpeg_cc}",
        f"--cxx={ffmpeg_cxx}",
        f"--ar={ffmpeg_env['AR']}",
        f"--strip={ffmpeg_env['STRIP']}",
        "--enable-cross-compile",

        "--disable-shared",
        "--enable-static",

        "--disable-everything",

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

    log("BUILD", f"Configuring FFmpeg (arch={info['ffmpeg_arch']}, cpu={info['ffmpeg_cpu']})...", GREEN)

    try:
        run(configure_cmd, env=ffmpeg_env, cwd=FFMPEG_DIR)
    except subprocess.CalledProcessError:
        log("ERROR", "FFmpeg configure failed. Skipping FFmpeg/ffprobe deployment.", RED)
        return False

    nproc = str(os.cpu_count() or 4)

    log("BUILD", f"Building FFmpeg with -j{nproc} (this can take a while)...", GREEN)

    try:
        run(["make", f"-j{nproc}"], env=ffmpeg_env, cwd=FFMPEG_DIR)
    except subprocess.CalledProcessError:
        log("ERROR", "FFmpeg build failed. Skipping FFmpeg/ffprobe deployment.", RED)
        return False

    return True


def deploy_ffmpeg():
    if shutil.which("make") is None:
        log("ERROR", "make not found in PATH. Cannot build FFmpeg.", RED)
        return

    ffmpeg_dir_existed = os.path.isdir(FFMPEG_DIR)

    if not ffmpeg_dir_existed:
        log("WARN", f"FFmpeg source folder not found at: {FFMPEG_DIR}", YELLOW)

        if shutil.which("git") is None:
            log("ERROR", "git not found in PATH. Cannot clone FFmpeg. Skipping.", RED)
            return

        if not ask_yes_no("Clone FFmpeg from GitHub into that folder now?"):
            log("INFO", "Skipping FFmpeg/ffprobe deployment.", YELLOW)
            return

        log("RUN", f"Cloning {FFMPEG_REPO}...", BLUE)

        try:
            run(["git", "clone", FFMPEG_REPO, FFMPEG_DIR])
        except subprocess.CalledProcessError:
            log("ERROR", "Failed to clone FFmpeg. Skipping.", RED)
            return

    if not build_ffmpeg(arch_info):
        return

    ffmpeg_bin = os.path.join(FFMPEG_DIR, "ffmpeg")
    ffprobe_bin = os.path.join(FFMPEG_DIR, "ffprobe")

    if not os.path.exists(ffmpeg_bin) or not os.path.exists(ffprobe_bin):
        log("ERROR", "ffmpeg/ffprobe binaries not found after build. Skipping push.", RED)
        return

    log("PUSH", "Uploading ffmpeg...", GREEN)
    adb(["push", ffmpeg_bin, f"{TARGET_DIR}/ffmpeg"])

    log("PUSH", "Uploading ffprobe...", GREEN)
    adb(["push", ffprobe_bin, f"{TARGET_DIR}/ffprobe"])

    adb(["shell", "chmod", "755", f"{TARGET_DIR}/ffmpeg", f"{TARGET_DIR}/ffprobe"])

    log("DONE", "FFmpeg/ffprobe deployed successfully.", GREEN)


print()
log("INFO", "FFmpeg/ffprobe deployment is optional and separate from the server deploy above.", YELLOW)
log("INFO", "You should (re)run it whenever you switch to a device with a different CPU "
            "architecture, or if it has never been done on the current device before.", YELLOW)
log("INFO", "If skipped, native thumbnail generation won't be available on the device: the "
            "app will have to rely on an external API for thumbnails, or won't generate "
            "thumbnails for uploaded media at all.", YELLOW)

if ask_yes_no("Build and deploy FFmpeg/ffprobe now?"):
    deploy_ffmpeg()
else:
    log("INFO", "Skipping FFmpeg/ffprobe deployment.", YELLOW)