# Wireless Drive

Wireless Drive is a self-hosted media server for storing and streaming media over your local network. It ships with an embedded static website for uploading, browsing, and editing files, so it can run entirely on its own as a "wireless drive."

It's primarily designed to work alongside the **Obsidian Plugin**, allowing you to open files across vaults without storing them on your main device. With that in mind, this README includes a full guide for running the API as a background process on an old phone, or even on another PC, turning it into a dedicated Wireless Drive that serves the plugin and other applications.

> **Note for: `background process on an old phone`**
>
> ***No root is required.*** To run the server this way keep in mind that if the phone shuts down or restarts, you’ll need to run the nohup command again via ADB—or re-run auto_deploy, which will recompile and redeploy everything automatically (with no data loss). It’s recommended to disable “Automatic restart” on the target phone; with it off, the server should stay up indefinitely. At most, the process may be paused after a long period of inactivity with the screen off due to battery optimization, but it will typically resume as soon as you interact with the device a few times or unlock the screen, which brings Android’s standby processes back to life.
>
> The server itself barely uses any battery, so it should last at least a week on a single charge — in my own testing, that's been the case with a Samsung Galaxy J4. It also uses very little RAM, so even old, low-spec phones will run it comfortably. The same holds true for thumbnail generation with FFmpeg, which handles heavy videos and images without issue.

### Check Related Repositories
> Obsidian Plugin: https://github.com/AlexisCarvalho/obsidian-plugin-wireless-drive

As mentioned above, this is used to open media from this API across all vaults; you only need to call the Markdown Code Block Processor.
````python
```wg # Loads the media with the specified ID
ID   
``` 

```wg # Opens a table listing all stored media for selection
```

```wg # Opens a pre-filtered table
search TEXT
```
````

> Android App: https://github.com/AlexisCarvalho/android-wireless-drive

There's also a companion **Android App** that uses this same API to browse your stored media more efficiently on smartphones.

> Thumbnail API: https://github.com/AlexisCarvalho/golang-wireless-drive-fallback-thumbnail-api

Here, the Thumbnail API is only used for very specific purposes: as a fallback when FFmpeg is not available on the system running the server.

### `auto_deploy.py`

