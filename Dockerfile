FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .

RUN mkdir -p /out/data && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/gokv ./cmd/gokv

FROM alpine:3.20

RUN adduser -D -H -s /sbin/nologin gokv && mkdir -p /data && chown gokv:gokv /data

COPY --from=build /out/gokv /usr/local/bin/gokv

VOLUME ["/data"]
EXPOSE 6379
ENTRYPOINT ["gokv"]
