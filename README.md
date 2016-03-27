# YTPodders
YTPodders creates subscribable MP3 podcasts from YouTube Users and Channels using Dropbox.

It uses [youtube-dl](https://rg3.github.io/youtube-dl/) and [ffmpeg](https://www.ffmpeg.org/) to do all of the heavy lifting and stores all of your subscriptions in SQLite. You can download these here:

* [youtube-dl]()
* [ffmpeg](https://www.ffmpeg.org/download.html) and [ffprobe](https://www.ffmpeg.org/download.html)

All the generated MP3s are stored in your Dropbox Folder and an rss.xml file is generated which can be used by your phone's Podcasting App to subscribe.

As it is written in Go, it should work on every platform where youtube-dl and ffmpeg are available.

## End-User Installation and First Time Run
* Go to https://www.dropbox.com/developers/apps/create
* Login with your Dropbox account
* In "Choose an API" Click the Dropbox API option
* In "Choose the type of access you need" click App Folder
* In "Name your App", enter YTPodders
* Click "Create App"
* Note the App Key
* Click "Show" on App Secret and note it
* Download conf.json, YTPodders binary, youtube-dl, ffmpeg and ffprobe binaries as one download here.
* Unzip that file and put everything in a directory of your choosing
* TODO: Possibly make the above a single Inno installer for Windows.
* Edit conf.json and put App Key and App Secret from above in the relevant place in the file and save it
* Open a Windows CMD window, cd to your directory and type ytpodders
* You'll be shown a URL, copy that and open it in your web browser
* You'll be asked to give permission. Do so and copy the code it gives you
* Go back to your command window and paste the key into the prompt
* You'll be given a permanent key
* Edit conf.json again and paste the key into the relevent place in the file
* YTPodders will have exited.
Now you can add subscriptions and have YTPodders generate your Podcast feed for you on Dropbox as follows:
  * ytpodders add url_of_youtube_user_or_channel
  * ytpodders run
  * After the run is completed, take the RSS URL presented and paste it into your Podcasting App on your phone e.g. [BeyondPod](http://www.beyondpod.mobi/android/index.htm) on Android or the built-in iPhone Podcasting App
  * You can continue to add/delete/list/enable/disable subscriptions and run ytpodders whenever you wish to get the latest
  * Linux and OSX users can setup a Cronjob to do this automatically whenever they want
  * TODO: Windows users scheduling

## Development Installation
* Basically the same as the above except
* Download, install and configure [Go](http://www.golang.org/)
* Download the [ytpodders source code](https://github.com/conoro/ytpodders) via git clone or [zip download](https://github.com/conoro/ytpodders/archive/master.zip) into $GOPATH/src/github.com/conoro/ytpodders
* Download youtube-dl, ffmpeg and ffprobe binaries to the same directory or somewhere on your path
* go build github.com/conoro/ytpodders to create the ytpodders binary


## Usage and Commands
You can use the following commands:
- [x] no command - updates everything as you'd expect. Normal one-off execution
- [x] help - print help out and exit
- [x] add - add a subscription. Pass it the URL of a YouTube Channel or User (possibly sanitize)
- [x] list - list all subscriptions as ID, URL, Title maybe
- [x] delete - delete a subscription by ID

TODO
- [ ] scheduler - runs it as some sort of background daemon. No idea how to to this on Windows
- [ ] dryrun - same as run except nothing is downloaded and the database is not modified but it lists what it would do
- [ ] prune - pass a number of days as param. Mark entries in DB as "expired". Delete mp3 files locally and from Dropbox. Do not re-download on next run!
- [ ] reauth - re-run the Auth flow to get a new Dropbox token
- [ ] reinit - Clear the DB completely and delete both local and Dropbox MP3s

## More TODOs
- [ ] Error Checking and Handling
- [ ] Add proper Go-style logging everywhere instead of all of these Print statements
- [ ] clean up output
- [ ] Add proper Set max retention time as parameter in conf.json
- [ ] sanitize the addition of URLs
- [ ] validate removal by ID.
- [ ] remove all MP3s locally and on Dropbox when deleting
- [ ] remove all the entries in subscription_entries when deleting
