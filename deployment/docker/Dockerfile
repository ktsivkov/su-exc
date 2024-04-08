FROM golang:1.22-alpine as builder

RUN export GOPATH=$HOME/go
RUN export PATH=$PATH:$GOPATH/bin

COPY ./ /app
WORKDIR /app

RUN go build -o application.out /app/cmd/server/main.go

FROM scratch

WORKDIR /

COPY --from=builder /app/application.out /application.out

ENTRYPOINT ["/application.out"]
