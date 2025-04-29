# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder
WORKDIR /app

COPY src .

# download cache & build static app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main

FROM scratch
WORKDIR /app

# Copy binary only
COPY --from=builder /app/main .

# Start binary
ENTRYPOINT ["./main"]
