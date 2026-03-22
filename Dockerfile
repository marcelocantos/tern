FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 go build -o /tern .

FROM alpine:3.21
COPY --from=build /tern /tern
EXPOSE 8080
CMD ["/tern"]
