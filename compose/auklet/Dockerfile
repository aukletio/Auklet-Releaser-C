FROM golang

# Install dep
RUN curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
RUN chmod +x $GOPATH/bin/dep

# Copy startup scripts.
COPY ./compose/auklet/runWrapper.sh /runWrapper.sh
COPY ./compose/auklet/makeRelease.sh /makeRelease.sh
RUN chmod +x /runWrapper.sh /makeRelease.sh

WORKDIR /go/src/github.com/user/myProject/app