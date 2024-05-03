# Use a smaller base image for the builder stage
FROM golang:1.21.3 as builder

# Set the working directory
WORKDIR /app

# Copy only the necessary dependency files
COPY go.mod go.sum ./

# Download dependencies in a separate layer to leverage Docker cache
RUN go mod download

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the source code files
COPY ./ ./

# Build the binary with minimal footprint
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scanner .

# Final stage: Use the official Alpine image for a smaller footprint
FROM alpine:3.16

# Install ca-certificates and create non-root user in a single layer to reduce layer count
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app

# Set user context
USER app

# Copy the built binary from the builder stage
COPY --from=builder /app/scanner /app/scanner

# Expose port and set environment variables
EXPOSE 8080
ENV CLAMD_HOST=localhost \
    CLAMD_PORT=3310 \
    LISTEN_PORT=8080

# Set the entrypoint to the scanner application
ENTRYPOINT ["/app/scanner"]
