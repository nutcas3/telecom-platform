# Packet Gateway with eBPF

A high-performance packet filtering gateway using eBPF XDP programs for real-time traffic control and credit management.

## Architecture

The packet gateway consists of two main components:

1. **eBPF Kernel Program** (`src/ebpf/packet_filter.c`): Runs in kernel space, filters packets at line rate
2. **Userspace Control Plane** (`src/main.rs`): Manages eBPF maps, syncs with Redis, handles configuration

## Features

- **XDP Packet Filtering**: Kernel-space packet filtering for maximum performance
- **Credit Management**: Real-time credit deduction and enforcement
- **Dynamic Blocking**: Block/unblock users via eBPF maps
- **Redis Integration**: Sync stats and configuration with Redis
- **Graceful Shutdown**: Clean up resources and sync final state

## eBPF Maps

### packet_stats
- **Key**: `u32` (IP address)
- **Value**: `u64` (byte count)
- **Purpose**: Track traffic usage per IP

### user_credits
- **Key**: `u32` (IP address)  
- **Value**: `i64` (credit balance)
- **Purpose**: Enforce credit limits, block when depleted

### block_list
- **Key**: `u32` (IP address)
- **Value**: `u8` (block flag: 0=allow, 1=block)
- **Purpose**: Manual blocking of specific IPs

## Build Requirements

### Prerequisites
```bash
# Install eBPF development tools
sudo apt-get install -y linux-headers-$(uname -r) clang llvm

# Install Rust eBPF tools
cargo install aya-cli
```

### Build Process
```bash
# Build eBPF program
cargo xtask build-ebpf

# Build userspace application
cargo build --release
```

## Usage

### Basic Usage
```bash
# Run with default settings
sudo ./target/release/packet-gateway

# Specify interface and Redis
sudo ./target/release/packet-gateway \
  --interface eth0 \
  --redis-url redis://127.0.0.1/ \
  --sync-interval 1
```

### Redis Integration

The gateway syncs with Redis using these key patterns:

**Packet Stats:**
```
packet_stats:192.168.1.100 -> 1048576  # bytes transferred
```

**User Credits:**
```
user_credit:192.168.1.100 -> 10485760  # remaining credit in bytes
```

**Blocked Users:**
```
blocked_user:192.168.1.100 -> 1  # blocked flag
```

### Managing Credits

```bash
# Set user credit (Redis)
redis-cli set user_credit:192.168.1.100 10485760

# Block a user
redis-cli set blocked_user:192.168.1.100 1

# Unblock a user  
redis-cli del blocked_user:192.168.1.100
```

## eBPF Program Details

### Packet Processing Flow

1. **Parse Packet**: Extract source/destination IPs
2. **Check Block List**: Block if IP is in block list
3. **Check Credit**: Block if insufficient credit
4. **Update Stats**: Increment packet statistics
5. **Deduct Credit**: Decrease user credit balance
6. **Return Action**: `XDP_PASS` or `XDP_DROP`

### Key Functions

- `packet_filter()`: Main XDP program entry point
- `should_block_packet()`: Determine if packet should be blocked
- `update_packet_stats()`: Update traffic statistics
- `update_credit_usage()`: Deduct credit from user balance

## Performance

- **Line Rate Processing**: XDP enables kernel-space packet filtering
- **Map Operations**: O(1) hash map lookups for credit/block checks
- **Atomic Updates**: Lock-free map updates for high concurrency

## Monitoring

### Metrics Available
- Packets processed per second
- Bytes transferred per IP
- Active user count
- Low credit warnings
- Block list size

### Log Levels
```bash
# Debug logging
RUST_LOG=debug ./packet-gateway

# Production logging
RUST_LOG=info ./packet-gateway
```

## Security Considerations

- **Root Required**: eBPF XDP programs require root privileges
- **Network Isolation**: Test in isolated network environments
- **Memory Safety**: eBPF verifier ensures kernel safety
- **Resource Limits**: Map sizes prevent memory exhaustion

## Troubleshooting

### Common Issues

**Permission Denied:**
```bash
# Run with sudo
sudo ./packet-gateway
```

**Interface Not Found:**
```bash
# Check available interfaces
ip link show
```

**Redis Connection Failed:**
```bash
# Verify Redis is running
redis-cli ping
```

**eBPF Program Failed to Load:**
```bash
# Check kernel logs
dmesg | grep -i ebpf
```

### Debug Mode

Enable detailed logging to troubleshoot issues:
```bash
RUST_LOG=debug ./packet-gateway --interface eth0
```

## Development

### Adding New Features

1. **eBPF Side**: Add new maps or modify packet_filter.c
2. **Userspace Side**: Add corresponding methods in EbpfManager
3. **Redis Integration**: Update sync functions
4. **Testing**: Add unit tests and integration tests

### Testing

```bash
# Run unit tests
cargo test

# Run integration tests (requires test environment)
cargo test --test integration

# Build and run in debug mode
cargo build && sudo ./target/debug/packet-gateway
```

## License

This project is licensed under the GPL v2 license, compatible with eBPF kernel requirements.
