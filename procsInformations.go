package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// ProcessInformation wrap basic process information and memory dump in a structure
type ProcessInformation struct {
	PID           uint32
	ProcessName   string
	ProcessPath   string
	ProcessMemory []byte
	MemoryHash    string
}

// ListProcess try to get all running processes and dump their memory, return a ProcessInformation slice
func ListProcess(verbose bool) (procsInfo []ProcessInformation) {
	runningPID := os.Getpid()

	procsIds, bytesReturned, err := GetProcessesList()
	if err != nil {
		log.Fatal(err)
	}
	for i := uint32(0); i < bytesReturned; i++ {
		if procsIds[i] != 0 && procsIds[i] != uint32(runningPID) {
			procHandle, err := GetProcessHandle(procsIds[i], windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ)
			if err != nil && verbose {
				log.Println("PID", procsIds[i], err)
			}

			if err == nil && procHandle > 0 {
				procFilename, modules, err := GetProcessModulesHandles(procHandle)
				if err == nil {
					for _, moduleHandle := range modules {
						if moduleHandle != 0 {
							moduleRawName, err := GetModuleFileNameEx(procHandle, moduleHandle, 512)
							if err != nil && verbose {
								log.Println(err)
							}
							moduleRawName = bytes.Trim(moduleRawName, "\x00")
							modulePath := strings.Split(string(moduleRawName), "\\")
							moduleFileName := modulePath[len(modulePath)-1]

							if procFilename == moduleFileName {
								memdump := DumpModuleMemory(procHandle, moduleHandle, verbose)
								if len(memdump) > 0 {
									proc := ProcessInformation{PID: procsIds[i], ProcessName: procFilename, ProcessPath: string(moduleRawName), ProcessMemory: memdump, MemoryHash: fmt.Sprintf("%x", md5.Sum(memdump))}
									procsInfo = append(procsInfo, proc)
								}
							}
						}
					}
				}
			}
			windows.CloseHandle(procHandle)
		}
	}
	return procsInfo
}

// KillProcessByID try to kill the specified PID
func KillProcessByID(procID uint32, verbose bool) (err error) {
	hProc, err := GetProcessHandle(procID, windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_TERMINATE)
	if err != nil && verbose {
		log.Println("PID", procID, err)
	}

	exitCode := GetExitCodeProcess(hProc)
	err = windows.TerminateProcess(hProc, exitCode)
	if err != nil {
		return err
	}

	return nil
}

// GetProcessesList return PID from running processes
func GetProcessesList() (procsIds []uint32, bytesReturned uint32, err error) {
	procsIds = make([]uint32, 2048)
	err = windows.EnumProcesses(procsIds, &bytesReturned)
	return procsIds, bytesReturned, err
}

// GetProcessHandle return the process handle from the specified PID
func GetProcessHandle(pid uint32, desiredAccess uint32) (handle windows.Handle, err error) {
	handle, err = windows.OpenProcess(desiredAccess, false, pid)
	return handle, err
}

// GetProcessModulesHandles list modules handles from a process handle
func GetProcessModulesHandles(procHandle windows.Handle) (processFilename string, modules []syscall.Handle, err error) {
	var processRawName []byte
	processRawName, err = GetProcessImageFileName(procHandle, 512)
	if err != nil {
		return "", nil, err
	}
	processRawName = bytes.Trim(processRawName, "\x00")
	processPath := strings.Split(string(processRawName), "\\")
	processFilename = processPath[len(processPath)-1]

	modules, err = EnumProcessModules(procHandle, 32)
	if err != nil {
		return "", nil, err
	}

	return processFilename, modules, nil
}

// DumpModuleMemory dump a process module memory and return it as a byte slice
func DumpModuleMemory(procHandle windows.Handle, modHandle syscall.Handle, verbose bool) []byte {
	moduleInfos, err := GetModuleInformation(procHandle, modHandle)
	if err != nil && verbose {
		log.Println(err)
	}

	memdump, err := ReadProcessMemory(procHandle, moduleInfos.BaseOfDll, uintptr(moduleInfos.SizeOfImage))
	if err != nil && verbose {
		log.Println(err)
	}

	memdump = bytes.Trim(memdump, "\x00")
	return memdump
}