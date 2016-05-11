:: Make sure AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are set as environment variables

set RELEASE=%1

set GOOS=linux
set GOARCH=amd64
go build github.com/conoro/ytpodders
mkdir dist
mkdir dist\ytpodders_linux
copy ytpodders dist\ytpodders_linux\
copy youtube-dl.sh dist\ytpodders_linux\
set GOOS=darwin
set GOARCH=amd64
go build github.com/conoro/ytpodders
mkdir dist\ytpodders_mac
copy ytpodders dist\ytpodders_mac\
copy youtube-dl.sh dist\ytpodders_mac\
set GOOS=windows
set GOARCH=amd64
go build github.com/conoro/ytpodders
mkdir dist\ytpodders_windows
copy ytpodders.exe dist\ytpodders_windows\
mkdir zips
cd dist
d:\apps\bin\zip -r zips\ytpodders_windows_%RELEASE%.zip ytpodders_windows
d:\apps\bin\zip -r zips\ytpodders_mac_%RELEASE%.zip ytpodders_mac
d:\apps\bin\zip -r zips\ytpodders_linux_%RELEASE%.zip ytpodders_linux
cd zips
gof3r cp --endpoint s3-eu-west-1.amazonaws.com --acl public-read ytpodders_windows_%RELEASE%.zip s3://ytpodders/dist/zips/ytpodders_windows_%RELEASE%.zip
gof3r cp --endpoint s3-eu-west-1.amazonaws.com --acl public-read ytpodders_mac_%RELEASE%.zip s3://ytpodders/dist/zips/ytpodders_mac_%RELEASE%.zip
gof3r cp --endpoint s3-eu-west-1.amazonaws.com --acl public-read ytpodders_linux_%RELEASE%.zip s3://ytpodders/dist/zips/ytpodders_linux_%RELEASE%.zip
