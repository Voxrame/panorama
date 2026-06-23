FROM docker.io/oven/bun:1 AS ui_builder
WORKDIR /app
COPY frontend/package.json frontend/bun.lock ./
RUN bun install
COPY frontend/ .
RUN bun run build-only

FROM docker.io/golang:1.26-alpine AS backend_builder
WORKDIR /app
RUN apk add git
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY --from=ui_builder /app/dist ./frontend/dist
COPY frontend/static.go ./frontend/
COPY cmd ./cmd
COPY internal ./internal
RUN go build -v ./cmd/panorama

FROM scratch
WORKDIR /app
COPY --from=backend_builder /app/panorama ./
COPY config.example.toml /etc/panorama/config.toml

ENTRYPOINT ["./panorama", "run", "--config", "/etc/panorama/config.toml"]
