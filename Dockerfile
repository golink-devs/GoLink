# Dockerfile
FROM alpine:3.19 AS builder

RUN apk add --no-cache \
    go gcc g++ cmake ninja zip unzip curl git \
    pkgconfig perl nasm build-base bash

WORKDIR /build

# Copy and install libdave first
COPY scripts/ scripts/
RUN chmod +x scripts/libdave_install.sh && \
    # Assuming the script works or we bypass it for now as a mock
    # VCPKG_FORCE_SYSTEM_BINARIES=1 \
    # CC=/usr/bin/gcc \
    # CXX=/usr/bin/g++ \
    # CXXFLAGS="-Wno-error=maybe-uninitialized" \
    # FORCE_BUILD=1 \
    ./scripts/libdave_install.sh v1.1.0

# Build GoLink binary
COPY . .
RUN CGO_ENABLED=1 GOOS=linux \
    go build -ldflags="-w -s" -o golink ./cmd/golink

# Runtime image
FROM alpine:3.19

RUN apk add --no-cache \
    ffmpeg \
    yt-dlp \
    ca-certificates \
    libstdc++ \
    libgcc

# Copy golink binary
COPY --from=builder /build/golink /usr/local/bin/golink

# Copy libdave shared library (if it was installed)
# COPY --from=builder /usr/local/lib/libdave* /usr/local/lib/
# RUN ldconfig /usr/local/lib || true

# Default config location
COPY config.yml /etc/golink/config.yml

EXPOSE 2333

WORKDIR /golink
CMD ["golink", "--config", "/etc/golink/config.yml"]
