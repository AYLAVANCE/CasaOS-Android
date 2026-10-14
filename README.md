# CasaOS Android (Termux Port)

[![Platform: Android](https://img.shields.io/badge/Platform-Android_10+-3DDC84?style=flat-square&logo=android&logoColor=white)](https://android.com)
[![Architecture: ARM64](https://img.shields.io/badge/Architecture-ARM64%20%2F%20aarch64-0091BD?style=flat-square&logo=arm&logoColor=white)](https://arm.com)
[![Environment: Termux](https://img.shields.io/badge/Environment-Termux-000000?style=flat-square&logo=gnu-bash&logoColor=white)](https://f-droid.org/packages/com.termux/)
[![Build: Native Go PIE](https://img.shields.io/badge/Build-Native_Go_PIE-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=flat-square)](LICENSE)

A native, zero-root port of **CasaOS** engineered specifically for **ARM64 Android devices** using **Termux** as the host environment abstraction layer.

This port completely strips out Linux distribution dependencies (`systemd`, `lsblk`, `udevil`, `mergerfs`, glibc FHS conventions) in favor of direct Android kernel sysfs inspection, Bionic libc path mapping, user-space process supervision, and rootless container orchestration.

---

## ⚡ Quick Start: One-Line Installation

Open **Termux** on your Android device and run:

```bash
curl -fsSL https://raw.githubusercontent.com/AYLAVANCE/CasaOS-Android/main/install/install-android.sh | bash
```

Once the installer finishes, open your browser and navigate to:
```text
http://<your-android-ip>:8080
```

---

## 📱 System Requirements

| Component | Requirement | Notes |
|:---|:---|:---|
| **OS Version** | Android 10.0+ (API 29+) | Supports Android 10, 11, 12, 13, 14+ |
| **CPU Architecture** | 64-bit ARM (`aarch64` / `arm64-v8a`) | Required for native binary execution |
| **RAM** | 2 GB minimum (4 GB+ recommended) | Depends on container workloads |
| **Termux** | Latest build from **F-Droid** or GitHub | ⚠️ Do **not** use the deprecated Google Play Store build |
| **Termux:API** | Termux:API App + Package | Required for battery, WiFi, and hardware sensor telemetry |
| **Storage Permissions** | Storage Access Framework (SAF) | Run `termux-setup-storage` prior to installation |
| **Root Access** | ❌ **Not Required** | Runs 100% in user-space |

---

## 🏗️ Core Architecture & Native Android Features

```
+--------------------------------------------------------------------------+
|                       Android Hardware (ARM64 SoC)                       |
+--------------------------------------------------------------------------+
|      Android Kernel (Linux 4.19 / 5.x / 6.x) + Bionic Libc Runtime       |
+--------------------------------------------------------------------------+
|                  Termux User-Space Abstraction Layer                     |
|                                                                          |
|  +--------------------+  +--------------------+  +--------------------+  |
|  | casaos-supervisor  |  |  CasaOS Daemon     |  | Hardware Telemetry |  |
|  | (Process Manager)  |  |  (Core REST API)   |  | (sysfs / procfs)   |  |
|  +--------------------+  +--------------------+  +--------------------+  |
|  | Doki Container Engine | Web Terminal (PTY) | Storage Service (SAF) |  |
|  +--------------------+  +--------------------+  +--------------------+  |
+--------------------------------------------------------------------------+
|                      Vue.js Responsive Web Dashboard                     |
|            (Accessible from any LAN browser at http://ip:8080)            |
+--------------------------------------------------------------------------+
```

### 1. Platform Abstraction Layer
- All hardcoded FHS paths (`/etc`, `/var/lib`, `/var/run`, `/DATA`) are mapped dynamically to Termux prefixes (`$PREFIX/etc/casaos`, `$HOME/.casaos/`, `$HOME/storage/shared/CasaOS`).
- Go build tags (`//go:build android`) decouple the Android implementation while maintaining build compatibility.

### 2. Standalone Process Supervisor (`casaos-supervisor`)
- Replaces Linux `systemd` with a native Go process supervisor.
- Manages startup ordering, exponential backoff restarts, PID health tracking, and graceful shutdown.
- Automatically handles Android `termux-wake-lock` to keep the dashboard and background services alive when the screen is off or the device enters Doze mode.

### 3. Native Hardware Telemetry
- Reads CPU utilization, frequency scaling, and big.LITTLE core grouping directly from `/proc/stat` and `/sys/devices/system/cpu/`.
- Parses thermal zones from `/sys/devices/virtual/thermal/`.
- Interacts with `termux-api` for real-time battery health, charging status, temperature, and active WiFi SSID/link speeds.

### 4. Rootless Container Runtime
- Implements a unified container runtime interface supporting the rootless **Doki** engine natively in Termux user-space.
- Includes automated port translation (offsets privileged ports `<1024` by `+8000`).
- Provides an optional lightweight QEMU Alpine VM fallback for complex containers requiring full kernel namespaces.

### 5. Scoped Storage & Disk Management
- Discards desktop Linux utilities (`lsblk`, `smartctl`, `mergerfs`, `fdisk`) in favor of direct POSIX `syscall.Statfs`.
- Automatically maps internal storage (`~/storage/shared/`), external MicroSD cards (`~/storage/external-1/`), and USB OTG storage.
- Bridges Android Storage Access Framework (SAF) folder requests through `termux-saf-*`.

### 6. Built-in Web Terminal
- Sandboxed web terminal embedded into the dashboard.
- Uses `creack/pty` to allocate a direct pseudoterminal connected to Termux `/bin/bash` over WebSockets.

### 7. Mobile-Optimized Dashboard UI
- Vue.js dashboard adapted for smartphone and tablet viewports.
- Bottom navigation bar, touch-friendly targets (44px+), glassmorphism cards, and dedicated Android system widgets.
- PWA manifest included for installing CasaOS Android directly to your home screen.

---

## ⚙️ Post-Installation Setup

### 1. Disable Android Phantom Process Killer (Android 12+)
Android 12 and newer may terminate long-running background processes spawned in Termux. To prevent this, run the following ADB command once from your computer:

```bash
adb shell "/system/bin/device_config put activity_manager max_phantom_processes 2147483647"
```

### 2. Service Management Commands
The supervisor can be controlled from Termux at any time:

```bash
# Start all CasaOS services in the background
casaos-supervisor start

# Check service status
casaos-supervisor status

# Stop all CasaOS services
casaos-supervisor stop

# Restart all services
casaos-supervisor restart
```

---

## 🛠️ Building From Source

If you want to compile the binaries yourself within Termux or cross-compile from a PC:

### On Termux (Direct Build)
```bash
pkg install golang git make -y
git clone https://github.com/AYLAVANCE/CasaOS-Android.git
cd CasaOS-Android

# Build Core and Supervisor
go build -tags android -buildmode=pie -ldflags="-s -w" -o $PREFIX/bin/casaos main.go
go build -tags android -buildmode=pie -ldflags="-s -w" -o $PREFIX/bin/casaos-supervisor cmd/supervisor/main.go
```

### From Linux / Windows / macOS (Cross-Compilation for ARM64)
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags android -buildmode=pie -ldflags="-s -w" -o casaos main.go
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags android -buildmode=pie -ldflags="-s -w" -o casaos-supervisor cmd/supervisor/main.go
```

---

## 🗑️ Uninstallation

To remove CasaOS Android and its configurations from Termux:

```bash
casaos-supervisor stop
rm -rf $PREFIX/bin/casaos $PREFIX/bin/casaos-supervisor $PREFIX/etc/casaos ~/.casaos
termux-wake-unlock
```

---

## 📄 License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
All original CasaOS concepts courtesy of IceWhale Technology, refactored and ported for native Android user-space operations.
