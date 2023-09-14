FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go get -d -v

RUN CGO_ENABLED=0 go build -v -o ksctl cmd/server/*

FROM scratch

WORKDIR /app

COPY --from=builder /app/ksctl /app/ksctl
COPY --from=builder /app/swaggerui /app/swaggerui
COPY --from=builder /app/gen /app/gen

ENTRYPOINT [ "/app/ksctl" ]

EXPOSE 8080
