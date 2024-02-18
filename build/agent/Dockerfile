FROM golang:1.21 as builder

WORKDIR /app

LABEL MAINTAINER="dipankar das"

COPY . .

RUN go mod tidy

RUN cd ksctl-components/agent && CGO_ENABLED=0 go build -v -o ksctlagent . && mv ksctlagent ../..


FROM alpine

LABEL MAINTAINER="dipankar das"

RUN apk add openssh

RUN adduser --disabled-password -s /bin/sh -u 1000 ksctl

USER ksctl

WORKDIR /app

COPY --from=builder /app/ksctlagent /app/ksctlagent

ENTRYPOINT [ "/app/ksctlagent" ]

EXPOSE 8080
