FROM golang:1.20-alpine as golang

WORKDIR /app
COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /dist .

FROM alpine:latest

ARG slack-token

COPY --from=golang /dist .

ENTRYPOINT ["/dist"]