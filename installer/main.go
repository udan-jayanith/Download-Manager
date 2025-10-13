package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	fmt.Println(
		`
		Download Manage installer
		`)

	fmt.Println("Do you want to install the Download-Manager if yes press 'Y' else 'N' ?")
	fmt.Print(": ")
	var userInput string
	_, err := fmt.Scanln(&userInput)
	if err != nil && err.Error() != "unexpected newline" {
		ErrorPrinter(err)
	} else if err != nil && err.Error() == "unexpected newline" {
		userInput = "y"
	}

	userInput = strings.ToLower(userInput)
	if userInput != "y" && userInput != "yes" {
		return
	}

	fmt.Println("\t * Installing")
	Install()
	fmt.Scanln()
}

func ErrorPrinter(err error) {
	fmt.Println("\tError occurred")
	fmt.Println("\t", err.Error())
}

func Install() {
	info, err := os.Stat("./build")
	if err != nil {
		ErrorPrinter(fmt.Errorf(err.Error(), "Build directory not found."))
		return
	} else if !info.IsDir() {
		ErrorPrinter(fmt.Errorf("Build directory not found."))
		return
	}

	buildFs := os.DirFS("./build")
	installFolder := "C:/Program Files/Download-Manager/"

	err = os.RemoveAll(installFolder)
	if err != nil {
		ErrorPrinter(fmt.Errorf(err.Error(), "Removal error"))
		return
	}

	err = os.CopyFS(installFolder, buildFs)
	if err != nil {
		ErrorPrinter(fmt.Errorf(err.Error(), "Copying error"))
		return
	}

	downloadManagerExe := filepath.Join(installFolder, "Download-Manager.exe")
	downloadManagerExe = filepath.Clean(downloadManagerExe)
	if _, err := os.Stat(downloadManagerExe); err != nil {
		ErrorPrinter(fmt.Errorf(downloadManagerExe, "not found."))
		return
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, "SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run", registry.ALL_ACCESS)
	if err != nil {
		ErrorPrinter(fmt.Errorf(err.Error(), "Key opening error."))
		return
	}
	defer key.Close()

	name := "Download-Manager"
	_, _, err = key.GetStringValue(name)
	if err != nil && err != registry.ErrNotExist {
		ErrorPrinter(fmt.Errorf(err.Error(), "Key retrieving error."))
		return
	} else if err == nil {
		if err := key.DeleteValue(name); err != nil {
			ErrorPrinter(fmt.Errorf(err.Error(), "Key deletion error"))
			return
		}
	}

	if err := key.SetStringValue(name, fmt.Sprintf(`"%s" -background`, downloadManagerExe)); err != nil {
		ErrorPrinter(fmt.Errorf(err.Error(), "Key setting error"))
		return
	}

	fmt.Println("\t * Done")
	fmt.Println("\t * Please restart this device")
}
