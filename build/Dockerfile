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

WORKDIR /root

COPY --from=builder /go/src/ksctl /usr/local/bin/ksctl

RUN mkdir -p ${HOME}/.ksctl/cred && \
    touch ${HOME}/.ksctl/cred/aws && \
    touch ${HOME}/.ksctl/cred/azure && \
    touch ${HOME}/.ksctl/cred/civo && \
    mkdir -p ${HOME}/.ksctl/config/civo && \
    mkdir ${HOME}/.ksctl/config/azure && \
    mkdir ${HOME}/.ksctl/config/aws && \
    mkdir ${HOME}/.ksctl/config/local

RUN wget -O kubectl https://dl.k8s.io/release/v1.25.0/bin/linux/amd64/kubectl && \
    chmod +x kubectl && mv ./kubectl /usr/local/bin/kubectl
