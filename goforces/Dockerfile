FROM golang:1.23.4-bookworm AS builder
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./goforces .
RUN GOOS=linux GOARCH=amd64 \
    go build -o app ./cmd/main.go
FROM alpine:latest AS final
EXPOSE 8080
COPY --from=builder /app/app /app/app
WORKDIR /app
CMD ["./app"]
