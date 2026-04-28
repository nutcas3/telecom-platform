FROM rust:1.95-slim as builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    clang \
    llvm \
    libelf-dev \
    libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy Cargo files
COPY Cargo.toml Cargo.lock ./
COPY apps/packet-gateway ./apps/packet-gateway
COPY apps/charging-engine/src/circuit_breaker.rs ./apps/charging-engine/src/circuit_breaker.rs

# Build the packet-gateway binary
RUN cargo build --release --bin packet-gateway

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libssl3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from builder
COPY --from=builder /app/target/release/packet-gateway /app/packet-gateway

# Set environment variables
ENV RUST_LOG=info

# Run the binary
CMD ["/app/packet-gateway"]
