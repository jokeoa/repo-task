FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /shipment-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /shipment-service /shipment-service

EXPOSE 50051

ENTRYPOINT ["/shipment-service"]
