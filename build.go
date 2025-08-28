package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSMessage struct {
	Type    string `json:"type"`    // "output", "error", "prompt", "done"
	Content string `json:"content"` // actual text
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	defer conn.Close()

	if err := Setup(conn); err != nil {
		sendMessage(conn, "error", err.Error())
	}
}

// Send message to browser
func sendMessage(conn *websocket.Conn, msgType, content string) {
	msg := WSMessage{Type: msgType, Content: content}
	conn.WriteJSON(msg)
}

// Wait for input from browser
func waitForInput(conn *websocket.Conn, prompt string) (string, error) {
	sendMessage(conn, "prompt", prompt)

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return "", err
		}
		if msg.Type == "input" {
			return msg.Content, nil
		}
	}
}

// Run command and stream output to websocket
func runCommandWS(conn *websocket.Conn, name string, args ...string) error {
	cmdStr := fmt.Sprintf("Running: %s %s", name, strings.Join(args, " "))
	sendMessage(conn, "output", cmdStr)

	cmd := exec.Command(name, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		sendMessage(conn, "error", "Failed to start command: "+err.Error())
		return err
	}

	done := make(chan error)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			sendMessage(conn, "output", scanner.Text())
		}
		done <- nil
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			sendMessage(conn, "error", scanner.Text())
		}
		done <- nil
	}()

	<-done
	if err := cmd.Wait(); err != nil {
		sendMessage(conn, "error", "Command finished with error: "+err.Error())
		return err
	}

	sendMessage(conn, "output", "Command finished successfully.")
	return nil
}

// Convert Windows path to WSL path
func windowsToWslPath(winPath string) string {
	path := strings.ReplaceAll(winPath, "\\", "/")
	if len(path) > 1 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		path = "/mnt/" + drive + path[2:]
	}
	return path
}

// Full Setup workflow
func Setup(conn *websocket.Conn) error {
	sendMessage(conn, "output", "Starting setup...")

	// Check WSL
	sendMessage(conn, "output", "Checking for WSL...")
	if err := runCommandWS(conn, "powershell", "-Command", "wsl --status"); err != nil {
		sendMessage(conn, "output", "Installing WSL with Ubuntu...")

		runCommandWS(conn, "powershell", "-Command", "wsl --install -d Ubuntu")

		// Register app to run again after reboot
		exePath, _ := os.Executable()
		runCommandWS(conn, "powershell", "-Command",
			fmt.Sprintf(`New-ItemProperty -Path 'HKCU:\Software\Microsoft\Windows\CurrentVersion\RunOnce' -Name 'SetupResume' -Value '"%s"' -PropertyType String -Force`, exePath))

		sendMessage(conn, "output", "Rebooting system to finish WSL installation...")
		runCommandWS(conn, "powershell", "-Command", "shutdown /r /t 5")
		return nil
	}

	// Find script
	exeDir, _ := os.Getwd()
	scriptPath := filepath.Join(exeDir, "/buildImage.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		sendMessage(conn, "error", "Script not found: "+scriptPath)
		return err
	}
	sendMessage(conn, "output", "Script found: "+scriptPath)

	// Ask for sudo password
	password, err := waitForInput(conn, "Enter sudo password:")
	if err != nil {
		return err
	}
	sendMessage(conn, "output", "Password received.")

	// Run script in WSL
	linuxPath := windowsToWslPath(scriptPath)
	cmdStr := fmt.Sprintf(`wsl -d Ubuntu -- bash -c "echo '%s' | sudo -S bash %s"`, password, linuxPath)
	runCommandWS(conn, "powershell", "-Command", cmdStr)

	sendMessage(conn, "done", "Setup completed successfully.")
	return nil
}