This repository ships with `auto_deploy.py`, a script that automates the entire Android deployment flow described in [Running the API on Android](#running-the-api-on-android-via-adb): it detects the connected device's CPU architecture, cross-compiles the Go server, optionally cross-compiles FFmpeg/FFprobe (cloning them straight from the official [FFmpeg repository](https://github.com/FFmpeg/FFmpeg.git) if you don't already have the source), pushes everything to the device via `adb`, and starts the server for you.

It's meant to replace the manual, step-by-step process for everyday use. The manual guide further down is still here in full for anyone who wants full control, wants to understand what the script is doing under the hood, or needs a fallback if the script fails on their setup — see [Automated Deployment](#automated-deployment-auto_deploypy) for details and requirements before using it.

---

## Table of Contents

- [Accessing the Server from Other Devices](#accessing-the-server-from-other-devices)
- [Configuration](#configuration)
- [Running the API on PC](#running-the-api-on-pc)
- [Running the API on Android (via ADB)](#running-the-api-on-android-via-adb)
  - [Automated Deployment (auto_deploy.py)](#automated-deployment-auto_deploypy)
  - [Manual Deployment](#manual-deployment)
    - [Requirements](#requirements)
    - [Checking the Target Phone's Architecture](#checking-the-target-phones-architecture)
    - [Building FFmpeg for Android](#building-ffmpeg-for-android)
    - [Building the Go Server](#building-the-go-server)
    - [Installing on the Android Device](#installing-on-the-android-device)
    - [Running the Server](#running-the-server)

---

## Accessing the Server from Other Devices

To access the server from other devices on the same local network (i.e., not through `localhost`), you'll need the local IP address of the device running it.

In most cases, you don't need to set up a static IP unless you have some issues with your router DHCP. Which leases are typically stable and won't change frequently as long as the device stays connected, or is only turned off for a few hours at a time. If you run into issues with the IP changing unexpectedly, consider enabling a static IP (or a DHCP reservation on your router) for the device.

---

## Configuration

Wireless Drive reads its configuration from a `.env` file located in the same directory as the server executable.

> **Note**
>
> This .env file must always be created manually (by renaming .env.example after reviewing it) — neither the manual steps nor auto_deploy.py generate or populate it for you. auto_deploy.py will only push the file to the device, and only if it already exists next to the script when you run it. Without a valid .env in place beforehand, the server may fail to start or fall back to unintended defaults. The .env.example file already contains everything you need, but you may want to adjust THUMBNAIL_API_URL if you plan to generate thumbnails on another device instead of locally. As for JWT_SECRET and STREAM_SECRET, you can also change them, but this is optional as long as you're on your local network.

Example:

```dotenv
# =====================================
# Database
# =====================================
DB_NAME=wireless_drive.db

# =====================================
# Security
# =====================================
JWT_SECRET=your-secret
STREAM_SECRET=your-stream-secret

# =====================================
# Storage
# =====================================
BASE_PATH=/storage/emulated/0/databases/wirelessDrive
UPLOADS_DIR=uploads
THUMBS_DIR=thumbs

# =====================================
# FFmpeg
# =====================================
FFMPEG_PATH=/data/local/tmp/wirelessDrive/ffmpeg
FFPROBE_PATH=/data/local/tmp/wirelessDrive/ffprobe

# =====================================
# Server
# =====================================
PORT=8085

# =====================================
# Thumbnail API
# =====================================
THUMBNAIL_API_URL=http://192.168.0.162:8086
```

### Configuration Reference

| Variable | Required | Description |
|----------|:--------:|-------------|
| `JWT_SECRET` | ✅ | Secret used to sign user authentication tokens. |
| `STREAM_SECRET` | ✅ | 	Secret used to sign temporary media streaming URLs. |
| `BASE_PATH` | ✅ | Root directory where Wireless Drive stores all uploaded files and generated thumbnails. |
| `UPLOADS_DIR` | ❌ | Upload directory relative to `BASE_PATH`. Defaults to `uploads`. |
| `THUMBS_DIR` | ❌ | Thumbnail directory relative to `BASE_PATH`. Defaults to `thumbs`. |
| `FFMPEG_PATH` | ❌ | Path to the ffmpeg executable. If not set, thumbnails will not be generated locally. |
| `FFPROBE_PATH` | ❌ | Path to the ffprobe executable. If not set, thumbnails will not be generated locally. |
| `THUMBNAIL_API_URL` | ❌ | External Thumbnail API URL used when FFmpeg is unavailable or not configured on the server. If none of these are set (`FFMPEG_PATH` and `FFPROBE_PATH`) for local, or (`THUMBNAIL_API_URL`) for external, the server will still work, but no thumbnails will be generated. |
| `DB_NAME` | ❌ | SQLite database filename. Defaults to `wireless_drive.db`. |
| `PORT` | ❌ | HTTP server port. Defaults to `8085`. |

---

## Running the API on PC

### Build

```bash
go build ./cmd/server
```

### Run

```bash
./server
```

---

## Running the API on Android (via ADB)

Running the server on an old Android phone lets you repurpose it as dedicated local storage that stays connected to your network at all times.

You can either let [`auto_deploy.py`](#automated-deployment-auto_deploypy) handle the whole process, or follow the [manual steps](#manual-deployment) yourself — both end up doing the same thing, so pick whichever fits your workflow. The manual steps are entirely optional if the script works for you.

### Automated Deployment (`auto_deploy.py`)

`auto_deploy.py` lives at the root of this repository, next to `go.mod`. Running it (`python3 auto_deploy.py`) does the following, in order:

1. Checks that `go` and `adb` are available in `PATH`.
2. Lists devices connected via `adb devices` and asks you to pick one if more than one is connected.
3. Reads the device's `ro.product.cpu.abi` and automatically maps it to the correct `GOARCH` and NDK Clang target (see the table below) — no manual editing needed.
4. Cross-compiles the Go server with `go build ./cmd/server`.
5. Kills any previously running `WirelessDrive` process on the device.
6. Creates `/data/local/tmp/wirelessDrive` on the device if it doesn't exist.
7. Pushes the compiled binary as `WirelessDrive`, and pushes your local `.env` file too — **only if it already exists** next to the script (see the note in [Configuration](#configuration)).
8. Starts the server on the device in the background via `nohup`.
9. Asks whether you also want to build and push **FFmpeg/FFprobe** (fully optional, see below).

If you opt into step 9, the script:

- Looks for a `FFmpeg/` folder next to `auto_deploy.py`. If it's missing, it warns you and offers to clone it directly from https://github.com/FFmpeg/FFmpeg.git — if you decline, or don't have `git` installed, it just skips this part and leaves the rest of the deployment untouched.
- Runs `make distclean` inside `FFmpeg/` before configuring, since the target architecture may differ from whatever it was last built for (e.g. after switching devices). This is expected to error out harmlessly on a fresh clone that's never been built.
- Runs the same minimal `./configure` and `make -j$(nproc)` described in [Building FFmpeg for Android](#building-ffmpeg-for-android), using the architecture flags it already detected in step 3.
- Pushes the resulting `ffmpeg`/`ffprobe` binaries to the device and marks them executable.

#### What you still need to set up yourself

| Item | Required for `auto_deploy.py`? | Notes |
|---|:---:|---|
| Android NDK r29 | ✅ Always | Set the `ANDROID_NDK` environment variable, or place it at `~/android-ndk-r29` (the script's default). Needed even if you skip the FFmpeg step, since the Go server itself is cross-compiled with CGO. |
| Go | ✅ Always | Must be in `PATH`. |
| `adb` / platform-tools | ✅ Always | Must be in `PATH`, with the phone connected and authorized (USB debugging enabled — see [Requirements](#requirements)). |
| `.env` file | ✅ Recommended | Not created by the script — you must write it yourself beforehand (see [Configuration](#configuration)). Without it, the server can still start, but without your storage path, secrets, or FFmpeg/thumbnail settings configured. |
| FFmpeg/FFprobe build | ❌ Optional | Purely opt-in prompt. Skipping it just means no native thumbnail generation on-device — see [Building FFmpeg for Android](#building-ffmpeg-for-android) for what that means and your fallback options. |
| `git` | ❌ Optional | Only needed if you want the script to auto-clone FFmpeg for you when the `FFmpeg/` folder doesn't already exist. |
| `make` | ❌ Optional | Only needed if you opt into building FFmpeg/FFprobe. |

Re-run the script whenever you switch to a different phone/architecture, or the very first time you set one up — it rebuilds both the server and (if you opt in) FFmpeg/FFprobe for whichever device is currently connected.

### Manual Deployment

> Everything in this section is **optional** if `auto_deploy.py` already works for you — it exists as a fallback for when the script fails on your setup, and as a reference for anyone who wants to understand or customize the individual steps.

#### Requirements

- **Android NDK r29** — download it from https://developer.android.com/ndk/downloads. Needed to cross-compile both the Go server and FFmpeg.
- **Go**, installed and in `PATH`.
- **ADB (platform-tools)**, installed and in `PATH`. On the phone, enable Developer Options and USB Debugging, connect it (USB or Wi-Fi debugging), and authorize the connection when prompted. Confirm the phone shows up with:

  ```bash
  adb devices
  ```

- **Git** and **Make** — only required if you're also building FFmpeg/FFprobe from source (skip these if you're not setting up thumbnails, see [Building FFmpeg for Android](#building-ffmpeg-for-android)).

> **Note**
>
> When running the scripts below, replace the NDK path with wherever the Android NDK is installed on your machine.
>
> Using `$HOME` is recommended over an absolute path:
>
> ```bash
> export ANDROID_NDK=$HOME/android-ndk-r29
> ```
>
> This guide targets **arm64-v8a**. If you're building for a different architecture, update the compiler and architecture flags accordingly (see [Checking the Target Phone's Architecture](#checking-the-target-phones-architecture)).

#### Checking the Target Phone's Architecture

Both FFmpeg and the Go server need to be compiled for the exact CPU architecture of the phone you're targeting. This guide uses **arm64-v8a** as an example, since it's the most common architecture on Android phones released in the last several years, but you should confirm it before building anything.

With the phone connected via ADB, run:

```bash
adb shell getprop ro.product.cpu.abi
```

This prints the phone's primary ABI (e.g. `arm64-v8a`). Use the table below to map it to the correct build flags:

| `ro.product.cpu.abi` | Go `GOARCH` | NDK Clang target (API 21) |
|---|---|---|
| `arm64-v8a` | `arm64` | `aarch64-linux-android21-clang` |
| `armeabi-v7a` | `arm` | `armv7a-linux-androideabi21-clang` |
| `x86_64` | `amd64` | `x86_64-linux-android21-clang` |
| `x86` | `386` | `i686-linux-android21-clang` |

If your phone reports anything other than `arm64-v8a`, substitute the corresponding `GOARCH` value and Clang target in the [Building FFmpeg for Android](#building-ffmpeg-for-android) and [Building the Go Server](#building-the-go-server) steps below — everywhere `aarch64`, `arm64`, or `aarch64-linux-android21-clang` appears.

#### Building FFmpeg for Android

The server uses FFmpeg and FFprobe to generate image and video thumbnails directly on Android. This whole section — and FFmpeg itself — is **optional**; the server works fine without it.

- If you only plan to use the server with the Obsidian plugin (which doesn't use thumbnails at all), you can skip straight to [Building the Go Server](#building-the-go-server). Not configuring FFmpeg locally or setting a Thumbnail API will simply result in thumbnails not being generated, but you can generate them later on if you need.
- If you want thumbnails but don't want to set up FFmpeg on the Android device the server is currently running on, you can instead run the standalone **Thumbnail API**, which contains just the thumbnail logic, on your PC or another smartphone. The application will automatically fall back to calling this API before leaving a thumbnail empty. When uploading media or triggering "generate missing thumbnails," the request will be forwarded to this API over the local network.

  This approach isn't recommended, since it will transfer every media file to the device running the Thumbnail API individually before generating. It's still fast, as it uses the full bandwidth of your device's Wi-Fi connection rather than the internet, but it's slower than generating thumbnails directly on the phone, where the media already resides. But it may be useful if you want to store media on one device and use another to generate the thumbnails. If you want to do that, no changes to the code need to be made — just don't push the `ffmpeg`/`ffprobe` binaries to the server folder. The server will try to detect them, fail, and automatically switch to the Thumbnail API set on `THUMBNAIL_API_URL`. If that fails too, it will leave the thumbnail empty for you to try to generate manually later.

Source: this section cross-compiles FFmpeg straight from the official repository — https://github.com/FFmpeg/FFmpeg.git.

##### 1. Get the NDK

Already covered in [Requirements](#requirements) above — make sure `$ANDROID_NDK` is set before continuing.

##### 2. Clone FFmpeg

```bash
git clone https://github.com/FFmpeg/FFmpeg.git
cd FFmpeg
```

##### 3. Configure the Android NDK toolchain

```bash
export ANDROID_NDK=$HOME/android-ndk-r29

export TOOLCHAIN=$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64

export CC=$TOOLCHAIN/bin/aarch64-linux-android21-clang
export CXX=$TOOLCHAIN/bin/aarch64-linux-android21-clang++
export AR=$TOOLCHAIN/bin/llvm-ar
export LD=$TOOLCHAIN/bin/ld.lld
export STRIP=$TOOLCHAIN/bin/llvm-strip
```

##### 4. Configure FFmpeg

The configuration below builds a minimal FFmpeg containing only the components Wireless Drive actually needs.

> If you're rebuilding for a different architecture than last time (e.g. switched phones), run `make distclean` before re-configuring — the previous build's object files won't match the new target and will cause the build to fail or produce a binary that doesn't run on the new device.

```bash
./configure \
    --target-os=android \
    --arch=aarch64 \
    --cpu=armv8-a \
    --cc="$CC" \
    --cxx="$CXX" \
    --ar="$AR" \
    --strip="$STRIP" \
    --enable-cross-compile \
    \
    --disable-shared \
    --enable-static \
    \
    --disable-everything \
    \
    --enable-ffmpeg \
    --enable-ffprobe \
    --disable-ffplay \
    \
    --enable-avcodec \
    --enable-avformat \
    --enable-avutil \
    --enable-swscale \
    \
    --enable-protocol=file \
    \
    --enable-demuxer=mov \
    --enable-demuxer=matroska \
    --enable-demuxer=avi \
    --enable-demuxer=image2 \
    --enable-demuxer=webm_dash_manifest \
    \
    --enable-muxer=image2 \
    \
    --enable-parser=h264 \
    --enable-parser=hevc \
    --enable-parser=vp8 \
    --enable-parser=vp9 \
    --enable-parser=av1 \
    \
    --enable-decoder=h264 \
    --enable-decoder=hevc \
    --enable-decoder=vp8 \
    --enable-decoder=vp9 \
    --enable-decoder=av1 \
    \
    --enable-decoder=mjpeg \
    --enable-decoder=png \
    --enable-decoder=webp \
    --enable-decoder=gif \
    \
    --enable-encoder=mjpeg \
    \
    --enable-filter=scale \
    --enable-filter=format
```

##### 5. Build FFmpeg

```bash
make -j$(nproc)
```

This produces `ffmpeg` and `ffprobe` binaries in the root of the `FFmpeg/` folder, which you'll push to the device in [Installing on the Android Device](#installing-on-the-android-device).

#### Building the Go Server

Compile the server for the same Android target used above.

```bash
export ANDROID_NDK=$HOME/android-ndk-r29

export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

export CC="$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang"

go build -v ./cmd/server
```

#### Installing on the Android Device

> **Note**
>
> This tutorial uses `/data/local/tmp/` as the install location for the binaries because, on most Android devices, it can be executed without root access. The database and media files, however, are stored separately under `/storage/emulated/0/databases/wirelessDrive`, since some phones wipe the contents of `/data/local/tmp/` on reboot. This also lets you browse your media directly through the phone's file system if you ever need to. The directories under `/storage/emulated/0/databases/wirelessDrive` including itself are created automatically, so there's no need to set them up by hand.

**Create the destination directory:**

```bash
adb shell mkdir -p /data/local/tmp/wirelessDrive
```

**Copy the binaries:**

`ffmpeg` and `ffprobe` are optional here — only push them if you built them in [Building FFmpeg for Android](#building-ffmpeg-for-android). The server and `.env` are the only required pieces.

```bash
# Optional — only if you built FFmpeg (see above)
adb push ffmpeg /data/local/tmp/wirelessDrive/
adb push ffprobe /data/local/tmp/wirelessDrive/

# Required — create this yourself first, see the Configuration section
adb push .env /data/local/tmp/wirelessDrive/

# Required
# The WirelessDrive in PascalCase at the end will auto-rename the server file to "WirelessDrive"
# This way it will be easier to find on the process list
adb push server /data/local/tmp/wirelessDrive/WirelessDrive
```

**Grant execution permission:**

```bash
# Optional — only if you pushed FFmpeg/FFprobe above
adb shell chmod +x /data/local/tmp/wirelessDrive/ffmpeg
adb shell chmod +x /data/local/tmp/wirelessDrive/ffprobe

# Required
adb shell chmod +x /data/local/tmp/wirelessDrive/WirelessDrive
```

**Verify the installation:**

```bash
# Only applicable if you installed FFmpeg/FFprobe
adb shell /data/local/tmp/wirelessDrive/ffmpeg -version
adb shell /data/local/tmp/wirelessDrive/ffprobe -version
```

If both commands print their version information, FFmpeg and FFprobe are correctly installed and ready to be used by the server. If you skipped FFmpeg, there's nothing to verify here — just move on.

#### Running the Server

> **Just a Reminder**
>
> The `.env` file must be located in the same directory as the `WirelessDrive` executable.

```bash
adb shell nohup /data/local/tmp/wirelessDrive/WirelessDrive >/dev/null 2>&1 &
disown
```