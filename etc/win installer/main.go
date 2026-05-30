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

// distroName acts as the global ID for your Virtual Machine.
const distroName = "Ubuntu" 

// debugMode is toggled if the --debug flag is passed.
var debugMode bool = false

func main() {
	// --- GLOBAL DEBUG CHECK ---
	for _, arg := range os.Args {
		if arg == "--debug" {
			debugMode = true
			fmt.Println("[DEBUG] Debug mode enabled. Windows will not be hidden.")
		}
	}

	// --- COMMAND ROUTING ENGINE ---
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-stage1":
			runStage1()
		case "-stage2":
			runStage2()
		case "-stage3":
			runStage3()
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
		case "-launch":
			runLaunch()
		case "-uninstall":
			runUninstall()
		default:
			fmt.Println("Unknown command.")
		}
		return
	}
	runLaunch()
}

// --- CONFIGURATION READER ---
func getConfigUser() string {
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	data, err := os.ReadFile(filepath.Join(dir, "config.ini"))
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Username=") {
				return strings.TrimPrefix(line, "Username=")
			}
		}
	}
	return "wp-os-user" 
}

// --- STAGE 1: Enable WSL ---
func runStage1() {
	cmd := exec.Command("dism.exe", "/online", "/get-featureinfo", "/featurename:Microsoft-Windows-Subsystem-Linux")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "State : Enabled") {
		enable := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:Microsoft-Windows-Subsystem-Linux", "/all", "/norestart")
		// Hide window only if NOT in debug mode
		if !debugMode { enable.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		enable.Run()
	}
	os.Exit(0)
}

// --- STAGE 2: Enable VM Platform ---
func runStage2() {
	cmd := exec.Command("dism.exe", "/online", "/get-featureinfo", "/featurename:VirtualMachinePlatform")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "State : Enabled") {
		enable := exec.Command("dism.exe", "/online", "/enable-feature", "/featurename:VirtualMachinePlatform", "/all", "/norestart")
		if !debugMode { enable.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		enable.Run()
	}
	os.Exit(0)
}

// --- STAGE 3: Install Ubuntu (Smart) ---
// Performs installation with visible progress tracking and validation.
func runStage3() {
	wslPath := "C:\\Windows\\System32\\wsl.exe" 
	
	// Check for existing installation to handle potential overwrites.
	checkCmd := exec.Command(wslPath, "-l", "-q")
	output, _ := checkCmd.CombinedOutput()
	cleanOutput := strings.ReplaceAll(string(output), "\x00", "")
	
	if strings.Contains(cleanOutput, distroName) {
		promptMsg := "Ubuntu is already installed. Do you want to wipe it and install fresh?"
		psCmd := fmt.Sprintf(`Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.MessageBox]::Show("%s", "Overwrite?", 4, 48)`, promptMsg)
		
		boxCmd := exec.Command("powershell", "-c", psCmd)
		if !debugMode { boxCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
		result, _ := boxCmd.Output()
		
		if strings.TrimSpace(string(result)) == "Yes" {
			exec.Command(wslPath, "--unregister", distroName).Run()
		} else {
			os.Exit(0) 
		}
	}

	// Perform the installation with output streaming if in debug mode.
	cmd := exec.Command(wslPath, "--install", "-d", distroName, "--no-launch")
	if debugMode {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	
	fmt.Println("[DEBUG] Starting WSL installation... please wait.")
	cmd.Run()
    
	// Validate that the installation completed by checking for the distro.
	fmt.Println("[DEBUG] Checking if installation finished...")
	for i := 0; i < 10; i++ {
		check := exec.Command(wslPath, "-l", "-q")
		out, _ := check.Output()
		if strings.Contains(strings.ReplaceAll(string(out), "\x00", ""), distroName) {
			fmt.Println("[DEBUG] Installation successful!")
			os.Exit(0)
		}
		time.Sleep(5 * time.Second) // Wait 5 seconds between checks for distro registration.
	}
	
	fmt.Println("[DEBUG] Timeout: Installation did not appear in WSL list.")
	os.Exit(1) 
}

// --- STAGE 4: Provision System ---
func runStage4(username string, password string, hostname string) {
	wslPath := "C:\\Windows\\System32\\wsl.exe" 
	// Redirect output to log for debugging
	installCmd := fmt.Sprintf("curl -sSL https://raw.githubusercontent.com/TrelosLeras/os/main/etc/install.sh | bash -s -- \"%s\" \"%s\" \"%s\" 2>&1 | tee /tmp/install.log", username, password, hostname)
	setupCmd := exec.Command(wslPath, "-d", distroName, "-u", "root", "-e", "bash", "-c", installCmd)
	if !debugMode { setupCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
	setupCmd.Run()
	if debugMode { fmt.Println("Provisioning complete. Check /tmp/install.log in WSL.") }
	os.Exit(0)
}

// --- LAUNCHER MODE ---
func runLaunch() {
    username := getConfigUser()
    // Fallback to root if user is missing
    checkUser := exec.Command("wsl", "-d", distroName, "-u", username, "id")
    if err := checkUser.Run(); err != nil { username = "root" }
    
    // This is the full script that includes your branding and warning banner
    bashScript := `echo -e '\e[92m========================================\n      WhiteoutProjectOS is Active!      \n========================================\e[0m\n\n\e[91m[!] WARNING: DO NOT CLOSE THIS WINDOW [!]\e[0m\nClosing this terminal may shut down your background services.\n\n' && exec bash`

    // Inject bashScript into the WT (Windows Terminal) command
    rawWt := fmt.Sprintf(`wt.exe --title "WhiteoutProjectOS Server" -- wsl.exe -d %s -u %s -e bash -c "%s"`, distroName, username, bashScript)
    
    wtCmd := exec.Command("wt.exe")
    if !debugMode { 
        wtCmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: rawWt} 
    }
    wtCmd.Start()
    
    time.Sleep(2 * time.Second)
    exec.Command("cmd.exe", "/c", "start", "http://localhost:8080").Start()
}

// --- UNINSTALLER MODE ---
func runUninstall() {
	exePath, _ := os.Executable()
	os.Remove(filepath.Join(filepath.Dir(exePath), "config.ini"))
	exec.Command("wsl", "--unregister", distroName).Run()
}