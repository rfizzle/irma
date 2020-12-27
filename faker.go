package main

import (
	"encoding/base64"
	"io/ioutil"
	"syscall"
)


// SpawnFakeProcess drop an useless process and execute it
func SpawnFakeProcess(processName string) (err error) {
	var sDec []byte
	sDec, err = base64.StdEncoding.DecodeString(faker)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(processName, sDec, 0644)
	if err != nil {
		return err
	}

	var sI syscall.StartupInfo
	var pI syscall.ProcessInformation
	argv := syscall.StringToUTF16Ptr(processName)
	err = syscall.CreateProcess(nil, argv, nil, nil, true, 0, nil, nil, &sI, &pI)
	if err != nil {
		return err
	}

	return nil
}