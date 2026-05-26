#!/usr/bin/env bash
# ============================================================
# WhiteoutProjectOS -- Universal One-Click Installer
# Run on a fresh Ubuntu/Debian machine as root.
# ============================================================
set -e

if [ "$EUID" -ne 0 ]; then
  echo -e "\033[0;31m[ERROR] Please run this installer as root (sudo bash install.sh)\033[0m"
  exit 1
fi

echo -e "\033[0;32m"
echo "╔══════════════════════════════════════════════════════╗"
echo "║          WhiteoutProjectOS Installer                 ║"
echo "╚══════════════════════════════════════════════════════╝"
echo -e "\033[0m"
echo "WARNING: This script will deeply modify this machine."
echo "It is highly recommended to run this on a FRESH installation of Ubuntu/Debian."
echo ""

# The </dev/tty addition forces bash to listen to the keyboard even when piped!
read -p "Press ENTER to continue or Ctrl+C to cancel..." </dev/tty

# --- Gather User Inputs ---
read -p "Enter a username for the OS admin [default: wp-os-user]: " INPUT_USER </dev/tty
OS_USERNAME=${INPUT_USER:-wp-os-user}

read -s -p "Enter a password for the OS admin: " OS_PASSWORD </dev/tty
echo ""
read -s -p "Confirm password: " OS_PASSWORD_CONFIRM </dev/tty
echo ""

if [ "$OS_PASSWORD" != "$OS_PASSWORD_CONFIRM" ]; then
  echo -e "\033[0;31m[ERROR] Passwords do not match. Exiting.\033[0m"
  exit 1
fi

if [ -z "$OS_PASSWORD" ]; then
  OS_PASSWORD="password123"
  echo "[WARN] No password provided. Defaulting to 'password123'"
fi

read -p "Enter a hostname for this machine [default: wp-os-server]: " INPUT_HOST </dev/tty
OS_HOSTNAME=${INPUT_HOST:-wp-os-server}

# --- Internal Variables ---
REPO_BASE="https://raw.githubusercontent.com/TrelosLeras/os/main/wp-os-x86"
GITHUB_REPO="TrelosLeras/os"

# -- Bot repositories ----------------------------------------
# WOS Python bot
BOT_MAIN_PY="https://raw.githubusercontent.com/whiteout-project/bot/main/main.py"
BOT_INSTALL_PY="https://raw.githubusercontent.com/whiteout-project/install/main/install.py"

# WOS JavaScript bot
BOT_JS_REPO="https://github.com/whiteout-project/Whiteout-Survival-Discord-Bot"
BOT_JS_BRANCH="main"

# Kingshot bot
BOT_KINGSHOT_REPO="https://github.com/kingshot-project/Kingshot-Discord-Bot"
BOT_KINGSHOT_BRANCH="main"
BOT_KINGSHOT_INSTALL_PY="https://raw.githubusercontent.com/kingshot-project/Kingshot-Discord-Bot/main/install/install.py"

# WOS VoiceChat Counter bot
BOT_VOICECHAT_REPO="https://github.com/ikketimnl/wos-voicechat-counter"
BOT_VOICECHAT_BRANCH="main"

DEFAULT_BOT="wos-py"
DEFAULT_BOT_LABEL="Wos Bot"
BOTS_DIR="/home/${OS_USERNAME}/bots"
WEBSERVER_DIR="/opt/wp-os-webserver"
WEBSERVER_PORT="8080"

# --- Pre-Flight Dependency Check ---
echo "[INFO] Verifying native dependencies..."
apt-get update -qq
apt-get install -y -qq wget curl sudo

# Remove Snap version of curl to prevent sandbox errors
if command -v snap &> /dev/null && snap list curl &> /dev/null; then
  echo "[INFO] Removing sandboxed Snap version of curl..."
  snap remove curl
fi

# --- Download & Setup ---
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"
echo "[INFO] Downloading setup files..."

# Using wget guarantees we bypass any remaining curl sandbox quirks
wget -qO wp-os-provision.sh "${REPO_BASE}/rootfs-overlay/usr/local/bin/wp-os-provision.sh"
wget -qO wp-os-install-bot.sh "${REPO_BASE}/rootfs-overlay/usr/local/bin/wp-os-install-bot.sh"
wget -qO wp-os-bot-start.sh "${REPO_BASE}/rootfs-overlay/usr/local/bin/wp-os-bot-start.sh"
wget -qO wp-os-bot-manager.sh "${REPO_BASE}/rootfs-overlay/usr/local/bin/wp-os-bot-manager.sh"
wget -qO wp-os-update.sh "${REPO_BASE}/rootfs-overlay/usr/local/bin/wp-os-update.sh"

mkdir -p "${WEBSERVER_DIR}"
wget -qO "${WEBSERVER_DIR}/app.py" "${REPO_BASE}/webserver/app.py"

chmod +x wp-os-*.sh
mv wp-os-*.sh /usr/local/bin/

# --- Substitute Variables ---
echo "[INFO] Configuring provisioning script..."
sed -i \
    -e "s|@@OS_USERNAME@@|${OS_USERNAME}|g" \
    -e "s|@@OS_PASSWORD@@|${OS_PASSWORD}|g" \
    -e "s|@@OS_HOSTNAME@@|${OS_HOSTNAME}|g" \
    -e "s|@@GITHUB_REPO@@|${GITHUB_REPO}|g" \
    -e "s|@@BOT_MAIN_PY@@|${BOT_MAIN_PY}|g" \
    -e "s|@@BOT_INSTALL_PY@@|${BOT_INSTALL_PY}|g" \
    -e "s|@@BOT_JS_REPO@@|${BOT_JS_REPO}|g" \
    -e "s|@@BOT_JS_BRANCH@@|${BOT_JS_BRANCH}|g" \
    -e "s|@@BOT_KINGSHOT_REPO@@|${BOT_KINGSHOT_REPO}|g" \
    -e "s|@@BOT_KINGSHOT_BRANCH@@|${BOT_KINGSHOT_BRANCH}|g" \
    -e "s|@@BOT_KINGSHOT_INSTALL_PY@@|${BOT_KINGSHOT_INSTALL_PY}|g" \
    -e "s|@@BOT_VOICECHAT_REPO@@|${BOT_VOICECHAT_REPO}|g" \
    -e "s|@@BOT_VOICECHAT_BRANCH@@|${BOT_VOICECHAT_BRANCH}|g" \
    -e "s|@@DEFAULT_BOT@@|${DEFAULT_BOT}|g" \
    -e "s|@@DEFAULT_BOT_LABEL@@|${DEFAULT_BOT_LABEL}|g" \
    -e "s|@@BACKGROUND_IMAGE_URL@@|${BACKGROUND_IMAGE_URL}|g" \
    -e "s|@@DESKTOP@@|${DESKTOP}|g" \
    -e "s|@@BOTS_DIR@@|${BOTS_DIR}|g" \
    -e "s|@@WEBSERVER_DIR@@|${WEBSERVER_DIR}|g" \
    -e "s|@@WEBSERVER_PORT@@|${WEBSERVER_PORT}|g" \
    -e "s|@@REPO_BASE@@|${REPO_BASE}|g" \
    /usr/local/bin/wp-os-provision.sh

# --- Execute Provisioning ---
echo "[INFO] Starting WhiteoutProjectOS System Provisioning..."
export WPOS_REBOOT=1
bash /usr/local/bin/wp-os-provision.sh

# Cleanup
cd /
rm -rf "$TMP_DIR"
