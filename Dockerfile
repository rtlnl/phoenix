# build stage
FROM golang:1.13 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/api

# final stage
FROM alpine:3.9
COPY --from=builder /app /app

# ENV GIN_MODE=release

RUN apk update && \
    apk add ca-certificates && \
    rm -rf /var/cache/apk/*

WORKDIR /app

EXPOSE 8082 8081
ENTRYPOINT ["/app/bin/api"]