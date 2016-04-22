# YTPodders
YTPodders creates subscribable MP3 podcasts from YouTube Users and Channels using Dropbox.

It uses [youtube-dl](https://rg3.github.io/youtube-dl/) and [ffmpeg](https://www.ffmpeg.org/) to do all of the heavy lifting and stores all of your subscriptions in SQLite. You can download these here:

* [youtube-dl]()
* [ffmpeg](https://www.ffmpeg.org/download.html) and [ffprobe](https://www.ffmpeg.org/download.html)

All the generated MP3s are stored in your Dropbox Folder and an rss.xml file is generated which can be used by your phone's Podcasting App to subscribe.

As it is written in Go, it should work on every platform where youtube-dl and ffmpeg are available.

## End-User Installation and First Time Run
* Download the App as [one download here]()
* Unzip into a directory of your choice
* Go to https://ytpodders.com and follow the permission flow
* Copy the access token provided above to the obvious place in conf_client.json
* Open a CMD prompt and cd to the directory where you unzipped ytpodders
* ytpodders add https://www.youtube.com/url_of_your_fave_youtuber
* ytpodders
* ytpodders.exe or ./ytpodders, depending on your platform
* After the run is completed, take the RSS URL presented and paste it into your Podcasting App on your phone e.g. [BeyondPod](http://www.beyondpod.mobi/android/index.htm) on Android or the built-in iPhone Podcasting App
* You can continue to add/delete/list/enable/disable subscriptions and run ytpodders whenever you wish to get the latest
* It takes just a minute to add ytpodders to the Windows Task Scheduler so that it runs automatically whenever you want. Ditto as a cronjob on Linux.

## Developer Installation

### Web Server
* Go to https://www.dropbox.com/developers/apps/create
* Login with your Dropbox account
* In "Choose an API" Click the Dropbox API option
* In "Choose the type of access you need" click App Folder
* In "Name your App", enter YTPodders
* In redirect URL enter TODO TODO
* Click "Create App"
* Note the App Key
* Click "Show" on App Secret and note it
* Download, install and configure [Go](http://www.golang.org/)
* On your Web Server (even local server) Download the [ytpodders source code](https://github.com/conoro/ytpodders) via git clone or [zip download](https://github.com/conoro/ytpodders/archive/master.zip) into $GOPATH/src/github.com/conoro/ytpodders
* go build github.com/conoro/ytpodders to create the ytpodders binary
* go server
* You can now browse to http://url_of_your_server to get an access token. Note it down

### Client
* Download conf.json, YTPodders binary, youtube-dl, ffmpeg and ffprobe binaries as [one download here]().
* On your client machine download the [ytpodders source code](https://github.com/conoro/ytpodders) via git clone or [zip download](https://github.com/conoro/ytpodders/archive/master.zip) into $GOPATH/src/github.com/conoro/ytpodders
* go build github.com/conoro/ytpodders to create the ytpodders binary
* The rest of the steps are the same as the End-User ones.


## Usage and Commands
You can use the following commands:
- [x] no command - updates everything as you'd expect. Normal one-off execution
- [x] help - print help out and exit
- [x] add - add a subscription. Pass it the URL of a YouTube Channel or User (TODO: sanitize input)
- [x] list - list all subscriptions as ID, URL, Title maybe
- [x] delete - delete a subscription by ID

TODO
- [ ] server - run a web-server as a developer which lets users authorise the app to access Dropbox
- [ ] dryrun - same as run except nothing is downloaded and the database is not modified but it lists what it would do
- [ ] prune - pass a number of days as param. Mark entries in DB as "expired". Delete mp3 files locally and from Dropbox. Do not re-download on next run!
- [ ] reauth - re-run the Auth flow to get a new Dropbox token
- [ ] reinit - Clear the DB completely and delete both local and Dropbox MP3s

## More TODOs
- [ ] Switch back to RSS (http://www.danfergusdesign.com/classfiles/generalReference/rssFeedSample.php) from ATOM so I can use elements like <itunes:image href="https://ytpodders.com/YTPodders_1400x1400_iTunes.png"/>
- [ ] scheduling on OSX how? cron?
- [ ] Add configurable limit on number of videos to download per channel (default 5 seems reasonable)
- [ ] Improved Error Checking and Handling
- [ ] Add proper Go-style logging everywhere instead of all of these Print statements
- [ ] Add proper Set max retention time as parameter in conf.json
- [ ] sanitize the addition of URLs
- [ ] validate removal by ID.
- [ ] remove all MP3s locally and on Dropbox when deleting
- [ ] remove all the entries in subscription_entries when deleting
