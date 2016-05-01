// TODO Main functionality doesn't work on OSX. Probably errors in the non-local flow. Try to add local flow for OSX and Linux too

package commands

import (
	"fmt"

	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/conoro/ytpodders/utils"

	"github.com/SlyMarbo/rss"
	"github.com/conoro/feeds"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Any way to avoid blank import?
	"github.com/spf13/cobra"
)

/*
sqlite> .schema subscriptions
CREATE TABLE subscriptions(
ID INTEGER PRIMARY KEY,
suburl CHAR(1024),
subtitle CHAR(1024),
substatus CHAR(64)
);
sqlite> .schema subscription_entries
CREATE TABLE subscription_entries(
subscription INTEGER,
url TEXT,
title TEXT,
date TEXT,
dropboxurl TEXT,
filesize INTEGER,
FOREIGN KEY(subscription) REFERENCES subscriptions(ID)
);

*/

// INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES ("https://www.youtube.com/user/durianriders", "durianrider", "enabled");
// INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES ("https://www.youtube.com/channel/UCYdkEm-NjhS8TmLVt_qZy9g", "Making Stuff", "enabled");
// INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES ("https://www.youtube.com/user/sciguy14", "Jeremy Blum", "enabled");
// INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES ("https://www.youtube.com/channel/UCh8rjWtGCIAbwPrZb3Te8bQ", "GEARIST", "enabled");
// INSERT INTO subscriptions(suburl, subtitle, substatus) VALUES ("https://www.youtube.com/channel/UCSUi7O_Fg6SgwkF2W2Au2Zw", "Anthony Ngu", "enabled");

// YTSubscription is just the URL of each YouTuber you are subscribed to
type YTSubscription struct {
	SubID     int64  `db:"ID"`
	SubURL    string `db:"suburl"`
	SubTitle  string `db:"subtitle"`
	SubStatus string `db:"substatus"`
}

// YTSubscriptionEntry has info about each of the parsed and downloaded "episodes" from YouTube
type YTSubscriptionEntry struct {
	Subscription int64  `db:"subscription"`
	URL          string `db:"url"`
	Title        string `db:"title"`
	Date         string `db:"date"`
	DropboxURL   string `db:"dropboxurl"`
	FileSize     int64  `db:"filesize"`
}

// RSSXML is used to build the rss.xml file which you subscribe to in your podcasting app
var RSSXML = &feeds.Feed{
	Title:       "YTPodders YouTube Podcasts",
	Link:        &feeds.Link{Href: "https://ytpodders.com/"},
	Description: "YouTube Videos converted to Podcasts by YTPodders",
	Author:      &feeds.Author{Name: "YTPodder", Email: "youtuber@example.com"},
}

// RootCmd is the Action to run if no command specified. In our case this is a full update of all the feeds and podcasts
var RootCmd = &cobra.Command{
	Use:   "ytpodders",
	Short: "YTPodders creates subscribable MP3 podcasts from YouTube Users and Channels using Dropbox",
	Long:  `Each time you run YTPodders, it checks the list of YouTube Users and Channels that you have added here for new uploads. It grabs those using youtube-dl and converts them to MP3s. It then copies or uploads the MP3s to your Dropbox account. Finally it updates rss.xml and provides you with its URL. You can add this URL to your podcast app on your phone e.g. BeyondPod on Android and then automatically get the audio of those YouTubers on your phone when your podcast app updates.`,
	Run:   RootRun,
}

