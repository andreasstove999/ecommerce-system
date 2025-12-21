FROM golang:1.24.5 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o inventory-service ./cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=build /app/inventory-service /app/inventory-service

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/inventory-service"]
