# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install Node.js for frontend build
RUN apk add --no-cache nodejs npm git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build frontend
ARG VITE_STRIPE_PUBLISHABLE_KEY
ENV VITE_STRIPE_PUBLISHABLE_KEY=${VITE_STRIPE_PUBLISHABLE_KEY}
RUN cd frontend && npm ci && npm run build

# Build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o azadi ./cmd/server

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S appuser && adduser -S appuser -G appuser

WORKDIR /app

COPY --from=builder /app/azadi .
COPY --from=builder /app/frontend/dist ./frontend/dist
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/seed ./seed

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

ENTRYPOINT ["./azadi"]
