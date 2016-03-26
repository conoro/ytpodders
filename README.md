# YTPodders
YTPodders creates subscribable MP3 podcasts from YouTube Users and Channels using Dropbox.

It uses [youtube-dl](https://rg3.github.io/youtube-dl/) and [ffmpeg](https://www.ffmpeg.org/) to do all of the heavy lifting and stores all of your subscriptions in SQLite. You can download these here:

* [youtube-dl]()
* [ffmpeg](https://www.ffmpeg.org/download.html) and [ffprobe](https://www.ffmpeg.org/download.html)

All the generated MP3s are stored in your Dropbox Folder and an rss.xml file is generated which can be used by your phone's Podcasting App to subscribe.

As it is written in Go, it should work on every platform where youtube-dl and ffmpeg are available.

## Development Installation
* TODO Download, install and configure Go
* Download the ytpodders source code via git clone or zip download
* Download youtube-dl, ffmpeg and ffprobe binaries to the same directory as ytpodders or somewhere on your path
* TODO Setup the App up on Dropbox
* TODO Setup conf.json with the Dropbox credentials

## End-User Installation
* TODO Download ytpodders binary for your platform
* Download youtube-dl, ffmpeg and ffprobe to the same directory as ytpodders or somewhere on your path
* TODO Setup path to download to on Dropbox
* ytpodders add url_of_youtube_user_or_channel

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
