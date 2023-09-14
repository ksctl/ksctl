FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 make build_httpserver

FROM alpine

WORKDIR /app

COPY --from=builder /app/ksctlserver /app/ksctlserver
COPY --from=builder /app/httpserver/swaggerui /app/swaggerui
COPY --from=builder /app/httpserver/gen /app/gen

ENTRYPOINT [ "/app/ksctlserver" ]

EXPOSE 8080
