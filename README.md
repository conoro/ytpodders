# YTPodders
YTPodders - Get your favourite YouTubers as audio podcasts on your phone.

YTPodders creates subscribable MP3 podcasts, for your personal use, from YouTube Users and Channels using Dropbox.


## End-User Installation and First Time Run
* Download the App (and all of the tools) as one zip file here:
  * [Windows](http://ytpodders.s3.amazonaws.com/dist/zips/ytpodders_windows_v2.0.1.zip)
  * [Mac](http://ytpodders.s3.amazonaws.com/dist/zips/ytpodders_mac_v2.0.1.zip). You need Python already installed and remember to

  ```
  chmod a+rx youtube-dl youtube-dl.sh ytpodders ffmpeg ffprobe

  ```

  * [Linux](http://ytpodders.s3.amazonaws.com/dist/zips/ytpodders_linux_v2.0.1.zip). Remember to

```
chmod a+rx youtube-dl youtube-dl.sh ytpodders ffmpeg ffprobe

```

* Unzip into a directory of your choice
* Go to https://ytpodders.com and follow the Authorisation flow to get an access token
* Copy the access token provided above to the obvious place in client_conf.json
* Open a CMD prompt or shell and cd to the directory where you unzipped ytpodders
* Add subscriptions to each of your favourite YouTubers using

```
ytpodders add https://www.youtube.com/url_of_your_fave_youtuber
```
* For example:

```
ytpodders add https://www.youtube.com/user/TheGingerRunner
ytpodders add https://www.youtube.com/channel/UCh8rjWtGCIAbwPrZb3Te8bQ
```

* Grab all the latest entries using:

```
ytpodders.exe or ./ytpodders, depending on your platform
```

* After the run is completed, take the RSS URL presented and paste it into your Podcasting App on your phone e.g. [BeyondPod](http://www.beyondpod.mobi/android/index.htm) on Android or the built-in iPhone Podcasting App
* You can continue to add/delete/list/enable/disable subscriptions and run ytpodders whenever you wish to get the latest
* It takes just a minute to add ytpodders to the Windows Task Scheduler so that it runs automatically whenever you want. Ditto as a cronjob on Linux.

## Developer Installation
It uses [youtube-dl](https://rg3.github.io/youtube-dl/) and [ffmpeg](https://www.ffmpeg.org/) to do all of the heavy lifting and stores all of your subscriptions in BoltDB.

You can download the tools here:

* [youtube-dl](https://rg3.github.io/youtube-dl/)
* [ffmpeg](https://www.ffmpeg.org/download.html)
* [ffprobe](https://www.ffmpeg.org/download.html)

All the generated MP3s are stored in your Dropbox Folder and an rss.xml file is generated which can be used by your phone's Podcasting App to subscribe.

As it is written in Go, it should work on every platform where youtube-dl and ffmpeg are available.

### Auth Flow
* Go to https://www.dropbox.com/developers/apps/create
* Login with your Dropbox account
* In "Choose an API" Click the Dropbox API option
* In "Choose the type of access you need" click App Folder
* In "Name your App", enter YTPodders
* Click "Create App"
* Note the App Key
* Click "Show" on App Secret and note it too
* In redirect URL enter http://127.0.0.1:7000 and/or whatever server URL you'll be using
* On your client machine download the [ytpodders source code](https://github.com/conoro/ytpodders) via git clone or [zip download](https://github.com/conoro/ytpodders/archive/master.zip) into $GOPATH/src/github.com/conoro/ytpodders
* Run a local web-server like [Caddy](https://caddyserver.com/) in the static directory
* You can now browse to http://url_of_your_local_server and follow the auth flow to get a token. Note it down

### Client
* Download, install and configure [Go](http://www.golang.org/)
* On your client machine download the [ytpodders source code](https://github.com/conoro/ytpodders) via git clone or [zip download](https://github.com/conoro/ytpodders/archive/master.zip) into $GOPATH/src/github.com/conoro/ytpodders
* Download youtube-dl, ffmpeg and ffprobe as above to that directory too
* go build github.com/conoro/ytpodders to create the ytpodders binary
* The rest of the steps are the same as the End-User ones.


## Usage and Commands
You can use the following commands:
- [x] no command - updates everything as you'd expect. Normal one-off execution
- [x] help - print help out and exit
- [x] add - add a subscription. Pass it the URL of a YouTube Channel or User (TODO: sanitize input)
- [x] list - list all subscriptions
- [x] delete - delete a subscription by ID

TODO
- [ ] dryrun - same as run except nothing is downloaded and the database is not modified but it lists what it would do
- [ ] prune - pass a number of days as param. Mark entries in DB as "expired". Delete mp3 files locally and from Dropbox. Do not re-download on next run!
- [ ] reinit - Clear the DB completely and delete both local and Dropbox MP3s

## More TODOs
- [ ] Switch back to RSS (http://www.danfergusdesign.com/classfiles/generalReference/rssFeedSample.php) from ATOM so I can use elements like &lt;itunes:image href="https://ytpodders.com/YTPodders_1400x1400_iTunes.png"/&gt;
- [ ] Add configurable limit on number of videos to download per channel (default 5 seems reasonable)
- [ ] Improved Error Checking and Handling
- [ ] Add proper Go-style logging everywhere
- [ ] Add proper Set max retention time as parameter in conf.json
- [ ] sanitize the addition of URLs
- [ ] validate removal by ID
- [ ] remove all MP3s locally and on Dropbox when deleting
- [ ] remove all the entries in subscription_entries when deleting
- [x] Generate releases and upload to S3 using some CI tool (OSX a challenge)
- [ ] Automatically do the Auth flow and save client_conf.json for non-technical users
- [ ] Use the Web App to add/remove/manage subscriptions for very non-technical users

## Release History
* 2016-05-02 - V1.0.0 - Initial release
* 2016-05-11 - V1.1.0 - Replaced SQLite/sqlx with BoltDB/storm for data store. Added simple build script for Windows.
* 2016-08-10 - V1.1.1 - Quick hack to handle filenames with non-ascii characters due to youtube-dl oddness.
* 2016-08-11 - V1.1.2 - Updated to latest version of youtube-dl to handle download errors on some videos
* 2017-03-27 - V1.1.3 - Updated to latest version of youtube-dl. Fixed delete command.
* 2017-10-14 - V2.0.0 - Updated to Dropbox V2 API. Removed server command and replaced with simple static flow. Updated youtube-dl
* 2017-10-14 - V2.0.1 - Fix some bugs with Dropbox URLs

