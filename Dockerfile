# ==============================================================================
# Stage 1: Frontend Builder (Vite + React + TypeScript + Tailwind)
# ==============================================================================
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# ==============================================================================
# Stage 2: Go Backend Builder
# ==============================================================================
FROM golang:1.26-alpine AS go-builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /app/bin/server ./cmd/server

# ==============================================================================
# Target: api (Backend Go Service)
# ==============================================================================
FROM alpine:3.21 AS api

RUN apk --no-cache add ca-certificates curl tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=go-builder /app/bin/server /usr/local/bin/server

RUN mkdir -p /data && chown -R appuser:appgroup /data /app

USER appuser
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

ENTRYPOINT ["/usr/local/bin/server"]

# ==============================================================================
# Target: nginx (Edge Reverse Proxy & Static Host)
# ==============================================================================
FROM nginx:1.27-alpine AS nginx

COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html
RUN mkdir -p /srv/static

COPY nginx/nginx.conf.template /etc/nginx/templates/default.conf.template

EXPOSE 80 443
