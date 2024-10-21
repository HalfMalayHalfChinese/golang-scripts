package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Windows API constants and structures
const (
	TokenQuery               = 0x0008
	TokenAdjustPrivileges    = 0x0020
	SePrivilegeEnabled       = 0x00000002
	SecurityMandatoryHighRid = 0x00002000
)

// Define structures for token privileges
type Luid struct {
	LowPart  uint32
	HighPart int32
}

type LuidAndAttributes struct {
	Luid       Luid
	Attributes uint32
}

type TokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]LuidAndAttributes
}

// Check if the process runs in high integrity level
func checkProcessIntegrityLevel() bool {
	var hToken syscall.Token
	err := syscall.OpenProcessToken(syscall.GetCurrentProcess(), TokenQuery, &hToken)
	if err != nil {
		fmt.Printf("[-] OpenProcessToken failed: %v\n", err)
		return false
	}
	defer hToken.Close()

	var infoSize uint32
	syscall.GetTokenInformation(hToken, syscall.TokenIntegrityLevel, nil, 0, &infoSize)

	info := make([]byte, infoSize)
	if err = syscall.GetTokenInformation(hToken, syscall.TokenIntegrityLevel, &info[0], infoSize, &infoSize); err != nil {
		fmt.Printf("[-] GetTokenInformation failed: %v\n", err)
		return false
	}

	// Extract integrity level
	label := (*syscall.TokenMandatoryLabel)(unsafe.Pointer(&info[0]))
	rid := *(*uint32)(unsafe.Pointer(uintptr(label.Label.Sid) + uintptr(8)))

	if rid >= SecurityMandatoryHighRid {
		fmt.Println("[+] High integrity level detected.")
		return true
	}

	fmt.Println("[-] Requires high integrity level.")
	return false
}

// Enable SeDebugPrivilege for the current process
func enableSeDebugPrivilege() bool {
	var hToken syscall.Token
	err := syscall.OpenProcessToken(syscall.GetCurrentProcess(), TokenAdjustPrivileges|TokenQuery, &hToken)
	if err != nil {
		fmt.Printf("[-] OpenProcessToken failed: %v\n", err)
		return false
	}
	defer hToken.Close()

	var luid Luid
	privilegeName, _ := syscall.UTF16PtrFromString("SeDebugPrivilege")
	if err = syscall.LookupPrivilegeValue(nil, privilegeName, &luid); err != nil {
		fmt.Printf("[-] LookupPrivilegeValue failed: %v\n", err)
		return false
	}

	tp := TokenPrivileges{
		PrivilegeCount: 1,
		Privileges: [1]LuidAndAttributes{{
			Luid:       luid,
			Attributes: SePrivilegeEnabled,
		}},
	}

	if err = syscall.AdjustTokenPrivileges(hToken, false, &tp, uint32(unsafe.Sizeof(tp)), nil, nil); err != nil {
		fmt.Printf("[-] AdjustTokenPrivileges failed: %v\n", err)
		return false
	}

	fmt.Println("[+] SeDebugPrivilege enabled.")
	return true
}

// Convert DOS path to NT path
func convertToNtPath(filePath string) (string, error) {
	buffer := make([]uint16, syscall.MAX_PATH)
	drive := filePath[:2]
	if err := syscall.QueryDosDevice(syscall.StringToUTF16Ptr(drive), &buffer[0], uint32(len(buffer))); err != nil {
		return "", fmt.Errorf("[-] QueryDosDevice failed: %v", err)
	}

	ntPath := syscall.UTF16ToString(buffer) + filePath[2:]
	return ntPath, nil
}

// Check if a file exists
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Main function
func main() {
	if !checkProcessIntegrityLevel() {
		fmt.Println("[-] Please run this program with elevated privileges.")
		return
	}

	if !enableSeDebugPrivilege() {
		fmt.Println("[-] Failed to enable SeDebugPrivilege.")
		return
	}

	filePath := "C:\\Windows\\System32\\notepad.exe"
	if !fileExists(filePath) {
		fmt.Println("[-] File does not exist.")
		return
	}

	ntPath, err := convertToNtPath(filePath)
	if err != nil {
		fmt.Printf("[-] Failed to convert to NT path: %v\n", err)
		return
	}

	fmt.Printf("[+] NT Path: %s\n", ntPath)
}
