# Wireless Drive

Wireless Drive is a self-hosted media server for storing and streaming media over your local network. It ships with an embedded static frontend for uploading, browsing, and editing files, so it can run entirely on its own as a "wireless drive."

It's primarily designed to work alongside the **Obsidian Plugin**, allowing you to open files across vaults without storing them on your main device. With that in mind, this README includes a full guide for running the API as a background process on an old phone, or even on another PC, turning it into a dedicated Wireless Drive that serves the plugin and other applications.

> **Note**
>
> ***No root is required*** to run the server this way. Keep in mind that if the phone shuts down or restarts, you'll need to run the `nohup` command again via ADB. It's recommended to disable "automatic restart" on the target phone — with it off, the server should never go down. At most, the process may be paused after a long period of inactivity with the screen off due to battery optimization, but it will typically resume as soon as you try to access it a few times or unlock the screen.
>
> The server itself barely uses any battery, so it should last at least a week on a single charge — in my own testing, that's been the case with a Samsung Galaxy J4. It also uses very little RAM, so even old, low-spec phones will run it comfortably. The same holds true for thumbnail generation with FFmpeg, which handles heavy videos and images without issue.

> Obsidian Plugin: https://github.com/AlexisCarvalho/obsidian-plugin-wireless-drive

There's also a companion **Android App** that uses this same API to browse your stored media more efficiently on smartphones.

> Android App: https://github.com/AlexisCarvalho/android-wireless-drive

---

## Table of Contents

- [Accessing the Server from Other Devices](#accessing-the-server-from-other-devices)
- [Running the API on PC](#running-the-api-on-pc)
- [Running the API on Android (via ADB)](#running-the-api-on-android-via-adb)
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
| `BASE_PATH` | ✅ | Root directory where Wireless Drive stores all uploaded files and generated thumbnails. |
| `JWT_SECRET` | ✅ | Secret used to sign user authentication tokens. |
| `STREAM_SECRET` | ✅ | Secret used to sign temporary media streaming URLs. |
| `DB_NAME` | ❌ | SQLite database filename. Defaults to `wireless_drive.db`. |
| `UPLOADS_DIR` | ❌ | Upload directory relative to `BASE_PATH`. Defaults to `uploads`. |
| `THUMBS_DIR` | ❌ | Thumbnail directory relative to `BASE_PATH`. Defaults to `thumbs`. |
| `FFMPEG_PATH` | ❌ | Path to the `ffmpeg` executable. |
| `FFPROBE_PATH` | ❌ | Path to the `ffprobe` executable. |
| `PORT` | ❌ | HTTP server port. Defaults to `8085`. |
| `THUMBNAIL_API_URL` | ❌ | External Thumbnail API used when FFmpeg is unavailable on the server. |

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

### Checking the Target Phone's Architecture

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

### Building FFmpeg for Android

The server uses FFmpeg and FFprobe to generate image and video thumbnails directly on Android. This is recommended, but the server works fine without it.

- If you only plan to use the server with the Obsidian plugin (which doesn't use thumbnails at all), you can skip straight to [Building the Go Server](#building-the-go-server) not configuring ffmpeg locally or setting a Thumbnail API will simply result on thumbnails not being generated, but you can generate them later on if you need.
- If you want thumbnails but don't want to set up FFmpeg on the Android device the server is currently running, you can instead run the standalone **Thumbnail API**, which contains just the thumbnail logic, on your PC or another Smartphone. The application will automatically fall back to calling this API before leaving a thumbnail empty. When uploading media or triggering "generate missing thumbnails," the request will be forwarded to this API over the local network.

  This approach isn't recommended, since it will transfer every media file to the device running the Thumbnail API individually before generating. It's still fast, as it uses the full bandwidth of your device's Wi-Fi connection rather than the internet, but it's slower than generating thumbnails directly on the phone, where the media already resides. But may be usefull if you want to store medias on one device and use other to generate the thumbnails. If you want to do that no changes on code need to be made, only don't push the binary of ffmpeg and ffprobe to the server folder, it will try to detect it, fail and after that will switch to the Thumbnail API set on the THUMBNAIL_API_URL automatically. If that fails too it will leave the thumbnail empty for you to try to generate manually later.

#### Requirements

- Android NDK r29
- Linux
- Git
- Make

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

#### 1. Get the NDK

Download it here: https://developer.android.com/ndk/downloads

#### 2. Clone FFmpeg

```bash
git clone https://github.com/FFmpeg/FFmpeg.git
cd FFmpeg
```

#### 3. Configure the Android NDK toolchain

```bash
export ANDROID_NDK=$HOME/android-ndk-r29

export TOOLCHAIN=$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64

export CC=$TOOLCHAIN/bin/aarch64-linux-android21-clang
export CXX=$TOOLCHAIN/bin/aarch64-linux-android21-clang++
export AR=$TOOLCHAIN/bin/llvm-ar
export LD=$TOOLCHAIN/bin/ld.lld
export STRIP=$TOOLCHAIN/bin/llvm-strip
```

#### 4. Configure FFmpeg

The configuration below builds a minimal FFmpeg containing only the components Wireless Drive actually needs.

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

#### 5. Build FFmpeg

```bash
make -j$(nproc)
```

### Building the Go Server

Compile the server for the same Android target used above.

```bash
export ANDROID_NDK=$HOME/android-ndk-r29

export GOOS=android
export GOARCH=arm64
export CGO_ENABLED=1

export CC="$ANDROID_NDK/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android21-clang"

go build -v ./cmd/server
```

### Installing on the Android Device

> **Note**
>
> This tutorial uses `/data/local/tmp/` as the install location for the binaries because, on most Android devices, it can be executed without root access. The database and media files, however, are stored separately under `/storage/emulated/0/databases/wirelessDrive`, since some phones wipe the contents of `/data/local/tmp/` on reboot. This also lets you browse your media directly through the phone's file system if you ever need to. The directories under `/storage/emulated/0/databases/wirelessDrive` including itself are created automatically, so there's no need to set them up by hand.

**Create the destination directory:**

```bash
adb shell mkdir -p /data/local/tmp/wirelessDrive
```

**Copy the binaries:**

```bash
adb push ffmpeg /data/local/tmp/wirelessDrive/
adb push ffprobe /data/local/tmp/wirelessDrive/
adb push server /data/local/tmp/wirelessDrive/
adb push .env /data/local/tmp/wirelessDrive/
```

**Grant execution permission:**

```bash
adb shell chmod +x /data/local/tmp/wirelessDrive/ffmpeg
adb shell chmod +x /data/local/tmp/wirelessDrive/ffprobe
adb shell chmod +x /data/local/tmp/wirelessDrive/server
```

**Verify the installation:**

```bash
adb shell /data/local/tmp/wirelessDrive/ffmpeg -version
adb shell /data/local/tmp/wirelessDrive/ffprobe -version
```

If both commands print their version information, FFmpeg and FFprobe are correctly installed and ready to be used by the server.

### Running the Server

```bash
cd /data/local/tmp/wirelessDrive
```

> **Important**
>
> The `.env` file must be located in the same directory as the `server` executable. The server loads it from its current working directory, which is why this `cd` command is required before starting it.

```bash
nohup ./server >/dev/null 2>&1 &
disown
```