FROM golang:1.23.4-alpine

# Install git and SSL ca certificates for private repos
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

# Create appuser and required directories
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/scraper

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Use appuser
USER appuser:appuser

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/scraper"]