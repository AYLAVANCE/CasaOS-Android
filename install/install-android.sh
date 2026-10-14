#!/data/data/com.termux/files/usr/bin/bash

# CasaOS for Android/Termux Installation Script

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=======================================${NC}"
echo -e "${GREEN}   CasaOS Android/Termux Installer     ${NC}"
echo -e "${BLUE}=======================================${NC}"

if [ -z "$TERMUX_VERSION" ]; then
    echo -e "${RED}Error: This script must be run inside Termux.${NC}"
    exit 1
fi

ARCH=$(uname -m)
if [ "$ARCH" != "aarch64" ]; then
    echo -e "${RED}Error: Architecture $ARCH is not supported. Only aarch64 is supported.${NC}"
    exit 1
fi

echo -e "${YELLOW}Updating packages...${NC}"
pkg update -y && pkg upgrade -y

echo -e "${YELLOW}Installing dependencies...${NC}"
pkg install -y termux-api termux-tools openssl git golang

echo -e "${YELLOW}Requesting Android storage permission...${NC}"
termux-setup-storage
sleep 2

echo -e "${YELLOW}Setting up directory structure...${NC}"
mkdir -p ~/.casaos/{data,run,log,db,www,appdata,containers,tmp,conf}
mkdir -p $PREFIX/etc/casaos

echo -e "${YELLOW}Downloading CasaOS Android source...${NC}"
if [ -d "CasaOS-Android" ]; then
    echo -e "${YELLOW}CasaOS-Android directory already exists, updating repo...${NC}"
    cd CasaOS-Android && git pull || true
else
    git clone https://github.com/AYLAVANCE/CasaOS-Android.git
    cd CasaOS-Android
fi

echo -e "${YELLOW}Building CasaOS...${NC}"
go build -tags android -buildmode=pie -o $PREFIX/bin/casaos main.go

echo -e "${YELLOW}Building CasaOS Supervisor...${NC}"
if [ -f "cmd/supervisor/main.go" ]; then
    go build -tags android -buildmode=pie -o $PREFIX/bin/casaos-supervisor cmd/supervisor/main.go
else
    echo -e "${YELLOW}cmd/supervisor/main.go not found, skipping supervisor build.${NC}"
fi

echo -e "${YELLOW}Acquiring wake lock...${NC}"
termux-wake-lock

echo -e "${YELLOW}Disabling Phantom Process Killer (Instructions)${NC}"
echo -e "${BLUE}To prevent Termux from being killed by Android, run this via ADB from a PC:${NC}"
echo -e "adb shell \"/system/bin/device_config put activity_manager max_phantom_processes 2147483647\""

echo -e "${GREEN}Installation Complete!${NC}"
IP=$(ifconfig wlan0 | grep 'inet ' | awk '{print $2}')
if [ -z "$IP" ]; then
    IP="localhost"
fi

echo -e "${BLUE}=======================================${NC}"
echo -e "${GREEN} CasaOS is ready to run!               ${NC}"
echo -e "${BLUE}=======================================${NC}"
echo -e "Run ${YELLOW}casaos${NC} to start."
echo -e "Access CasaOS at: ${GREEN}http://${IP}:8080${NC}"

cat << 'EOF' > $PREFIX/bin/uninstall-casaos
#!/data/data/com.termux/files/usr/bin/bash
echo -e "\033[0;31mUninstalling CasaOS...\033[0m"
rm -rf ~/.casaos
rm -rf $PREFIX/etc/casaos
rm -f $PREFIX/bin/casaos
rm -f $PREFIX/bin/casaos-supervisor
echo -e "\033[0;32mCasaOS uninstalled.\033[0m"
EOF
chmod +x $PREFIX/bin/uninstall-casaos
