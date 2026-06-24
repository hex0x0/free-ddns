#!/usr/bin/env bash
set -euo pipefail

# build-deb.sh - Build a Debian package for free-ddns
#
# This script creates a .deb package that:
# - Installs the binary to /usr/local/bin
# - Installs config to /usr/local/share/free-ddns/config.yaml
# - Installs systemd unit file to /lib/systemd/system
# - Does NOT automatically install the package

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${SCRIPT_DIR}"

BINARY_NAME="free-ddns"
PACKAGE_NAME="free-ddns"

# Read version from version.go
VERSION=$(grep 'const version' "${PROJECT_ROOT}/version.go" | sed 's/.*"\(.*\)".*/\1/')
if [[ -z "${VERSION}" ]]; then
    VERSION="1.0.0"
fi

# Architecture detection
ARCH=$(dpkg --print-architecture 2>/dev/null || echo "amd64")

# Debian package structure
BUILD_DIR="${PROJECT_ROOT}/build"
PKG_ROOT="${BUILD_DIR}/pkgroot"
DEBIAN_DIR="${PKG_ROOT}/DEBIAN"

DEB_FILENAME="${PACKAGE_NAME}_${ARCH}.deb"

rm -rf ${DEB_FILENAME}

echo "==> Building ${PACKAGE_NAME} version ${VERSION} for ${ARCH}"

# Clean and create build directory
rm -rf "${BUILD_DIR}"
mkdir -p "${DEBIAN_DIR}"

# Create directory structure
mkdir -p "${PKG_ROOT}/usr/local/bin"
mkdir -p "${PKG_ROOT}/usr/local/share/${PACKAGE_NAME}"

# Build binary
echo "==> Building ${BINARY_NAME} binary..."
rm -rf "${PROJECT_ROOT}/${BINARY_NAME}"
go build -ldflags "-s -w" -o "${BINARY_NAME}" .

# Copy binary to package staging area
echo "==> Packaging binary..."
cp "${PROJECT_ROOT}/${BINARY_NAME}" "${PKG_ROOT}/usr/local/bin/${BINARY_NAME}"
chmod 755 "${PKG_ROOT}/usr/local/bin/${BINARY_NAME}"

# Copy default config file
echo "==> Packaging default config..."
if [[ -f "${PROJECT_ROOT}/config/default_config.yaml" ]]; then
    cp "${PROJECT_ROOT}/config/default_config.yaml" "${PKG_ROOT}/usr/local/share/${PACKAGE_NAME}/config.yaml"
    chmod 644 "${PKG_ROOT}/usr/local/share/${PACKAGE_NAME}/config.yaml"
else
    echo "Warning: config/default_config.yaml not found" >&2
fi

# Create DEBIAN/control file
echo "==> Creating control file..."
cat > "${DEBIAN_DIR}/control" <<EOF
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Section: net
Priority: optional
Architecture: ${ARCH}
Maintainer: free-ddns <noreply@example.com>
Description: Dynamic DNS client
 A lightweight dynamic DNS client that supports multiple DNS providers
 including Tencent, Aliyun, and Cloudflare.
EOF

# Create postinst script to handle user-specific installation
echo "==> Creating postinst script..."
cat > "${DEBIAN_DIR}/postinst" <<'EOF'
#!/bin/bash
set -e

# This script runs as root and creates a default systemd unit when absent.

if [[ -n "${SUDO_USER}" ]] && [[ "${SUDO_USER}" != "root" ]]; then
    ACTUAL_USER="${SUDO_USER}"
else
    # If not run via sudo, try to detect from who invoked dpkg
    ACTUAL_USER="${USER}"
fi

if [[ -z "${ACTUAL_USER}" ]] || [[ "${ACTUAL_USER}" == "root" ]]; then
    ACTUAL_USER="root"
fi

USER_HOME=$(eval echo "~${ACTUAL_USER}")

# Get the actual group name
ACTUAL_GROUP=$(id -gn "${ACTUAL_USER}")

# Create systemd unit file with actual username only if it does not already exist
SYSTEMD_FILE_DST=/lib/systemd/system/free-ddns.service
if [[ ! -f ${SYSTEMD_FILE_DST} ]]; then
    echo "Installing systemd unit file to ${SYSTEMD_FILE_DST}"
    cat > ${SYSTEMD_FILE_DST} <<SYSTEMD_EOF
[Unit]
Description=A free DDNS client
Documentation=https://github.com/hex0x0/free-ddns
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
NoNewPrivileges=true
User=${ACTUAL_USER}
Group=${ACTUAL_GROUP}
Environment=HOME=${USER_HOME}
ExecStart=/usr/local/bin/free-ddns run -c /usr/local/share/free-ddns/config.yaml
Restart=on-failure
RestartPreventExitStatus=23

[Install]
WantedBy=multi-user.target
SYSTEMD_EOF

    chmod 644 /lib/systemd/system/free-ddns.service
else
    echo "Systemd unit file already exists at /lib/systemd/system/free-ddns.service, leaving it unchanged."
fi

# Reload systemd daemon
if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    echo "To enable the service at boot, run:"
    echo "  sudo systemctl enable free-ddns.service"
    echo "To start the service manually, run:"
    echo "  sudo systemctl start free-ddns.service"
fi

echo ""
echo "Installation complete!"
echo "Binary installed to: /usr/local/bin/free-ddns"
echo "Config file: /usr/local/share/free-ddns/config.yaml"
echo ""
echo "Next steps:"
echo "1. Edit your config: /usr/local/share/free-ddns/config.yaml"
echo "2. Enable the service: sudo systemctl enable free-ddns.service"
echo "3. Start the service manually when ready: sudo systemctl start free-ddns.service"

exit 0
EOF

chmod 755 "${DEBIAN_DIR}/postinst"

# Create prerm script to handle service removal
echo "==> Creating prerm script..."
cat > "${DEBIAN_DIR}/prerm" <<'EOF'
#!/bin/bash
set -e

# Stop the service if it's running
if command -v systemctl >/dev/null 2>&1; then
    systemctl stop free-ddns.service 2>/dev/null || true
    systemctl disable free-ddns.service 2>/dev/null || true
fi

exit 0
EOF

chmod 755 "${DEBIAN_DIR}/prerm"

# Build the .deb package
echo "==> Building .deb package..."
cd "${BUILD_DIR}"
dpkg-deb --build pkgroot "${DEB_FILENAME}"

# Move the .deb to project root
mv "${DEB_FILENAME}" "${PROJECT_ROOT}/"

echo ""
echo "==> Success! Package created: ${PROJECT_ROOT}/${DEB_FILENAME}"
echo ""
echo "To install the package, run:"
echo "  sudo dpkg -i ${DEB_FILENAME}"
echo ""
