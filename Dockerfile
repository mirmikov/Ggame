FROM node:20-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.22-alpine AS backend-builder
WORKDIR /src/backend
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/prometheus-battle ./cmd/server

FROM alpine:3.21
RUN addgroup -S game \
    && adduser -S -G game game
WORKDIR /app
COPY --from=backend-builder /out/prometheus-battle /app/prometheus-battle
COPY --from=frontend-builder /src/frontend/dist /app/public
USER game
ENV PORT=8080 STATIC_DIR=/app/public
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/health >/dev/null || exit 1
ENTRYPOINT ["/app/prometheus-battle"]
