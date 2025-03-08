FROM golang:1.23-alpine AS builder

RUN apk add --no-cache bash curl git ca-certificates build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o scrape404 .

FROM mcr.microsoft.com/playwright:v1.50.0-jammy

RUN useradd -m appuser

RUN apt-get update && apt-get install -y curl && apt-get clean

WORKDIR /home/appuser

COPY --from=builder /app/scrape404 /usr/local/bin/

RUN chmod +x /usr/local/bin/scrape404

USER appuser

EXPOSE 8080

ENTRYPOINT ["scrape404"]