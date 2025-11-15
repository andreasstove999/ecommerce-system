# Build stage
FROM golang:1.24.5 as builder

WORKDIR /app

# Go modules
COPY go.mod go.sum ./
RUN go mod download

# Source
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cart-service ./cmd/cart-service

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/cart-service /app/cart-service

ENV PORT=8081

EXPOSE 8081

ENTRYPOINT ["/app/cart-service"]
