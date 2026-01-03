# Build stage
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# main is in ./cmd (cmd/main.go)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway ./cmd

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/api-gateway /app/api-gateway

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/api-gateway"]
