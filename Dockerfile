FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

# Stage 2: Run
FROM alpine:3.21

WORKDIR /root/

COPY --from=builder /app/main .

COPY .env .

EXPOSE 50052

CMD ["./main"]
