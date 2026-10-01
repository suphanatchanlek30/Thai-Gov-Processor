# bimg (image resize/compress) uses cgo bindings to libvips, so the build
# stage needs the vips headers/pkg-config files; the runtime stage only
# needs the shared libraries the compiled binary links against.
FROM golang:1.24.13-bookworm AS build

RUN apt-get update && apt-get install -y --no-install-recommends \
    libvips-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    libvips42 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --uid 10001 app

WORKDIR /app
COPY --from=build /out/api ./api

USER app
EXPOSE 8080
ENTRYPOINT ["./api"]
