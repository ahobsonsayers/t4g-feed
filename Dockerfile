# Builder Image
FROM golang:1.21 as builder

WORKDIR /t4g-feed
COPY . .
RUN go mod download
RUN go build -v -o bin/t4g-feed

# Ditribution Image
FROM alpine:latest

RUN apk add --no-cache libc6-compat

COPY --from=builder /t4g-feed/bin/t4g-feed /t4g-feed

EXPOSE 5656

ENTRYPOINT ["/t4g-feed"]
