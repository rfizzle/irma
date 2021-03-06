package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hillu/go-yara/v4"
	golnk "github.com/parsiya/golnk"
)

// StartMenuAnalysisRoutine analyse system artefacts every 15 seconds
func StartMenuAnalysisRoutine(pQuarantine string, pKill bool, pNotifications bool, pVerbose bool, pInfiniteLoop bool, rules *yara.Rules) {
	for {
		lnk, errors := ListStartMenuLnkPersistence(pVerbose)
		if errors != nil && pVerbose {
			for _, err := range errors {
				logMessage(LOG_ERROR, "[ERROR]", err)
			}
		}

		for _, l := range lnk {
			paths := FormatPathFromComplexString(l)
			for _, p := range paths {
				FileAnalysis(p, pQuarantine, pKill, pNotifications, pVerbose, rules, "STARTMENU")
			}
		}

		if !pInfiniteLoop {
			wg.Done()
			break
		} else {
			time.Sleep(15 * time.Second)
		}
	}
}

// ListStartMenuFolders return a []string of all available StartMenu folders
func ListStartMenuFolders(verbose bool) (startMenu []string, err error) {
	var usersDir []string

	startMenu = append(startMenu, os.Getenv("SystemDrive")+`\ProgramData\Microsoft\Windows\Start Menu\Programs\StartUp`)

	usersDir, err = RetrivesFilesFromUserPath(os.Getenv("SystemDrive")+"\\Users", false, nil, false, verbose)
	if err != nil {
		return startMenu, err
	}

	for _, uDir := range usersDir {
		if !strings.HasSuffix(uDir, "Public") && !strings.HasSuffix(uDir, "Default") {
			startMenu = append(startMenu, uDir+`\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup`)
		}
	}

	return startMenu, err
}

// ListStartMenuLnkPersistence check for every *.lnk in Windows StartMenu folders and extract every executable path in those links
func ListStartMenuLnkPersistence(verbose bool) (exePath []string, errors []error) {

	startMenuFolders, err := ListStartMenuFolders(verbose)
	if err != nil {
		errors = append(errors, err)
	}

	for _, path := range startMenuFolders {

		files, err := RetrivesFilesFromUserPath(path, true, []string{".lnk"}, false, verbose)

		if err != nil {
			errors = append(errors, fmt.Errorf("%s - %s", path, err.Error()))
		}

		for _, p := range files {
			lnk, lnkErr := golnk.File(p)
			if lnkErr != nil {
				errors = append(errors, fmt.Errorf("%s - Lnk parse error", p))
				continue
			}

			exePath = append(exePath, lnk.LinkInfo.LocalBasePath)
		}
	}

	return exePath, errors
}
