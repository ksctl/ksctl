FROM golang:1.19 AS builder

WORKDIR /go/src/

COPY go.mod go.sum ./

RUN go mod download

CMD ["go", "test", "-v", "."]