FROM golang:1.19

WORKDIR /go/src

COPY go.sum .

COPY go.mod .

RUN go get -d || true

COPY . .

RUN go build -o ksctl

RUN cp -v ksctl /usr/local/bin

RUN mkdir /etc/ksctl

RUN mkdir -p /home/$(whoami)/.kube/ksctl/cred && \
    touch /home/$(whoami)/.kube/ksctl/cred/aws && \
    touch /home/$(whoami)/.kube/ksctl/cred/azure && \
    touch /home/$(whoami)/.kube/ksctl/cred/civo && \
    mkdir -p /home/$(whoami)/.kube/ksctl/config/civo && \
    mkdir /home/$(whoami)/.kube/ksctl/config/azure && \
    mkdir /home/$(whoami)/.kube/ksctl/config/aws && \
    mkdir /home/$(whoami)/.kube/ksctl/config/local
