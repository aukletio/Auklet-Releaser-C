FROM golang

# Install dep
RUN curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
RUN chmod +x $GOPATH/bin/dep

WORKDIR /go/src/github.com/aukletio/Auklet-Releaser-C/app
