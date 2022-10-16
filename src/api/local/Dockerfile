FROM golang:1.19 AS localTest

WORKDIR /go/src

COPY ../go.sum .

COPY ../go.mod .

RUN go get -d || true

COPY ../ .

RUN mkdir -p /home/$(whoami)/.kube/kubesimpctl/config/local

WORKDIR /go/src/local

CMD ["go", "test", "-v", "."]