FROM golang

# Install dep
RUN curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
RUN chmod +x $GOPATH/bin/dep

# Copy startup scripts.
COPY ./compose/auklet/startAuklet.sh /startAuklet.sh
COPY ./compose/auklet/makeRelease.sh /makeRelease.sh
RUN chmod +x /startAuklet.sh /makeRelease.sh

WORKDIR /go/src/github.com/user/myProject/app