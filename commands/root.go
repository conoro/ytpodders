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
	"github.com/asdine/storm"
	"github.com/conoro/feeds"
	"github.com/spf13/cobra"
)

// YTSubscription is just the URL of each YouTuber you are subscribed to
type YTSubscription struct {
	ID        int
	SubURL    string `storm:"unique"`
	SubTitle  string
	SubStatus string
}

// YTSubscriptionEntry has info about each of the parsed and downloaded "episodes" from YouTube
type YTSubscriptionEntry struct {
	ID           int
	Subscription int    `storm:"index"`
	URL          string `storm:"unique"`
	Title        string
	Date         time.Time `storm:"index"`
	DropboxURL   string
	FileSize     int64
}

// RSSXML is used to build the rss.xml file which you subscribe to in your podcasting app
var RSSXML = &feeds.Feed{
	Title:       "YTPodders YouTube Podcasts",
	Link:        &feeds.Link{Href: "https://ytpodders.com/"},
	Description: "YouTube Videos converted to Podcasts by YTPodders",
	Author:      &feeds.Author{Name: "YTPodder", Email: "youtuber@example.com"},
}

const remote string = "remote"

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
		fmt.Fprintf(os.Stderr, "error12: %v\n", err)
		os.Exit(1)
	}

	vidcmd := "youtube-dl"

	// Deal with OSX and Linux not finding ffmpeg/ffprobe in current directory
	if dropboxFolder == remote {
		vidcmd = "./youtube-dl.sh"
	}

	db, err := storm.Open("ytpodders.boltdb", storm.AutoIncrement())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error1: %v\n", err)
	}
	defer db.Close()

	// query
	var ytSubscriptions []YTSubscription
	err = db.All(&ytSubscriptions)
	if err != nil && err.Error() != "bucket YTSubscription not found" && strings.Contains(err.Error(), "YTSubscriptionEntry") == false {
		fmt.Fprintf(os.Stderr, "error2: %v\n", err)
	}

	for _, subscription := range ytSubscriptions {

		var ytSubscriptionEntries []YTSubscriptionEntry
		err = db.Find("Subscription", subscription.ID, &ytSubscriptionEntries)
		if err != nil && err.Error() != "bucket YTSubscription not found" && strings.Contains(err.Error(), "YTSubscriptionEntry") == false {
			fmt.Fprintf(os.Stderr, "error3: %v\n", err)
		}
		//fmt.Println(subscription)
		//fmt.Println(ytSubscriptionEntries)

		var feedURL string

		split := strings.Split(subscription.SubURL, "/")
		feedid := split[len(split)-1]
		//fmt.Println(feedid)

		if strings.Contains(subscription.SubURL, "channel") {
			feedURL = "https://www.youtube.com/feeds/videos.xml?channel_id=" + feedid
		} else {
			feedURL = "https://www.youtube.com/feeds/videos.xml?user=" + feedid
		}
		feed, err := rss.Fetch(feedURL)
		if err != nil && err.Error() != "bucket YTSubscription not found" && strings.Contains(err.Error(), "YTSubscriptionEntry") == false {
			fmt.Fprintf(os.Stderr, "error11: %v\n", err)
			os.Exit(1)
		}

		// TODO: Make limit configurable and handle situation where limit > range
		feedSlice := feed.Items[:5]
		for _, item := range feedSlice {
			fmt.Println(item.Title)
			if RSSEntryInDB(item.Link, ytSubscriptionEntries) == false {

				args1 := []string{"-v", "--newline", "--restrict-filenames", "--extract-audio", "--audio-format", "mp3", "-o", "./podcasts/%(uploader)s/%(title)s.%(ext)s", item.Link}
				cmd := exec.Command(vidcmd, args1...)

				//fmt.Println(args1)

				var out bytes.Buffer
				cmd.Stdout = &out
				err = cmd.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "error10: %v\n", err)
					os.Exit(1)
				}

				defer cmd.Wait()

				var ytdlPath string
				scanner := bufio.NewScanner(&out)
				for scanner.Scan() {
					if strings.Contains(scanner.Text(), "[ffmpeg] Destination:") {
						ytdlPath = strings.Split(scanner.Text(), ": ")[1]
					}
				}
				//fmt.Println(ytdlPath)
				//fmt.Println(strings.Split(scanner.Text(), ": "))

				mp3FileLocalStyle := ytdlPath
				mp3FileRemoteStyle := "/" + strings.Replace(ytdlPath, "\\", "/", -1)
				//fmt.Println(mp3FileLocalStyle)
				//fmt.Println(mp3FileRemoteStyle)

				fileSize, _ = getFileSize(mp3FileLocalStyle)

				if dropboxFolder != remote {
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
						fmt.Fprintf(os.Stderr, "error9: %v\n", err)
						os.Exit(1)
					}
				}
				var dropboxURL string
				dropboxURL, err = utils.GetDropboxURLWhenSyncComplete(mp3FileRemoteStyle)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error8: %v\n", err)
					os.Exit(1)
				}
				// fmt.Println(dropboxURL)

				//TODO - Don't add DB entry until I'm 100% sure the whole end-to-end flow has worked for that entry including Dropbox Sync
				fmt.Printf("Adding new RSS Entry:   %s \n", item.Title)
				entry := YTSubscriptionEntry{
					Subscription: subscription.ID,
					URL:          item.Link,
					Title:        item.Title,
					Date:         item.Date,
					DropboxURL:   dropboxURL,
					FileSize:     fileSize,
				}

				fmt.Println(entry)

				err = db.Save(&entry)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error4: %v\n", err)
				}

			}
		}

	}

	// Add all entries to struct which will be used to generate the rss.xml file
	var ytAllSubscriptionEntries []YTSubscriptionEntry
	err = db.AllByIndex("Date", &ytAllSubscriptionEntries)
	if err != nil && err.Error() != "bucket YTSubscription not found" && err.Error() != "bucket YTSubscriptionEntry not found" {
		fmt.Fprintf(os.Stderr, "error5: %v\n", err)
	}

	for _, ytItem := range ytAllSubscriptionEntries {
		addEntrytoRSSXML(ytItem)
		//fmt.Println(ytItem)
	}

	// Create and update rss.xml
	_, err = generateRSS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error7: %v\n", err)
		os.Exit(1)
	}

	// Copy rss.xml to Dropbox
	if dropboxFolder != remote {
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
	RSSFileURL, err := utils.GetDropboxURLWhenSyncComplete("/rss.xml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error6: %v\n", err)
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

	//	layOut := "2006-01-02 15:04:05-07:00"
	//	timeStamp, err := time.Parse(layOut, ytItem.Date)
	timeStamp := ytItem.Date

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

	rss, error := RSSXML.ToAtom()
	//rss, err := RSSXML.ToRss()
	if error != nil {
		return "", error
	}

	rssOut := []byte(rss)

	rssFile := "rss.xml"
	err := ioutil.WriteFile(rssFile, rssOut, 0644)
	if err != nil {
		return "", err
	}
	return rssFile, nil
}

func getFileSize(srcFile string) (int64, error) {
	fi, error := os.Stat(srcFile)
	if error != nil {
		return 0, error
	}
	// fmt.Printf("The file is %d bytes long", fi.Size())
	return fi.Size(), nil

}
