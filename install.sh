#!/bin/bash

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     os=linux;;
    Darwin*)    os=darwin;;
    CYGWIN*)    os=windows;;
    MINGW*)     os=windows;;
    MSYS*)      os=windows;;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64*)    arch=amd64;;
    amd64*)     arch=amd64;;
    i386*)      arch=386;;
    i686*)      arch=386;;
    arm64*)     arch=arm64;;
    aarch64*)   arch=arm64;;
    arm*)       arch=arm;;
    *)          echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

# Set file extension based on OS
if [ "$os" = "windows" ]; then
    ext="zip"
else
    ext="tar.gz"
fi

# Get latest release URL from GitHub API
echo "Detecting latest version..."
LATEST_URL=$(curl -s https://api.github.com/repos/mchl18/certcheckbot/releases/latest | grep "browser_download_url.*-$os-$arch.$ext" | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
    echo "Error: Could not find release for $os-$arch"
    exit 1
fi

echo "Downloading certchecker for $os-$arch..."
curl -L -o "certchecker-$os-$arch.$ext" "$LATEST_URL"

# Create install directory
install_dir="$HOME/.certchecker"
mkdir -p "$install_dir"

# Extract the archive
echo "Installing to $install_dir..."
if [ "$os" = "windows" ]; then
    unzip -o "certchecker-$os-$arch.$ext" -d "$install_dir"
else
    tar xzf "certchecker-$os-$arch.$ext" -C "$install_dir"
fi

# Clean up downloaded archive
rm "certchecker-$os-$arch.$ext"

# Move binary to bin directory
mkdir -p "$install_dir/bin"
if [ "$os" = "windows" ]; then
    mv "$install_dir/certchecker.exe" "$install_dir/bin/"
else
    mv "$install_dir/certchecker" "$install_dir/bin/"
fi

# Move .env to config directory
mkdir -p "$install_dir/config"
mv "$install_dir/.env" "$install_dir/config/"

echo "Installation complete!"
echo "Certchecker installed to: $install_dir/bin"
echo "Configuration file at: $install_dir/config/.env"
echo
echo "Please add the following to your shell configuration (.bashrc, .zshrc, etc.):"
echo "export PATH=\"\$PATH:$install_dir/bin\""
echo
echo "Don't forget to edit $install_dir/config/.env with your configuration!" 