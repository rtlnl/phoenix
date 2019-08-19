# build stage
FROM golang:1.12 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/api

# final stage
FROM scratch
COPY --from=builder /app /app

ENV GIN_MODE=release

WORKDIR /app

EXPOSE 8082 8081
ENTRYPOINT ["/app/bin/api"]