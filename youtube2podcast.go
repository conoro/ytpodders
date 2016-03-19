package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/SlyMarbo/rss"
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
}

// YTdl has info about the YouTube file itself from youtube-dl
type YTdl struct {
	Uploader string `json:"uploader"`
	Title    string `json:"title"`
}

func main() {
	var youtubeInfo YTdl
	var dropboxFolder string
	dropboxFolder, err = getDropboxFolder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	vidcmd := "youtube-dl"
	//vidcmd := "echo"

	db, err := sqlx.Connect("sqlite3", "youtube2podcast.db")
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

	//fmt.Println(ytSubscriptions)

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
		fmt.Println(feed)

		for _, item := range feed.Items {
			fmt.Println(item.Title)
			if RSSEntryInDB(item.Link, ytSubscriptionEntries) == false {
				fmt.Println("Adding new RSS Entry")
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO subscription_entries(subscription,url,title,date) VALUES($1,$2,$3,$4)", subscription.SubID, item.Link, item.Title, item.Date)
				tx.Commit()

				//TODO: Need to figure out what filename youtube-dl is generating here
				args1 := []string{"-v", "--extract-audio", "--audio-format", "mp3", "-o", "./podcasts/%(uploader)s/%(title)s.%(ext)s", item.Link}
				cmd := exec.Command(vidcmd, args1...)

				var out bytes.Buffer
				cmd.Stdout = &out
				err = cmd.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}

				//err = json.Unmarshal([]byte(out.String()), &youtubeInfo)
				//if err != nil {
				//	fmt.Fprintf(os.Stderr, "error: %v\n", err)
				//	os.Exit(1)
				//}

				scanner := bufio.NewScanner(&out)

				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "[ffmpeg] Destination:") {
						fmt.Println(scanner.Text())
					}
				}

				mp3FileLocalStyle := "\\podcasts\\" + youtubeInfo.Uploader + "\\" + youtubeInfo.Title + ".mp3"
				mp3FileRemoteStyle := "/podcasts/" + youtubeInfo.Uploader + "/" + youtubeInfo.Title + ".mp3"
				fmt.Println(mp3FileLocalStyle)
				fmt.Println(mp3FileRemoteStyle)

				getFileSize(mp3FileLocalStyle)

				if dropboxFolder != "remote" {
					err = copyLocallyToDropbox(mp3FileLocalStyle, dropboxFolder+"\\Apps\\YouTube2Podcast\\podcasts\\"+youtubeInfo.Uploader+"\\")
					if err != nil {
						fmt.Fprintf(os.Stderr, "error: %v\n", err)
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
				dropboxURL, err = getDropboxURL(mp3FileRemoteStyle)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
				fmt.Println(dropboxURL)

			}
		}

	}

	// TODO: Need to create and update rss.xml
	RSSFile, err := generateRSS(dropboxFolder + "\\Apps\\YouTube2Podcast\\")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(RSSFile)

	RSSFileURL, err := getDropboxURL("rss.xml")
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

func generateRSS(dropboxFolder string) (string, error) {
	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Conor's YouTube Podcasts",
		Link:        &feeds.Link{Href: "http://conoroneill.net"},
		Description: "YouTube Videos converted to Podcasts",
		Author:      &feeds.Author{Name: "Conor O'Neill", Email: "conor@conoroneill.com"},
		Created:     now,
	}

	feed.Items = []*feeds.Item{
		{
			Title:       "GINGER RUNNER LIVE 107b The Joe Grant Episode",
			Link:        &feeds.Link{Href: "https://www.dropbox.com/s/d5vei33k0xcdndm/GINGER%20RUNNER%20LIVE%20%23107%20_%20The%20Joe%20Grant%20Episode-pYHbkSNAR7I.mp3?raw=1", Length: "83503158", Type: "audio/mpeg"},
			Description: "GINGER RUNNER LIVE 107b The Joe Grant Episode",
			Author:      &feeds.Author{Name: "Ginger Runner", Email: "conor@conoroneill.com"},
			Created:     now,
		},
	}

	rss, err := feed.ToRss()
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
