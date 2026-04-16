# Build stage
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod ./
# Copy go.sum if it exists
# COPY go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 9001

CMD ["./main"]
