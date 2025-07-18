FROM golang:1.24-alpine AS builder

WORKDIR /app
# Copy go.mod and go.sum from the parent directory if present, else just go.mod
COPY ../go.mod ./
COPY ../go.sum ./
RUN go mod download
COPY ../ .
RUN go build -o main ./app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]