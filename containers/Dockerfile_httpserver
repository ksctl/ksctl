FROM golang:1.21 as builder

WORKDIR /app

LABEL MAINTAINER="dipankar das"

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 make build_httpserver


FROM alpine

LABEL MAINTAINER="dipankar das"

WORKDIR /app

COPY --from=builder /app/ksctlserver /app/ksctlserver
COPY --from=builder /app/httpserver/swaggerui /app/swaggerui
COPY --from=builder /app/httpserver/gen /app/gen

RUN apk add openssh

ENV KSCTL_TEST_DIR_ENABLED=/app/ksctl-data

RUN mkdir -p /app/ksctl-data/config/civo/ha && \
	mkdir /app/ksctl-data/config/civo/managed && \
	mkdir -p /app/ksctl-data/config/azure/managed && \
	mkdir /app/ksctl-data/config/azure/ha

ENTRYPOINT [ "/app/ksctlserver" ]

EXPOSE 8080
