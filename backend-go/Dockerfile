# Build stage
FROM golang:alpine AS builder

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN swag init

RUN go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 9001

CMD ["./main"]
