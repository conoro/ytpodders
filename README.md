# YTPodders
YTPodders creates subscribable MP3 podcasts from YouTube Users and Channels using Dropbox.

It uses [youtube-dl]() and [ffmpeg]() to do all of the heavy lifting and stores all of your subscriptions in SQLite. You can download these here:

* [youtube-dl]()
* [ffmpeg]() and [ffprobe]()

All the generated MP3s are stored in your Dropbox Folder and an rss.xml file is generated which can be used by your phone's Podcasting App to subscribe.

You can use the following commands:
- [x] no command - updates everything as you'd expect. Normal one-off execution
- [x] help - print help out and exit

TODO
- [ ] add - add a subscription. Pass it the URL of a YouTube Channel or User (possibly sanitize)
- [ ] list - list all subscriptions as ID, URL, (Title maybe? or Uploader maybe?)
- [ ] remove - remove a subscription by ID
- [ ] scheduler - runs it as some sort of background daemon. No idea how to to this on Windows
- [ ] dryrun - same as run except nothing is downloaded and the database is not modified but it lists what it would do
- [ ] prune - pass a number of days as param. Mark entries in DB as "expired". Delete mp3 files locally and from Dropbox. Do not re-download on next run!
- [ ] reauth - re-run the Auth flow to get a new Dropbox token
- [ ] reinit - Clear the DB completely and delete both local and Dropbox MP3s
