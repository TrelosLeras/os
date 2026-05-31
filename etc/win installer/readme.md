# ❄️WhiteoutProjectOS - Windows Deployment Toolchain

Welcome to the Windows Deployment for **WhiteoutProjectOS**. 

While WhiteoutProjectOS is natively designed for high-performance bare-metal deployments, the code in this directory provides a fully automated toolchain to deploy a sandboxed, virtualized version of the OS directly on top of Windows using Windows Subsystem for Linux (WSL).

This toolchain consists of a **Go (Golang)** backend that manages the Windows Kernel and WSL lifecycle, and an **Inno Setup** frontend that provides a seamless, 60FPS graphical installation wizard.

---

## 📂 Directory Contents

* `WhiteoutProjectOS_Installer.exe` - **The pre-compiled, ready-to-use Windows installer.**
* `main.go` - The backend system manager. Handles WSL provisioning, Windows feature toggling (DISM), and headless script injection.
* `setup.iss` - The Pascal-based Inno Setup script. Manages the UI, custom credential inputs, NTFS file-lock circumvention, and dynamic error handling.
* `icon.ico` - Visual assets for the Windows executable.

---

## ✨ Installer Features

* **Isolated Subsystem:** Utilizes WSL's custom naming architecture to spin up a standalone VM (`WhiteoutProjectOS`) without touching or overwriting a user's existing Linux or WSL environments.
* **Automated Prerequisites:** Automatically enables the *Windows Subsystem for Linux* and *Virtual Machine Platform* features on the host machine, gracefully handling required system reboots.
* **Headless Provisioning:** Injects a custom `99wos-headless` configuration into the APT package manager, enabling blazing-fast, silent installations of massive dependencies (like Node.js) with zero user prompts.
* **Smart Dashboard Integration:** When launched, it automatically spins up the background services and opens the custom Web Control Panel (`localhost:8080`) in the user's default Windows browser alongside the terminal.
* **Custom Terminal MOTD:** Injects a custom, color-coded warning message directly into the Linux `/etc/profile.d/` directory to greet users natively when the terminal opens.

---

## 🚀 Usage (End Users)

Distributing and installing the OS on Windows is incredibly simple:

1. Double-click the included **`WhiteoutProjectOS_Installer.exe`** in this directory.
2. The setup wizard guides you through configuring custom Server Credentials (Username, Password, Hostname) or leave the default.
3. Choose the program's installation folder.
4. The automated backend provisions the environment (safely pausing to request a restart if Windows virtualization features need to be enabled).
5. Launching **WhiteoutProjectOS** from your Start Menu boots the Linux kernel, waits 3 seconds for Node.js services to initialize, and automatically opens the web dashboard!

---

## 🧹 Uninstallation (Windows)
WhiteoutProjectOS is designed to leave no trace. Running the uninstaller (via Windows Settings or the Start Menu shortcut) will trigger the Go backend to completely unregister and delete the isolated virtual hard drive before cleaning up all Windows shortcuts and configuration files.

---

## 🛠️ Building the Installer from Source

To compile the `WhiteoutProjectOS_Installer.exe` yourself, you will need to install the required build tools on your Windows machine:

* **[Download Go (Golang)](https://go.dev/dl/)** - Required to compile the backend executable.
* **[Download Inno Setup 6](https://jrsoftware.org/isdl.php)** - Required to compile the frontend installer.

### 1. Compile the Go Backend
The Go backend must be compiled as a hidden GUI application so it doesn't flash empty command prompt windows during the installation process. Open your terminal in this directory and run:
```bash
go build -H=windowsgui -o WhiteoutProjectOS.exe main.go
```
### 2. Compile the Frontend Wizard
Open setup.iss in Inno Setup.

Ensure the newly compiled WhiteoutProjectOS.exe, along with icon.ico and logo.png, are present in this directory.

Click Compile at the top of the Inno Setup window.

The final WhiteoutProjectOS_Installer.exe will be generated and placed in a new folder alongside the source files.
