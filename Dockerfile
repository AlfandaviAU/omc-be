# Build stage
FROM golang:alpine AS builder

# Install gcc and musl-dev for CGO (required by go-sqlite3)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -a -o omc-be .

# Final stage
FROM alpine:latest

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/omc-be .

# Default ENV vars (these can be overridden by fly.toml)
ENV PORT=8081
ENV DB_PATH=/data/omc.db
ENV GIN_MODE=release

EXPOSE 8081

CMD ["./omc-be"]
