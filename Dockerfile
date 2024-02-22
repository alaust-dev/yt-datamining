FROM golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/

RUN go build -a -o yt-datamining cmd/yt-datamining/main.go

FROM alpine:latest

RUN apk update && apk upgrade

USER 1000
WORKDIR /app

COPY --from=build /app/yt-datamining .

CMD ["./yt-datamining"]
