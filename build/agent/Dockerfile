# Build the manager binary
ARG GO_VERSION=1.21
FROM golang:${GO_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

LABEL MAINTAINER="dipankar das"

COPY . .

RUN go mod tidy

RUN cd ksctl-components/agent && CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -v -o ksctlagent . && mv ksctlagent ../..


FROM alpine

LABEL MAINTAINER="dipankar das"

RUN apk add openssh

RUN adduser --disabled-password -s /bin/sh -u 1000 ksctl

USER ksctl

WORKDIR /app

COPY --from=builder /app/ksctlagent /app/ksctlagent

RUN mkdir -p /home/ksctl/.config/helm && \
    touch /home/ksctl/.config/helm/repositories.yaml

ENTRYPOINT [ "/app/ksctlagent" ]

EXPOSE 8080

## Use distroless as minimal base image to package the manager binary
## Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
#WORKDIR /
#COPY --from=builder /workspace/manager .
#USER 65532:65532
#
#ENTRYPOINT ["/manager"]