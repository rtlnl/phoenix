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
COPY --from=builder /app/bin /app/

EXPOSE 8080 8081
ENTRYPOINT ["/app/api"]