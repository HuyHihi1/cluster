# Step 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies before bringing source over (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Step 2: Create small runtime image
FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
