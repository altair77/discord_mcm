FROM golang:1.17.8-buster

WORKDIR /go/src/app

COPY *.go go.mod go.sum ./

RUN go mod download

CMD ["go", "run", "."]
