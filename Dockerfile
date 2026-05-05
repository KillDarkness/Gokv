FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gokv ./cmd/gokv

FROM alpine:3.20

RUN adduser -D -H gokv
USER gokv

COPY --from=build /out/gokv /usr/local/bin/gokv

EXPOSE 6379
ENTRYPOINT ["gokv"]
