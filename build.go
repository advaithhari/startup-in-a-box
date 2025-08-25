package main

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
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

func runPowershell(command string) error {
	return runCommand("powershell", "-Command", command)
}

func main() {
	fmt.Println("Starting setup...")

	// Step 1: Ensure WSL exists
	fmt.Println("🔍 Checking for WSL...")
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


	fmt.Println(" Installing Docker in WSL (if missing)...")
	dockerInstallCmd := `wsl -d Ubuntu -- bash -c "command -v docker >/dev/null 2>&1 || (sudo apt update && sudo apt install -y docker.io)"`

	err = runPowershell(dockerInstallCmd)
	if err != nil {
		fmt.Println(" Failed to install Docker in WSL.")
		return
	}

	tmpPath := "/buildImage.sh"
	scriptContent, _ := embeddedScript.ReadFile("buildImage.sh")
	
	psWrite := fmt.Sprintf(
		`wsl -d Ubuntu -- bash -c "cat > %s <<'EOF'\n%s\nEOF\nchmod +x %s"`,
		tmpPath, string(scriptContent), tmpPath,
	)
	err = runPowershell(psWrite)
	if err != nil {
		fmt.Println(" Failed to write script inside WSL.")
		return
	}

	fmt.Println("▶ Running script inside WSL...")
	runPowershell(fmt.Sprintf(`wsl -d Ubuntu -- bash %s`, tmpPath))
}

