package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/SlyMarbo/rss"
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

func main() {

	vidcmd := "youtube-dl"
	//vidcmd := "echo"

	db, err := sqlx.Connect("sqlite3", "youtube2podcast.db")
	checkErr(err)

	// query
	ytSubscriptions := []YTSubscription{}
	err = db.Select(&ytSubscriptions, "SELECT ID, suburl FROM subscriptions")
	checkErr(err)
	//fmt.Println(ytSubscriptions)

	for _, subscription := range ytSubscriptions {

		ytSubscriptionEntries := []YTSubscriptionEntry{}
		err = db.Select(&ytSubscriptionEntries, "SELECT subscription, url, title, date FROM subscription_entries WHERE subscription=$1", subscription.SubID)
		checkErr(err)
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
			panic(error)
		}
		fmt.Println(feed)

		for _, item := range feed.Items {
			fmt.Println(item.Title)
			if RSSEntryInDB(item.Link, ytSubscriptionEntries) == false {
				fmt.Println("Adding new RSS Entry")
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO subscription_entries(subscription,url,title,date) VALUES($1,$2,$3,$4)", subscription.SubID, item.Link, item.Title, item.Date)
				tx.Commit()

				args1 := []string{"--extract-audio", "--audio-format", "mp3", "-o", "./podcasts/%(uploader)s/%(title)s.%(ext)s", item.Link}
				cmd := exec.Command(vidcmd, args1...)
				err5 := cmd.Run()
				if err5 != nil {
					fmt.Println(err5)
					os.Exit(1)
				}
			}
		}

	}

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

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Subscriptions
// YouTube Channel ID or User ID
// Create RSS URL from https://www.youtube.com/feeds/videos.xml?user=specificuserid or https://www.youtube.com/feeds/videos.xml?channel_id=specificchannelid

// vid, err := ytdl.GetVideoInfo("https://www.youtube.com/watch?v=1rZ-JorHJEY")
// The video ID
// ID  string `json:"id"`
// Title string `json:"title"`
// Description string `json:"description"`
// DatePublished time.Time `json:"datePublished"`
// Author string `json:"author"`

// Subscription Entries
// Subscription ID
// Entry URL
// Entry URL date
// Entry URL title
// Entry URl GUID?

// Podcast Items
// Title
// GUID = Original YouTube Link
// MP3 URL on Server (or S3)
// Description
// Enclosure Same as GUID
// Category Maybe
// pubDate
