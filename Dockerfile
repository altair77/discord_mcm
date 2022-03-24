FROM golang:1.17.8-alpine

RUN apk add --no-cache openjdk17-jre

WORKDIR /go/src/app

COPY *.go go.mod go.sum ./

RUN go mod download

CMD ["go", "run", "."]