// RootRun is executed when user passes no arguments to ytpodders
func RootRun(cmd *cobra.Command, args []string) {
	var dropboxFolder string
	var fileSize int64
	var err error

	dropboxFolder, err = utils.GetDropboxFolder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	vidcmd := "youtube-dl"

	db, err := sqlx.Connect("sqlite3", "ytpodders.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// query
	ytSubscriptions := []YTSubscription{}
	err = db.Select(&ytSubscriptions, "SELECT DISTINCT ID, suburl, subtitle, substatus FROM subscriptions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, subscription := range ytSubscriptions {

		ytSubscriptionEntries := []YTSubscriptionEntry{}
		err = db.Select(&ytSubscriptionEntries, "SELECT DISTINCT subscription, url, title, date FROM subscription_entries WHERE subscription=$1", subscription.SubID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		// fmt.Println(subscription)
		//fmt.Println(ytSubscriptionEntries)

		var feedURL string

		split := strings.Split(subscription.SubURL, "/")
		feedid := split[len(split)-1]
		// fmt.Println(feedid)

		if strings.Contains(subscription.SubURL, "channel") {
			feedURL = "https://www.youtube.com/feeds/videos.xml?channel_id=" + feedid
		} else {
			feedURL = "https://www.youtube.com/feeds/videos.xml?user=" + feedid
		}
		feed, error := rss.Fetch(feedURL)
		if error != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		// TODO: Make limit configurable and handle situation where limit > range
		feedSlice := feed.Items[:5]
		for _, item := range feedSlice {
			//fmt.Println(item.Title)
			if RSSEntryInDB(item.Link, ytSubscriptionEntries) == false {

				args1 := []string{"-v", "--extract-audio", "--audio-format", "mp3", "-o", "./podcasts/%(uploader)s/%(title)s.%(ext)s", item.Link}
				cmd := exec.Command(vidcmd, args1...)

				var out bytes.Buffer
				cmd.Stdout = &out
				err = cmd.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}

				var ytdlPath string
				scanner := bufio.NewScanner(&out)
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "[ffmpeg] Destination:") {
						ytdlPath = strings.Split(scanner.Text(), ": ")[1]
					}
				}

				mp3FileLocalStyle := ytdlPath
				mp3FileRemoteStyle := "/" + strings.Replace(ytdlPath, "\\", "/", -1)
				// fmt.Println(mp3FileLocalStyle)
				// fmt.Println(mp3FileRemoteStyle)

				fileSize, _ = getFileSize(mp3FileLocalStyle)

				if dropboxFolder != "remote" {
					// Running locally on Windows with Dropbox installed
					err = utils.CopyLocallyToDropbox(mp3FileLocalStyle, dropboxFolder+"\\Apps\\YTPodders\\")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Copy to Local Dropbox Error: %v\n", err)
						os.Exit(1)
					}
				} else {
					// Running on OSX or Linux or somewhere where Dropbox is not installed
					err = utils.CopyRemotelyToDropbox("."+mp3FileRemoteStyle, mp3FileRemoteStyle)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %v\n", err)
						os.Exit(1)
					}
				}
				var dropboxURL string
				dropboxURL, err = utils.GetDropboxURLWhenSyncComplete(mp3FileRemoteStyle)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
				// fmt.Println(dropboxURL)

				//TODO - Don't add DB entry until I'm 100% sure the whole end-to-end flow has worked for that entry including Dropbox Sync
				// TODO: Seem to be getting duplicates in RSS file but not the DB. Why?
				fmt.Printf("Adding new RSS Entry %s \n", item.Title)
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO subscription_entries(subscription,url,title,date, dropboxurl, filesize) VALUES($1,$2,$3,$4,$5,$6)", subscription.SubID, item.Link, item.Title, item.Date, dropboxURL, fileSize)
				tx.Commit()

			}
		}

	}

	// Add all entries to RSSXML struct which will be used to generate the rss.xml file
	ytAllSubscriptionEntries := []YTSubscriptionEntry{}
	err = db.Select(&ytAllSubscriptionEntries, "SELECT DISTINCT subscription, url, title, date, dropboxurl, filesize FROM subscription_entries ORDER BY date DESC")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for _, ytItem := range ytAllSubscriptionEntries {
		addEntrytoRSSXML(ytItem)
		//fmt.Println(ytItem)
	}

	// Create and update rss.xml
	_, err = generateRSS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Copy rss.xml to Dropbox
	if dropboxFolder != "remote" {
		// Running locally on Windows with Dropbox installed
		err = utils.CopyLocallyToDropbox("rss.xml", dropboxFolder+"\\Apps\\YTPodders\\")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Copy to Local Dropbox Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Running on OSX or Linux or somewhere where Dropbox is not installed
		err = utils.CopyRemotelyToDropbox("./rss.xml", "/rss.xml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// When Dropbox has synced, return the URL of rss.xml to the User
	RSSFileURL, err := utils.GetDropboxURLWhenSyncComplete("rss.xml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n\nSubscribe to this RSS URL in your Podcasting App: %s\n\n", RSSFileURL)

}

// RSSEntryInDB checks if we have already downloaded this "episode"
func RSSEntryInDB(link string, dbentries []YTSubscriptionEntry) bool {
	for _, b := range dbentries {
		if b.URL == link {
			return true
		}
	}
	return false
}

func addEntrytoRSSXML(ytItem YTSubscriptionEntry) error {

	layOut := "2006-01-02 15:04:05-07:00"
	timeStamp, err := time.Parse(layOut, ytItem.Date)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//TODO Add Icon to feed
	//TODO Either fix the RSS feed or change the Atom file to atom.xml
	Item := feeds.Item{
		Title:       ytItem.Title,
		Link:        &feeds.Link{Href: ytItem.DropboxURL, Length: strconv.FormatInt(ytItem.FileSize, 10), Type: "audio/mpeg"},
		Description: ytItem.Title,
		Author:      &feeds.Author{Name: "YTPodder@example.com (YTPodder)", Email: "YTPodder@example.com"},
		Created:     timeStamp,
	}

	RSSXML.Add(&Item)
	return nil

}

// Create rss.xml
func generateRSS() (string, error) {
	now := time.Now()
	RSSXML.Created = now

	rss, err := RSSXML.ToAtom()
	//rss, err := RSSXML.ToRss()
	if err != nil {
		return "", err
	}

	rssOut := []byte(rss)

	rssFile := "rss.xml"
	err = ioutil.WriteFile(rssFile, rssOut, 0644)
	if err != nil {
		return "", err
	}
	return rssFile, nil
}

func getFileSize(srcFile string) (int64, error) {
	fi, err := os.Stat(srcFile)
	if err != nil {
		return 0, err
	}
	// fmt.Printf("The file is %d bytes long", fi.Size())
	return fi.Size(), nil

}
