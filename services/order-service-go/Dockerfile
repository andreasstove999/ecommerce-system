FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o order-service ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/order-service /app/order-service

ENV PORT=8082

EXPOSE 8082

ENTRYPOINT ["/app/order-service"]
