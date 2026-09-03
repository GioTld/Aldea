FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/trackerd ./cmd/trackerd
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/noded ./cmd/noded
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/aldea ./cmd/aldea

FROM alpine:latest

RUN apk add --no-cache ca-certificates bash curl

COPY --from=builder /app/bin/trackerd /usr/local/bin/trackerd
COPY --from=builder /app/bin/noded /usr/local/bin/noded
COPY --from=builder /app/bin/aldea /usr/local/bin/aldea

WORKDIR /data
EXPOSE 8080 9001 9002 9003 9004 9005

CMD ["trackerd"]
