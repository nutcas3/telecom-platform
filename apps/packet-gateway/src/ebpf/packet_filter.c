#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/udp.h>
#include <linux/tcp.h>
#include <linux/icmp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u32);      // IP address
    __type(value, __u64);    // Byte count
} packet_stats SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u32);      // IP address
    __type(value, __s64);    // Credit balance (can be negative)
} user_credits SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10240);
    __type(key, __u32);      // IP address
    __type(value, __u8);     // Block flag (0=allow, 1=block)
} block_list SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} sync_interval SEC(".maps");

#define BLOCK_THRESHOLD 0
#define SYNC_INTERVAL_NS 1000000000ULL 

static __always_inline __u32 get_packet_size(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    return data_end - data;
}

static __always_inline __u32 get_src_ip(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return 0;
        
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return 0;
        
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return 0;
        
    return bpf_ntohl(ip->saddr);
}

static __always_inline __u32 get_dst_ip(struct xdp_md *ctx) {
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return 0;
        
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return 0;
        
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return 0;
        
    return bpf_ntohl(ip->daddr);
}

static __always_inline bool should_block_packet(__u32 src_ip, __u32 dst_ip) {
    __u8 *block_flag_src = bpf_map_lookup_elem(&block_list, &src_ip);
    __u8 *block_flag_dst = bpf_map_lookup_elem(&block_list, &dst_ip);
    
    if (block_flag_src && *block_flag_src == 1)
        return true;
    if (block_flag_dst && *block_flag_dst == 1)
        return true;
        
    __s64 *credit_src = bpf_map_lookup_elem(&user_credits, &src_ip);
    if (credit_src && *credit_src <= BLOCK_THRESHOLD)
        return true;
        
    return false;
}

static __always_inline void update_packet_stats(__u32 ip, __u32 bytes) {
    __u64 *current_stats = bpf_map_lookup_elem(&packet_stats, &ip);
    __u64 new_stats = bytes;
    
    if (current_stats) {
        new_stats += *current_stats;
    }
    
    bpf_map_update_elem(&packet_stats, &ip, &new_stats, BPF_ANY);
}

static __always_inline void update_credit_usage(__u32 ip, __u32 bytes) {
    __s64 *current_credit = bpf_map_lookup_elem(&user_credits, &ip);
    if (current_credit) {
        __s64 new_credit = *current_credit - bytes;
        bpf_map_update_elem(&user_credits, &ip, &new_credit, BPF_ANY);
    }
}

SEC("xdp")
int packet_filter(struct xdp_md *ctx) {
    __u32 packet_size = get_packet_size(ctx);
    if (packet_size < sizeof(struct ethhdr) + sizeof(struct iphdr))
        return XDP_PASS;
        
    __u32 src_ip = get_src_ip(ctx);
    __u32 dst_ip = get_dst_ip(ctx);
    
    if (src_ip == 0 || dst_ip == 0)
        return XDP_PASS;
        
    if (should_block_packet(src_ip, dst_ip)) {
        bpf_printk("Blocking packet from %x to %x - insufficient credit or blocked\n", 
                   src_ip, dst_ip);
        return XDP_DROP;
    }
    
    update_packet_stats(src_ip, packet_size);
    update_packet_stats(dst_ip, packet_size);
    
    update_credit_usage(src_ip, packet_size);
    
    return XDP_PASS;
}

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} sync_control SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1024);
    __type(key, __u32);
    __type(value, struct sync_entry);
} sync_buffer SEC(".maps");

struct sync_entry {
    __u32 ip;
    __u64 bytes;
    __s64 credit;
    __u8 blocked;
    __u8 valid;
};

SEC("xdp")
int sync_stats(struct xdp_md *ctx) {
    // Get sync control flag
    __u32 key = 0;
    __u64 *sync_flag = bpf_map_lookup_elem(&sync_control, &key);
    if (!sync_flag || *sync_flag == 0) {
        return XDP_PASS;
    }
    
    // Reset sync flag
    __u64 reset_flag = 0;
    bpf_map_update_elem(&sync_control, &key, &reset_flag, BPF_ANY);
    
    // Batch sync packet stats
    __u32 buffer_idx = 0;
    struct sync_entry *entry = bpf_map_lookup_elem(&sync_buffer, &buffer_idx);
    if (!entry) {
        return XDP_PASS;
    }
    
    // Iterate through packet_stats map and batch into buffer
    __u32 stats_key = 0;
    __u64 *bytes = bpf_map_lookup_elem(&packet_stats, &stats_key);
    if (bytes) {
        entry->ip = stats_key;
        entry->bytes = *bytes;
        entry->valid = 1;
        
        // Get corresponding credit info
        __s64 *credit = bpf_map_lookup_elem(&user_credits, &stats_key);
        if (credit) {
            entry->credit = *credit;
        }
        
        // Get block status
        __u8 *blocked = bpf_map_lookup_elem(&block_list, &stats_key);
        if (blocked) {
            entry->blocked = *blocked;
        }
        
        // Signal to userspace that data is ready
        bpf_printk("Synced stats for IP %x: %lu bytes, credit %ld, blocked %d\n", 
                   stats_key, *bytes, credit ? *credit : 0, blocked ? *blocked : 0);
    }
    
    return XDP_PASS;
}

// License
char _license[] SEC("license") = "GPL";
