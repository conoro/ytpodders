package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/conoro/ytpodders/commands"
	"github.com/gorilla/feeds"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

/*
sqlite> .schema subscriptions
CREATE TABLE subscriptions(
ID INTEGER PRIMARY KEY,
suburl CHAR(1024)
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

// INSERT INTO subscriptions(suburl) VALUES ("https://www.youtube.com/user/durianriders");
// INSERT INTO subscriptions(suburl) VALUES ("https://www.youtube.com/channel/UCYdkEm-NjhS8TmLVt_qZy9g");
// INSERT INTO subscriptions(suburl) VALUES ("https://www.youtube.com/user/sciguy14");
// INSERT INTO subscriptions(suburl) VALUES ("https://www.youtube.com/channel/UCh8rjWtGCIAbwPrZb3Te8bQ");
// INSERT INTO subscriptions(suburl) VALUES ("https://www.youtube.com/channel/UCSUi7O_Fg6SgwkF2W2Au2Zw");

// YTSubscription is just the URL of each YouTuber you are subscribed to
type YTSubscription struct {
	SubID  int64  `db:"ID"`
	SubURL string `db:"suburl"`
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

var RSSXML = &feeds.Feed{
	Title:       "YTPodders YouTube Podcasts",
	Link:        &feeds.Link{Href: "https://ytpodders.com"},
	Description: "YouTube Videos converted to Podcasts by YTPodders",
	Author:      &feeds.Author{Name: "YTPodder", Email: "youtuber@example.com"},
}

func main() {
	var dropboxFolder string
	var fileSize int64

	// TODO: Add proper Go-style logging everywhere instead of all of these Print statements
	// TODO: Figure out Windows Scheduler again
	// TODO: Add proper Set max retention time as parameter in conf.json

	// TODO: Offer a range of commands as follows:
	// TODO: without params - print help out and exit
	// TODO: help - print help out and exit
	// TODO: run - updates everything as you'd expect. Normal one-off execution
	// TODO: add - add a subscription. Pass it the URL of a YouTube Channel or User (possibly sanitize)
	commands.AddFeed("test URL")

	// TODO: list - list all subscriptions as ID, URL, (Title maybe? or Uploader maybe?)
	// TODO: remove - remove a subscription by ID
	// TODO: scheduler - runs it as some sort of background daemon. No idea how to to this on Windows
	// TODO: dryrun - same as run except nothing is downloaded and the database is not modified but it lists what it would do
	// TODO: prune - pass a number of days as param. Mark entries in DB as "expired". Delete mp3 files locally and from Dropbox. Do not re-download on next run!
	// TODO: reauth - re-run the Auth flow to get a new Dropbox token
	// TODO: reinit - Clear the DB completely and delete both local and Dropbox MP3s

	dropboxFolder, err = getDropboxFolder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	vidcmd := "youtube-dl"
	//vidcmd := "echo"

	db, err := sqlx.Connect("sqlite3", "ytpodders.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// query
	ytSubscriptions := []YTSubscription{}
	err = db.Select(&ytSubscriptions, "SELECT ID, suburl FROM subscriptions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, subscription := range ytSubscriptions {

		ytSubscriptionEntries := []YTSubscriptionEntry{}
		err = db.Select(&ytSubscriptionEntries, "SELECT subscription, url, title, date FROM subscription_entries WHERE subscription=$1", subscription.SubID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(subscription)
		fmt.Println(ytSubscriptionEntries)

		var feedURL string

		split := strings.Split(subscription.SubURL, "/")
		feedid := split[len(split)-1]
		fmt.Println(feedid)

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

		for _, item := range feed.Items {
			fmt.Println(item.Title)
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
				fmt.Println(mp3FileLocalStyle)
				fmt.Println(mp3FileRemoteStyle)

				fileSize, _ = getFileSize(mp3FileLocalStyle)

				if dropboxFolder != "remote" {
					err = copyLocallyToDropbox(mp3FileLocalStyle, dropboxFolder+"\\Apps\\YTPodders\\")
					if err != nil {
						fmt.Fprintf(os.Stderr, "Copy to Local Dropbox Error: %v\n", err)
						os.Exit(1)
					}
				} else {
					err = copyRemotelyToDropbox("."+mp3FileRemoteStyle, mp3FileRemoteStyle)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %v\n", err)
						os.Exit(1)
					}
				}
				var dropboxURL string
				dropboxURL, err = getDropboxURLWhenSyncComplete(mp3FileRemoteStyle)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(dropboxURL)
				fmt.Println("Adding new RSS Entry")
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO subscription_entries(subscription,url,title,date, dropboxurl, filesize) VALUES($1,$2,$3,$4,$5,$6)", subscription.SubID, item.Link, item.Title, item.Date, dropboxURL, fileSize)
				tx.Commit()

			}
		}

		// Add all entries to RSSXML struct which will be used to generate the rss.xml file
		ytAllSubscriptionEntries := []YTSubscriptionEntry{}
		err = db.Select(&ytAllSubscriptionEntries, "SELECT subscription, url, title, date, dropboxurl, filesize FROM subscription_entries")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		for _, ytItem := range ytAllSubscriptionEntries {
			addEntrytoRSSXML(ytItem)
		}
	}

	// Create and update rss.xml
	RSSFile, err := generateRSS(dropboxFolder + "\\Apps\\YTPodders\\")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(RSSFile)

	// When Dropbox has synced, return the URL of rss.xml to the User
	RSSFileURL, err := getDropboxURLWhenSyncComplete("rss.xml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(RSSFileURL)

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

	Item := feeds.Item{
		Title:       ytItem.Title,
		Link:        &feeds.Link{Href: ytItem.DropboxURL, Length: strconv.FormatInt(ytItem.FileSize, 10), Type: "audio/mpeg"},
		Description: ytItem.Title,
		Author:      &feeds.Author{Name: "YTPodder", Email: "youtuber@example.com"},
		Created:     timeStamp,
	}

	RSSXML.Add(&Item)
	return nil

}

func generateRSS(dropboxFolder string) (string, error) {
	now := time.Now()
	RSSXML.Created = now

	rss, err := RSSXML.ToRss()
	if err != nil {
		return "", err
	}

	rssOut := []byte(rss)

	// TODO: Need to add remote upload of this file when not running on local PC
	rssFile := dropboxFolder + "rss.xml"
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
	fmt.Printf("The file is %d bytes long", fi.Size())
	return fi.Size(), nil
}
