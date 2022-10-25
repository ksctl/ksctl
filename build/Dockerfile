FROM golang:1.19 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /go/src

COPY go.sum .

COPY go.mod .

RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -o ksctl


FROM docker
#FROM alpine:3.16.2

WORKDIR /home/root

COPY --from=builder /go/src/ksctl /usr/local/bin/ksctl

RUN mkdir -p /home/$(whoami)/.kube/ksctl/cred && \
    touch /home/$(whoami)/.kube/ksctl/cred/aws && \
    touch /home/$(whoami)/.kube/ksctl/cred/azure && \
    touch /home/$(whoami)/.kube/ksctl/cred/civo && \
    mkdir -p /home/$(whoami)/.kube/ksctl/config/civo && \
    mkdir /home/$(whoami)/.kube/ksctl/config/azure && \
    mkdir /home/$(whoami)/.kube/ksctl/config/aws && \
    mkdir /home/$(whoami)/.kube/ksctl/config/local

RUN wget -O kubectl https://dl.k8s.io/release/v1.25.0/bin/linux/amd64/kubectl

RUN chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
