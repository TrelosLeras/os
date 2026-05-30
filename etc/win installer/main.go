package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var (
	distroName = "Ubuntu"
	logFile    = "C:\\Windows\\Temp\\wos_install.log"
	debugMode  = false
)

func main() {
	if len(os.Args) < 2 {
		return
	}

	arg := os.Args[1]

	switch arg {
	case "-stage1":
		runStage1()
	case "-stage2":
		runStage2()
	case "-stage3":
		runStage3()
	case "-stage4":
		if len(os.Args) >= 5 {
			runStage4(os.Args[2], os.Args[3], os.Args[4])
		}
	case "-launch":
		runLaunch()
	case "-uninstall":
		runUninstall()
	}
}

func runStage1() {
	cmd := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:Microsoft-Windows-Subsystem-Linux", "/all", "/norestart")
	if !debugMode {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Run()
}

func runStage2() {
	cmd := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:VirtualMachinePlatform", "/all", "/norestart")
	if !debugMode {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Run()
}

func runStage3() {
	wslPath := "C:\\Windows\\System32\\wsl.exe"

	checkCmd := exec.Command(wslPath, "-l", "-q")
	if !debugMode { checkCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	output, _ := checkCmd.CombinedOutput()

	if strings.Contains(strings.ReplaceAll(string(output), "\x00", ""), distroName) {
		msg := "Ubuntu is already installed. How do you want to proceed?\n\nYes - Overwrite (Wipes existing data)\nNo - Install OS Here (Keep existing data)\nCancel - Abort Installation"
		psCmd := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show("%s", "Existing Subsystem Detected", 3, 32)`, msg)
		
		boxCmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psCmd)
		if !debugMode { boxCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		
		result, _ := boxCmd.Output()
		choice := string(result)

		if strings.Contains(choice, "Yes") {
			// Unregister the old environment
			unreg := exec.Command(wslPath, "--unregister", distroName)
			if !debugMode { unreg.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
			unreg.Run()
			
			// Actively wait for Windows to finish deleting the massive VHDX file
			for i := 0; i < 15; i++ {
				check := exec.Command(wslPath, "-l", "-q")
				if !debugMode { check.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
				out, _ := check.Output()
				if !strings.Contains(strings.ReplaceAll(string(out), "\x00", ""), distroName) { 
					break // The drive is gone, safe to proceed!
				}
				time.Sleep(2 * time.Second)
			}
			
		} else if strings.Contains(choice, "No") {
			outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				outFile.WriteString("\n[System] Existing Ubuntu environment retained.\n")
				outFile.Close()
			}
			os.Exit(0)
		} else {
			outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				outFile.WriteString("\n[ABORT] Installation cancelled.\n")
				outFile.Close()
			}
			os.Exit(1)
		}
	}

	// --- FRESH INSTALL BLOCK ---
	cmd := exec.Command(wslPath, "--install", "-d", distroName, "--no-launch")
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

	outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		outFile.WriteString("\n[System] Installing Ubuntu Environment. Please wait...\n")
		outFile.WriteString("[System] Downloading and extracting OS packages... (This may take 2 to 10 minutes)\n")
		outFile.Close() 
	}

	errRun := cmd.Run()

	outFile2, err2 := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 == nil {
		if errRun != nil {
			outFile2.WriteString(fmt.Sprintf("\n[ERROR] Ubuntu download failed: %v\n", errRun))
			outFile2.WriteString("[ABORT] Installation cancelled due to system error.\n")
			outFile2.Close()
			os.Exit(1)
		}
		outFile2.WriteString("\n[System] Ubuntu Virtual Machine successfully installed!\n")
		outFile2.Close()
	}

	// Validation loop
	for i := 0; i < 10; i++ {
		check := exec.Command(wslPath, "-l", "-q")
		if !debugMode { check.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		out, _ := check.Output()
		if strings.Contains(strings.ReplaceAll(string(out), "\x00", ""), distroName) { os.Exit(0) }
		time.Sleep(5 * time.Second)
	}
	
	// Failsafe: Trigger the Restart abort message
	outFile3, _ := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	outFile3.WriteString("\n[ABORT] System validation failed. A computer restart is required to finish enabling WSL.\n")
	outFile3.Close()
	os.Exit(1)
}

func runStage4(username, password, hostname string) {
	wslPath := "C:\\Windows\\System32\\wsl.exe" 
	
	installCmd := fmt.Sprintf(
		`export DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a; ` +
		`echo 'Dpkg::Options { "--force-confdef"; "--force-confold"; }; APT::Get::Assume-Yes "true"; APT::Get::quiet "true";' > /etc/apt/apt.conf.d/99wos-headless; ` +
		
		// THE SPEED HACK & PROMPT KILLER
		`rm -rf /var/cache/man/*; echo "man-db man-db/auto-update boolean false" | debconf-set-selections; ` +
		`rm -f /etc/apt/keyrings/nodesource.gpg; ` + 
		
		// --- THE TERMINAL WARNING MESSAGE (Green Header & Red Warning) ---
		// We use \e[1m (Bold), \e[32m (Green), and \e[31m (Red) stacked without semicolons!
		`echo 'echo -e "\n\e[1m\e[32m=======================================\e[0m"' > /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[32m      WhiteoutProjectOS is Active!     \e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[32m=======================================\n\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[31m[!] WARNING: DO NOT CLOSE THIS WINDOW [!]\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[31mClosing this terminal wil terminate background services.\n\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`chmod +x /etc/profile.d/99-wos-warning.sh; ` +
		// -----------------------------------------------------------------
		
		`curl -sSL https://raw.githubusercontent.com/TrelosLeras/os/main/etc/install.sh | DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a bash -s -- "%s" "%s" "%s"; ` +
		`rm -f /etc/apt/apt.conf.d/99wos-headless`, 
		username, password, hostname)
		
	setupCmd := exec.Command(wslPath, "-d", distroName, "-u", "root", "-e", "bash", "-c", installCmd)
	if !debugMode { setupCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	
	outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		setupCmd.Stdout = outFile
		setupCmd.Stderr = outFile
		outFile.WriteString("\n[System] Provisioning WhiteoutProjectOS services...\n")
	}
	setupCmd.Run()
	if err == nil { outFile.Close() } 
	os.Exit(0)
}

func runLaunch() {
	// 1. Pop open the visible terminal window (This boots the Linux kernel & background services)
	cmd := exec.Command("cmd.exe", "/c", "start", "wsl.exe", "-d", distroName, "--cd", "~")
	cmd.Run()

	// 2. Give the backend services 3 seconds to fully initialize the web server
	time.Sleep(3 * time.Second)

	// 3. Now that the server is ready, open the Web Dashboard!
	exec.Command("cmd.exe", "/c", "start", "http://localhost:8080").Run()
}

func runUninstall() {
	// Silently unregisters and deletes the Linux VM when the user uninstalls the app
	cmd := exec.Command("C:\\Windows\\System32\\wsl.exe", "--unregister", distroName)
	if !debugMode {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Run()
}