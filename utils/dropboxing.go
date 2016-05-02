package utils

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"time"

	"fmt"
	"os"

	"strings"

	"github.com/stacktic/dropbox"
)

// Configuration from conf.json
type Configuration struct {
	Token string `json:"token"`
}

var db *dropbox.Dropbox
var dropboxLink *dropbox.Link
var err error

// GetDropboxFolder figures out the local Dropbox folder in the FS. Windows only at the moment
func GetDropboxFolder() (string, error) {

	var clientid, clientsecret, token string
	var dropboxFolder string
	// Read configuration from conf.json
	conffile, _ := os.Open("client_conf.json")
	decoder := json.NewDecoder(conffile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("config error:", err)
		return "", err
	}

	// Using placeholders since you just need token
	clientid = "placeholder"
	clientsecret = "placeholder"
	token = config.Token

	db = dropbox.NewDropbox()
	db.SetAppInfo(clientid, clientsecret)
	db.SetAccessToken(token)

	folder := "podcasts"
	if _, err = db.CreateFolder(folder); err != nil {
		if err.Error() != "Cannot create folder 'podcasts' because a file or folder already exists at path '/podcasts'" {
			fmt.Printf("Error creating folder %s: %s\n", folder, err)
		}
	} else {
		fmt.Printf("Folder %s successfully created\n", folder)
	}

	// Checking for the existence of (on Windows only obvs) %APPDATA%\Dropbox\host.db
	// If it exists then read it and base64 decode the second line of its contents to get the root path of the Dropbox files for this user
	// Then just do a local OS file copy to the right path instead of using db.UploadFile
	if _, err = os.Stat(os.Getenv("LOCALAPPDATA") + "\\Dropbox\\host.db"); err == nil {
		fmt.Println("Running in Windows Local Dropbox Mode")

		var dropboxDefFile *os.File
		dropboxDefFile, err = os.Open(os.Getenv("LOCALAPPDATA") + "\\Dropbox\\host.db")
		if err != nil {
			return "", err
		}
		defer dropboxDefFile.Close()

		scanner := bufio.NewScanner(dropboxDefFile)

		// Skip first line
		scanner.Scan()

		// Dropbox Store path on second line
		scanner.Scan()

		// Base64 decode the second line to get the path
		var sDec []byte
		sDec, err = b64.StdEncoding.DecodeString(scanner.Text())
		if err != nil {
			return "", err
		}
		dropboxFolder = string(sDec)
		//fmt.Println(dropboxFolder)
		return dropboxFolder, nil
	}
	// Only use this if running on Linux or OSX or somewhere where Dropbox is not actually installed
	// If running locally on Linux/OSX with Dropbox installed, the bandwidth usage is doubled since it uploads to Cloud and then Dropbox syncs back down. So double the bandwidth
	fmt.Println("Running in non-Windows Dropbox Mode")
	return "remote", nil
}

// CopyLocallyToDropbox copies file to local Dropbox FS - Windows only at the moment
func CopyLocallyToDropbox(srcFile string, destFolder string) error {

	s, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer s.Close()

	// Create the directory in the local Dropbox FS if it's not already there
	destPath := filepath.Dir(destFolder + srcFile)
	err = os.MkdirAll(destPath, 0777)
	if err != nil {
		return err
	}

	// Create the empty file on the local Dropbox FS
	d, err := os.Create(destFolder + srcFile)
	if err != nil {
		return err
	}

	// Copy the MP3 file to the local Dropbox FS
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	d.Close()
	return nil
}

// CopyRemotelyToDropbox uploads the file to Dropbox - Should work on Linux and OSX too but doesn't yet
func CopyRemotelyToDropbox(srcFile string, destPath string) error {
	var rev string

	// fmt.Printf("Running in Remote Dropbox Mode\n")

	if _, err = db.UploadFile(srcFile, destPath, true, rev); err != nil {
		fmt.Printf("Error uploading file: %s\n", err)
		return err
	}
	fmt.Printf("File successfully uploaded: %s\n", srcFile)
	return nil
}

// GetDropboxURLWhenSyncComplete will keep trying to getDropboxURL() until either
// we get a result from getDropboxURL() or the timeout expires
func GetDropboxURLWhenSyncComplete(destFile string) (string, error) {

	// 6 minutes seems a reasonable timeout for an MP3 to upload from Local to Remote
	// TODO: Make configurable
	timeout := time.After(6 * time.Minute)

	// Check every 20 seconds
	tick := time.Tick(10000 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return "", errors.New("timed out")
		// Got a tick, we should check on getDropboxURL()
		case <-tick:
			dropboxURL, err := GetDropboxURL(destFile)
			// Error from getDropboxURL(), we should bail
			if err == nil {
				return dropboxURL, err
			}
			// getDropboxURL() didn't work yet, but it didn't fail, so let's try again
			// this will exit up to the for loop
		}
	}
}

// GetDropboxURL retrieves the direct download URL for a file
func GetDropboxURL(destFile string) (string, error) {
	//fmt.Println(destFile)
	// Need to get Download URL of the Dropbox file so I can add to rss.xml
	if dropboxLink, err = db.Shares(destFile, false); err != nil {
		fmt.Printf("dropboxLink: %s\n", dropboxLink.URL)
		fmt.Printf("%s: %s\n", destFile, err)
		return "", err
	}
	s := strings.Split(dropboxLink.URL, "?")
	dlLink := s[0] + "?raw=1"
	//fmt.Printf("MP3 Direct download URL is: %s\n", dlLink)
	return dlLink, nil
}
