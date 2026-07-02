FROM golang:1.26 AS build
WORKDIR /src
COPY backend/ ./backend/
WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata wget \
    && rm -rf /var/lib/apt/lists/*
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=build /out/server /app/server
COPY deploy/config.docker.yml /app/config/config.yml
COPY backend/data/sql.sql /app/data/sql.sql
RUN mkdir -p /app/data /app/qdrant /app/surrealdb /app/logs
ENV SKIP_ENGINE_LAUNCH=1 \
    SKIP_SIDECAR_LAUNCH=1
EXPOSE 8899
CMD ["/app/server"]
