package main

import (
	"bufio"
	"embed"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var embeddedScript embed.FS

func runCommand(name string, args ...string) error {
	fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runSudo(command string) error {
	fmt.Print("Enter sudo password: ")
	password, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	password = strings.TrimSpace(password)
	fmt.Println("Password entered:", password)

	cmd := exec.Command("powershell", "-Command", command)

	stdin, _ := cmd.StdinPipe()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting:", err)

	}

	// Write password into sudo
	stdin.Write([]byte(password + "\n"))
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error:", err)
	}

	return cmd.Run()
}

func runPowershell(command string) error {
	return runCommand("powershell", "-Command", command)
}

func main() {
	fmt.Println("Starting setup...")

	// Step 1: Ensure WSL exists
	fmt.Println(" Checking for WSL...")
	err := runPowershell("wsl --status")
	if err != nil {
		fmt.Println(" Installing WSL with Ubuntu...")
		runPowershell("wsl --install -d Ubuntu")

		// Step 1a: Register app to run again after reboot
		exePath, _ := os.Executable()
		runPowershell(fmt.Sprintf(
			`New-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\RunOnce' -Name 'SetupResume' -Value '"%s"' -PropertyType String -Force`,
			exePath,
		))

		fmt.Println("Rebooting system to finish WSL installation...")
		runPowershell("shutdown /r /t 5")
		return
	}

	runSudo(`wsl -d Ubuntu -- bash -c "sudo -S apt update && sudo -S apt install -y docker.io"`)

	// Get the directory where the executable is located
	exePath, err := os.Getwd()
	if err != nil {
		fmt.Println(" Failed to get executable path:", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	scriptPath := filepath.Join(exeDir, "startup-in-a-box/buildImage.sh")

	// Check if the script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Printf(" Script not found: %s\n", scriptPath)
		fmt.Println(" Please make sure buildImage.sh is in the same directory as the executable.")
		return
	}
	fmt.Printf(" Script  found: %s\n", scriptPath)
	// Read the script content
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		fmt.Println(" Failed to read script file:", err)
		return
	}

	tmpPath := "/tmp/buildImage.sh"

	// Write script to WSL using a simpler approach
	// First, create the file with content using echo
	encoded := base64.StdEncoding.EncodeToString(scriptContent)
	psWrite := fmt.Sprintf(
		`wsl -d Ubuntu -- bash -c "echo '%s' | base64 -d > %s && chmod +x %s"`,
		encoded, tmpPath, tmpPath,
	)

	runSudo(psWrite)

	fmt.Println(" Running script inside WSL...")
	err = runPowershell(fmt.Sprintf(`wsl -d Ubuntu -- bash %s`, tmpPath))

	if err != nil {
		fmt.Println(" Script execution failed:", err)
	} else {
		fmt.Println(" Setup completed successfully!")
	}

	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
