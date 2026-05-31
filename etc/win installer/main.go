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
	sourceDistro = "Ubuntu"              // The template we download from Microsoft
	distroName   = "WhiteoutProjectOS"   // The custom name of YOUR isolated OS
	logFile      = "C:\\Windows\\Temp\\wos_install.log"
	debugMode    = false
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

	// Now it only checks if WhiteoutProjectOS exists, completely ignoring their personal Ubuntu!
	if strings.Contains(strings.ReplaceAll(string(output), "\x00", ""), distroName) {
		msg := "WhiteoutProjectOS is already installed. How do you want to proceed?\n\nYes - Overwrite (Wipes existing data)\nNo - Keep Existing Data\nCancel - Abort Installation"
		psCmd := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show("%s", "Existing Subsystem Detected", 3, 32)`, msg)
		
		boxCmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command", psCmd)
		if !debugMode { boxCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		
		result, _ := boxCmd.Output()
		choice := string(result)

		if strings.Contains(choice, "Yes") {
			unreg := exec.Command(wslPath, "--unregister", distroName)
			if !debugMode { unreg.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
			unreg.Run()
			
			for i := 0; i < 15; i++ {
				check := exec.Command(wslPath, "-l", "-q")
				if !debugMode { check.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
				out, _ := check.Output()
				if !strings.Contains(strings.ReplaceAll(string(out), "\x00", ""), distroName) { 
					break 
				}
				time.Sleep(2 * time.Second)
			}
			
		} else if strings.Contains(choice, "No") {
			outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				outFile.WriteString("\n[System] Existing WhiteoutProjectOS environment retained.\n")
				outFile.Close()
			}
			os.Exit(0)
		} else {
			outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				outFile.WriteString("\n[ABORT] Installation cancelled by user.\n")
				outFile.Close()
			}
			os.Exit(1)
		}
	}

	// --- FRESH INSTALL BLOCK (Now uses --name to isolate your OS!) ---
	cmd := exec.Command(wslPath, "--install", "-d", sourceDistro, "--name", distroName, "--no-launch")
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

	outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		outFile.WriteString("\n[System] Installing WhiteoutProjectOS Environment. Please wait...\n")
		outFile.WriteString("[System] Downloading and extracting OS packages... (This may take 2 to 10 minutes)\n")
		outFile.Close() 
	}

	errRun := cmd.Run()

	outFile2, err2 := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 == nil {
		if errRun != nil {
			outFile2.WriteString(fmt.Sprintf("\n[ERROR] Environment download failed: %v\n", errRun))
			outFile2.WriteString("[ABORT] Installation cancelled due to system error.\n")
			outFile2.Close()
			os.Exit(1)
		}
		outFile2.WriteString("\n[System] Virtual Machine successfully installed!\n")
		outFile2.Close()
	}

	for i := 0; i < 10; i++ {
		check := exec.Command(wslPath, "-l", "-q")
		if !debugMode { check.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		out, _ := check.Output()
		if strings.Contains(strings.ReplaceAll(string(out), "\x00", ""), distroName) { os.Exit(0) }
		time.Sleep(5 * time.Second)
	}
	
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
		
		`rm -rf /var/cache/man/*; echo "man-db man-db/auto-update boolean false" | debconf-set-selections; ` +
		`rm -f /etc/apt/keyrings/nodesource.gpg; ` + 
		
		// --- THE WINDOW TITLE LOCK ---
		// We inject this into the "skeleton" bashrc so any new user created gets the custom title!
		`echo 'PS1="$PS1\[\e]0;WhiteoutProjectOS\a\]"' >> /etc/skel/.bashrc; ` +
		`echo 'PS1="$PS1\[\e]0;WhiteoutProjectOS\a\]"' >> /root/.bashrc; ` +
		// -----------------------------

		// --- TERMINAL WARNING MESSAGE (Green Header & Red Warning) ---
		`echo 'echo -e "\n\e[1m\e[32m=======================================\e[0m"' > /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[32m      WhiteoutProjectOS is Active!     \e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[32m=======================================\n\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[31m[!] WARNING: DO NOT CLOSE THIS WINDOW [!]\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
		`echo 'echo -e "\e[1m\e[31mClosing this terminal will shut down background services.\n\e[0m"' >> /etc/profile.d/99-wos-warning.sh; ` +
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
	cmd := exec.Command("cmd.exe", "/c", "start", "wsl.exe", "-d", distroName, "--cd", "~")
	cmd.Run()
	time.Sleep(3 * time.Second)
	exec.Command("cmd.exe", "/c", "start", "http://localhost:8080").Run()
}

func runUninstall() {
	cmd := exec.Command("C:\\Windows\\System32\\wsl.exe", "--unregister", distroName)
	if !debugMode {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Run()
}