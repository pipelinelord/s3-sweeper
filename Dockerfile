# STAGE 1: Build (Keep this exactly the same)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o s3-sweeper main.go

# STAGE 2: Run (The "Distroless" Way)
# "static" is perfect for Go binaries. "nonroot" runs as a safe user by default.
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from the builder
COPY --from=builder /app/s3-sweeper .

# Use the binary as the entrypoint
ENTRYPOINT ["/s3-sweeper"]