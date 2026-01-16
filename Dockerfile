# Stage 1: Build the Vue.js frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Install pnpm
RUN corepack enable && corepack prepare pnpm@latest --activate

# Copy package files first for better caching
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy frontend source code
COPY frontend/ ./

# Build the frontend
RUN pnpm build

# Stage 2: Build the Go backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install git (required for some Go modules)
RUN apk add --no-cache git

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the built frontend from the previous stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o vachancms .

# Stage 3: Final minimal image
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS requests (GitHub API)
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=backend-builder /app/vachancms .

# Copy the frontend dist (Go server serves this in production)
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

# Expose the port
EXPOSE 8080

# Set production mode
ENV PRODUCTION=true
ENV GIN_MODE=release
ENV PORT=8080

# Run the application
CMD ["./vachancms"]
