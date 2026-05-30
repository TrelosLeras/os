package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const distroName = "Ubuntu" 
const logFile = "C:\\Windows\\Temp\\wos_install.log"
var debugMode bool = false

func main() {
	for _, arg := range os.Args {
		if arg == "--debug" { debugMode = true }
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-stage1": runStage1()
		case "-stage2": runStage2()
		case "-stage3": runStage3()
		case "-stage4":
			username := "wp-os-user"
			password := "Password123"
			hostname := "wp-os-server"
			if len(os.Args) >= 5 {
				username = os.Args[2]
				password = os.Args[3]
				hostname = os.Args[4]
			}
			runStage4(username, password, hostname)
		case "-launch": runLaunch()
		case "-uninstall": runUninstall()
		}
		return
	}
	runLaunch()
}

func getConfigUser() string {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	data, err := os.ReadFile(filepath.Join(dir, "config.ini"))
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Username=") { return strings.TrimPrefix(line, "Username=") }
		}
	}
	return "wp-os-user" 
}

func runStage1() {
	cmd := exec.Command("dism.exe", "/online", "/get-featureinfo", "/featurename:Microsoft-Windows-Subsystem-Linux")
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	output, _ := cmd.CombinedOutput()
	
	if !strings.Contains(string(output), "State : Enabled") {
		enable := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:Microsoft-Windows-Subsystem-Linux", "/all", "/norestart")
		if !debugMode { enable.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		enable.Run()
	}
	os.Exit(0)
}

func runStage2() {
	cmd := exec.Command("dism.exe", "/online", "/get-featureinfo", "/featurename:VirtualMachinePlatform")
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	output, _ := cmd.CombinedOutput()
	
	if !strings.Contains(string(output), "State : Enabled") {
		enable := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:VirtualMachinePlatform", "/all", "/norestart")
		if !debugMode { enable.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		enable.Run()
	}
	os.Exit(0)
}

func runStage3() {
	wslPath := "C:\\Windows\\System32\\wsl.exe"
	logFile := "C:\\Windows\\Temp\\wos_install.log"

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
			unreg := exec.Command(wslPath, "--unregister", distroName)
			if !debugMode { unreg.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
			unreg.Run()
			time.Sleep(3 * time.Second)
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
				outFile.WriteString("\n[ABORT] Installation cancelled by user.\n")
				outFile.Close()
			}
			os.Exit(1)
		}
	}

	// --- FRESH INSTALL BLOCK ---
	cmd := exec.Command(wslPath, "--install", "-d", distroName, "--no-launch")
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }

	// Open the file, write the message ONCE, and immediately close it
	outFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		outFile.WriteString("\n[System] Installing Ubuntu Environment. Please wait...\n")
		outFile.WriteString("[System] Downloading and extracting OS packages... (This may take 2 to 10 minutes)\n")
		outFile.Close() 
	}

	// This completely blocks the Go script until the installation finishes
	cmd.Run()

	// Once it finishes, open the log again and write the success message
	outFile2, err2 := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err2 == nil {
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
	os.Exit(1)
}

func runStage4(username, password, hostname string) {
	wslPath := "C:\\Windows\\System32\\wsl.exe" 
	
	installCmd := fmt.Sprintf(
		`export DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a; ` +
		`echo 'Dpkg::Options { "--force-confdef"; "--force-confold"; }; APT::Get::Assume-Yes "true"; APT::Get::quiet "true";' > /etc/apt/apt.conf.d/99wos-headless; ` +
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
    username := getConfigUser()
    
    checkUser := exec.Command("wsl", "-d", distroName, "-u", username, "id")
    if !debugMode { checkUser.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
    if err := checkUser.Run(); err != nil { username = "root" }
    
    bashScript := `echo -e '\e[92m========================================\n      WhiteoutProjectOS is Active!      \n========================================\e[0m\n\n\e[91m[!] WARNING: DO NOT CLOSE THIS WINDOW [!]\e[0m\nClosing this terminal may shut down your background services.\n\n' && exec bash`
    rawWt := fmt.Sprintf(`wt.exe --title "WhiteoutProjectOS Server" -- wsl.exe -d %s -u %s -e bash -c "%s"`, distroName, username, bashScript)
    
    wtCmd := exec.Command("wt.exe")
    if !debugMode { wtCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: rawWt} }
    wtCmd.Start()
    
    time.Sleep(2 * time.Second)
    
    browserCmd := exec.Command("cmd.exe", "/c", "start", "http://localhost:8080")
    if !debugMode { browserCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
    browserCmd.Start()
}

func runUninstall() {
	exePath, _ := os.Executable()
	os.Remove(filepath.Join(filepath.Dir(exePath), "config.ini"))
	
	cmd := exec.Command("wsl", "--unregister", distroName)
	if !debugMode { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	cmd.Run()
}