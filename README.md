# castdown

A golang based app to download podcasts locally as a backup.

## Quickstart

    go get -u github.com/chpwssn/castdown
    go get -u github.com/mmcdole/gofeed
    cd $GOPATH/src/github.com/chpwssn/castdown
    go build
    cp castdown /where/you/want/to/download/podcasts/castdown
    cp config.example.json /where/you/want/to/download/podcasts/config.json
    cd /where/you/want/to/download/podcasts/
    # edit config.json
    ./castdown
