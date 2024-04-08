FROM golang:1.22-alpine

RUN export GOPATH=$HOME/go
RUN export PATH=$PATH:$GOPATH/bin

WORKDIR /app
