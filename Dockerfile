FROM golang

RUN go get -u github.com/mmcdole/gofeed

ADD . /go/src/github.com/chpwssn/castdown

RUN go install github.com/chpwssn/castdown

ENTRYPOINT /go/bin/castdown
