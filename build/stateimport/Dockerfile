# Build the manager binary
ARG GO_VERSION=1.21
FROM golang:${GO_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

LABEL MAINTAINER="dipankar das"

COPY go.mod go.mod
COPY go.sum go.sum

COPY pkg/ /app/pkg/
COPY internal/ /app/internal/
COPY ksctl-components/stateimport/ /app/ksctl-components/stateimport/

RUN go mod download

RUN cd ksctl-components/stateimport && CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -v -o ksctlstateimport . && mv ksctlstateimport ../..


FROM alpine

LABEL MAINTAINER="dipankar das"

RUN apk add openssh

RUN adduser --disabled-password -s /bin/sh -u 1000 ksctl

USER ksctl

WORKDIR /app


COPY --from=builder /app/ksctlstateimport /app/ksctlstateimport

ENTRYPOINT [ "/app/ksctlstateimport" ]

EXPOSE 80
