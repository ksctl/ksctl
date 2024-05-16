# Build the manager binary
ARG GO_VERSION=1.21
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

LABEL MAINTAINER="dipankar das"

COPY . .

RUN go mod download

RUN cd ksctl-components/agent && CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -v -o ksctlagent . && mv ksctlagent ../..


FROM alpine

LABEL MAINTAINER="dipankar das"

RUN apk add openssh

RUN adduser --disabled-password -s /bin/sh -u 1000 ksctl

USER ksctl

WORKDIR /app

COPY --from=builder /app/ksctlagent /app/ksctlagent

ENTRYPOINT [ "/app/ksctlagent" ]

EXPOSE 8080
