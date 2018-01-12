FROM debian:8
USER root
ENV DEBIAN_FRONTEND noninteractive

# Install curl
RUN apt-get -qq update
RUN apt-get install -y curl

# Install Go
RUN curl -O https://dl.google.com/go/go1.9.2.linux-amd64.tar.gz && tar xvf go1.9.2.linux-amd64.tar.gz
ENV GOPATH=$HOME/go
ENV PATH=$PATH:$GOPATH/bin

# Install dep
RUN curl -L -s https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 -o $GOPATH/bin/dep
RUN chmod +x $GOPATH/bin/dep

# Copy startup scripts.
COPY ./compose/auklet/startAuklet.sh /startAuklet.sh
RUN chmod +x /startAuklet.sh

# Set working dir to app root for running commands.
WORKDIR /app
