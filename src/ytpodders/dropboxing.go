package main

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"

	"fmt"
	"os"

	"strings"

	"github.com/stacktic/dropbox"
)

// Configuration from conf.json
type Configuration struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"clientsecret"`
	Token        string `json:"token"`
}

var db *dropbox.Dropbox
var dropboxLink *dropbox.Link
var err error

func getDropboxFolder() (string, error) {

	var clientid, clientsecret, token string
	var dropboxFolder string
	// Read configuration from conf.json
	conffile, _ := os.Open("conf.json")
	decoder := json.NewDecoder(conffile)
	config := Configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("config error:", err)
		return "", err
	}

	clientid = config.ClientID
	clientsecret = config.ClientSecret

	//token = config.Token

	// TODO: Check if Token defined in conf.json. If no, do longer OAuth flow and then save token to conf.json

	// 1. Create a new dropbox object.
	db = dropbox.NewDropbox()

	// 2. Provide your clientid and clientsecret (see prerequisite).
	db.SetAppInfo(clientid, clientsecret)

	// This method will ask the user to visit an URL and paste the generated code.
	if err = db.Auth(); err != nil {
		fmt.Println(err)
		return "", err
	}
	// You can now retrieve the token if you want.
	token = db.AccessToken()
	fmt.Println(token)

	// 3. Provide the user token.
	db.SetAccessToken(token)

	// 4. Send your commands.
	// In this example, you will create a new folder named "demo".
	folder := "podcasts"
	if _, err = db.CreateFolder(folder); err != nil {
		fmt.Printf("Error creating folder %s: %s\n", folder, err)
	} else {
		fmt.Printf("Folder %s successfully created\n", folder)
	}

	// Only use this if running on a Server, not your local PC. Because it uploads to Cloud and then Dropbox syncs back down. So double the bandwidth
	// Do this by checking for the existence of (on Windows only obvs) %APPDATA%\Dropbox\host.db
	// If it exists then read it and base64 decode the second line of its contents to get the root path of the Dropbox files for this user
	// Then just do a local OS file copy to the right path instead of using db.UploadFile
	if _, err = os.Stat(os.Getenv("LOCALAPPDATA") + "\\Dropbox\\host.db"); err == nil {
		fmt.Printf("It's local baby!\n")

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
		fmt.Println(dropboxFolder)
		return dropboxFolder, nil
	}
	return "remote", nil

}

// TODO: Add logging everywhere
// TODO: Move all of the code below to working with a list of files
// TODO: Should probably run the code inside a Dropbox folder when running locally to avoid the file copying (or have CMD file in Dropbox Dir)
// TODO: Need to figure out Hierarchy (possibly just one directory per Sub)
// TODO: Need to figure out clean naming for Windows filesystem e.g. remove underscores, hashes, colons, slashes etc (or escape them)
// TODO: Figure out Windows Scheduler again

// TODO: Need to add the filelength to the RSS file

func copyLocallyToDropbox(srcFile string, destFolder string) error {

	s, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer s.Close()
	//destFileSplit := strings.Split(destFolder+srcFile, "\\")
	//destFile := destFileSplit[len(destFileSplit)-1]
	//err = os.MkdirAll(destFile, 0777)
	//if err != nil {
	//	return err
	//}

	d, err := os.Create(destFolder + srcFile)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	d.Close()
	return nil
}

func copyRemotelyToDropbox(srcFile string, destPath string) error {
	var rev string

	fmt.Printf("It's remote baby!\n")

	if _, err = db.UploadFile(srcFile, destPath, true, rev); err != nil {
		fmt.Printf("Error uploading file: %s\n", err)
		return err
	}
	fmt.Printf("File successfully uploaded\n")
	return nil
}

// getDropboxURLwhenSyncComplete will keep trying to getDropboxURL() until either
// we get a result from getDropboxURL() or the timeout expires
func getDropboxURLWhenSyncComplete(destFile string) (string, error) {

	// 2 minutes seems a reasonable timeout for an MP3 to upload from Local to Remote
	timeout := time.After(2 * time.Minute)
	tick := time.Tick(10000 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return "", errors.New("timed out")
		// Got a tick, we should check on getDropboxURL()
		case <-tick:
			dropboxURL, err := getDropboxURL(destFile)
			// Error from getDropboxURL(), we should bail
			if err == nil {
				return dropboxURL, err
			}
			// getDropboxURL() didn't work yet, but it didn't fail, so let's try again
			// this will exit up to the for loop
		}
	}
}

func getDropboxURL(destFile string) (string, error) {
	// Need to get Download URL of the Dropbox file so I can add to rss.xml
	if dropboxLink, err = db.Shares(destFile, false); err != nil {
		fmt.Printf("%s: %s\n", destFile, err)
		return "", err
	}
	s := strings.Split(dropboxLink.URL, "?")
	dlLink := s[0] + "?raw=1"
	fmt.Printf("MP3 Direct download URL is: %s\n", dlLink)
	return dlLink, nil
}
