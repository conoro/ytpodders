#!/bin/sh
# Watch out for Windows Editors and Git changing the line ending to CRLF. Need to stay as LF
export PATH=$PATH:$PWD
./youtube-dl "$@"
