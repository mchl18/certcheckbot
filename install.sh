#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing SSL Certificate Checker...${NC}"

# Determine system info
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Map architecture names
case ${ARCH} in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: ${ARCH}${NC}"
        exit 1
        ;;
esac

# Create .certchecker directory structure
INSTALL_DIR="$HOME/.certchecker"
echo -e "${BLUE}Creating directory structure in ${INSTALL_DIR}...${NC}"

mkdir -p "${INSTALL_DIR}/"{bin,config,logs,data}

# Get the latest release URL
echo -e "${BLUE}Fetching latest release...${NC}"
LATEST_RELEASE_URL=$(curl -s https://api.github.com/repos/mchl18/ssl-expiration-check-bot/releases/latest | grep "browser_download_url.*${OS}-${ARCH}.tar.gz" | cut -d '"' -f 4)

if [ -z "$LATEST_RELEASE_URL" ]; then
    echo -e "${RED}Failed to find release for ${OS}-${ARCH}${NC}"
    exit 1
fi

# Download and extract the release
echo -e "${BLUE}Downloading ${OS}-${ARCH} release...${NC}"
curl -L --progress-bar -o "${INSTALL_DIR}/release.tar.gz" "${LATEST_RELEASE_URL}"

echo -e "${BLUE}Extracting release...${NC}"
cd "${INSTALL_DIR}"
tar xzf release.tar.gz
mv bin/*/* bin/
rm -rf bin/*/
rm release.tar.gz

# Make binary executable
chmod +x "${INSTALL_DIR}/bin/certchecker"

echo -e "${GREEN}Installation complete!${NC}"
echo
echo -e "${BLUE}To complete the installation:${NC}"
echo
echo -e "1. ${BLUE}Add to your PATH by adding this line to your shell config (.bashrc, .zshrc, etc.):${NC}"
echo -e "   ${GREEN}export PATH=\"\$PATH:\$HOME/.certchecker/bin\"${NC}"
echo
echo -e "2. ${BLUE}Configure the service by running:${NC}"
echo -e "   ${GREEN}certchecker -configure${NC}" 