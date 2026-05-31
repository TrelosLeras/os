# ❄️ WhiteoutProjectOS

WhiteoutProjectOS is a standalone, fully automated Linux Subsystem designed to run seamlessly on top of Windows. Built with a bulletproof **Go (Golang)** backend and a beautiful **Inno Setup** frontend, it provisions an isolated Ubuntu-based virtual machine specifically tuned for custom backend services and web dashboards.

Unlike standard WSL installations, WhiteoutProjectOS runs in its own sandboxed environment, ensuring your personal Linux workspaces remain completely untouched.

---

## ✨ Key Features

* **Isolated Subsystem:** Utilizes WSL's custom naming architecture to spin up a standalone VM (`WhiteoutProjectOS`) without touching or overwriting a user's existing Ubuntu installation.
* **Automated Prerequisites:** The Go backend automatically communicates with the Windows Kernel via DISM to enable Windows Subsystem for Linux and the Virtual Machine Platform, gracefully handling required system reboots.
* **Headless Provisioning:** Injects a custom `99wos-headless` configuration into the APT package manager, enabling blazing-fast, completely silent installations of massive dependencies (like Node.js) with zero user prompts.
* **Smart Dashboard Integration:** Automatically spins up your background services and opens the custom Web Control Panel (`localhost:8080`) in the user's default Windows browser alongside the terminal.
* **Custom Terminal MOTD:** Injects a custom, color-coded warning message directly into the Linux `/etc/profile.d/` directory to greet users natively when the terminal opens.
* **Bulletproof Installer:** Features NTFS file-lock circumvention (The MFT Rename Trick), shallow-scan error parsing, and dynamic user prompts (Overwrite/Retain/Cancel) natively hooked into Windows Forms.

---

## 🏗️ Architecture / Tech Stack

* **Backend / System Manager:** `Go (Golang)` - Handles all WSL lifecycle commands, Windows feature management, and headless script injection.
* **Frontend / UI:** `Inno Setup (Pascal)` - Provides a smooth, 60FPS graphical installation wizard, custom credential inputs, and dynamic error handling.
* **Virtualization:** `Windows Subsystem for Linux (WSL2)`
* **Core OS:** `Ubuntu (Debian-based)`

---

## 🚀 Installation

### End Users
1. Download the latest `WhiteoutProjectOS_Installer.exe` from the [Releases](#) page.
2. Run the installer.
3. Choose your custom Server Credentials (Username, Password, Hostname) during setup or user the default.
4. Let the automated wizard provision your environment. (If your PC does not have WSL enabled, the setup will safely pause and ask you to restart your computer).
5. Launch **WhiteoutProjectOS** from your Desktop or Start Menu!

### Launch Behavior
When launched, the system will:
1. Boot the isolated Linux kernel in the background.
2. Open a dedicated terminal window (displaying the custom System MOTD).
3. Wait exactly 3 seconds for background Node.js services to initialize.
4. Automatically open your default web browser to the OS Dashboard (`http://localhost:8080`).

---

## 🛠️ Building from Source

If you want to contribute or compile the project yourself, you will need to install the required build tools on your Windows machine:
* **[Download Go (Golang)](https://go.dev/dl/)** - Required to compile the backend executable.
* **[Download Inno Setup 6](https://jrsoftware.org/isdl.php)** - Required to compile the frontend installer.

### 1. Compile the Go Backend
The Go backend must be compiled as a hidden GUI application so it doesn't flash empty command prompt windows during installation. Open your terminal in the project directory and run:
```bash
go build -H=windowsgui -o WhiteoutProjectOS.exe main.go
```

## 2. Compile the Frontend
Open setup.iss in Inno Setup.

Ensure the newly compiled WhiteoutProjectOS.exe, your icon.ico, and logo.png are in the same directory as the script.

Click Compile.

The output will be generated in your Documents folder as WhiteoutProjectOS_Installer.exe.

## 🧹 Uninstallation
WhiteoutProjectOS is designed to leave no trace. Running the uninstaller (via Windows Settings or the Start Menu shortcut) will trigger the Go backend to completely unregister and delete the isolated virtual hard drive before cleaning up all Windows shortcuts and configuration files.
