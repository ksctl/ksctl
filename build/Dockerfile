FROM golang:1.19

WORKDIR /go/src

COPY go.sum .

COPY go.mod .

RUN go get -d || true

COPY . .

RUN go build -o kubesimpctl
