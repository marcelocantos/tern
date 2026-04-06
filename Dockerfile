FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go agents-guide.md ./
COPY crypto/ crypto/
COPY protocol/ protocol/
COPY qr/ qr/
COPY cmd/pigeon/ cmd/pigeon/
RUN CGO_ENABLED=0 go build -o /pigeon ./cmd/pigeon

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && mkdir -p /data/certmagic
COPY --from=build /pigeon /pigeon
EXPOSE 443/udp 443/tcp 4433/udp
CMD ["/pigeon"]
