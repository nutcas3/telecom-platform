use aya::{programs::Xdp, Bpf, maps::HashMap};
use aya_log::BpfLogger;
use anyhow::{Context, Result};
use tracing::{info, warn, error, debug};
use redis::AsyncCommands;

mod ip_utils;
mod redis_sync;
use ip_utils::{u32_to_ipv4_string, ipv4_string_to_u32};
use redis_sync::RedisSyncer;

pub struct EbpfManager {
    bpf: Bpf,
    interface: String,
}

#[derive(Debug, Clone)]
pub struct PacketStats {
    pub ip: u32,
    pub bytes: u64,
}

#[derive(Debug, Clone, Copy)]
pub struct SyncEntry {
    pub ip: u32,
    pub bytes: u64,
    pub credit: i64,
    pub blocked: u8,
    pub valid: u8,
}

#[derive(Debug, Clone)]
pub struct CreditInfo {
    pub ip: u32,
    pub credit: i64,
}

impl EbpfManager {
    pub async fn new(interface: String) -> Result<Self> {
        let mut bpf = Bpf::load(include_bytes_aligned!("packet_filter"))?;
        
        if let Err(e) = BpfLogger::init(&mut bpf) {
            warn!("Failed to initialize eBPF logger: {}", e);
        }
        
        info!("Loaded eBPF program successfully");
        
        Ok(Self { bpf, interface })
    }
    
    pub fn attach(&mut self) -> Result<()> {
        let program: &mut Xdp = self.bpf.program_mut("packet_filter")
            .ok_or_else(|| anyhow::anyhow!("Program 'packet_filter' not found"))?
            .try_into()?;
        
        program.load()?;
        info!("Loaded XDP program into kernel");
        
        program.attach(&self.interface, aya::programs::XdpFlags::default())
            .context("Failed to attach XDP program")?;
        
        info!("Attached XDP program to interface: {}", self.interface);
        Ok(())
    }
    
    pub fn get_packet_stats(&self) -> Result<Vec<PacketStats>> {
        let packet_stats_map: &HashMap<_, u32, u64> = self.bpf.map("packet_stats")
            .ok_or_else(|| anyhow::anyhow!("packet_stats map not found"))?
            .try_into()?;
        
        let mut stats = Vec::new();
        
        for (ip, bytes) in packet_stats_map.iter()? {
            stats.push(PacketStats { ip: *ip, bytes: *bytes });
        }
        
        Ok(stats)
    }
    
    pub fn get_credit_info(&self) -> Result<Vec<CreditInfo>> {
        let user_credits_map: &HashMap<_, u32, i64> = self.bpf.map("user_credits")
            .ok_or_else(|| anyhow::anyhow!("user_credits map not found"))?
            .try_into()?;
        
        let mut credits = Vec::new();
        
        for (ip, credit) in user_credits_map.iter()? {
            credits.push(CreditInfo { ip: *ip, credit: *credit });
        }
        
        Ok(credits)
    }
    
    pub fn update_user_credit(&self, ip: u32, credit: i64) -> Result<()> {
        let user_credits_map: &HashMap<_, u32, i64> = self.bpf.map("user_credits")
            .ok_or_else(|| anyhow::anyhow!("user_credits map not found"))?
            .try_into()?;
        
        user_credits_map.insert(&ip, &credit, 0)?;
        debug!("Updated credit for IP {} to {}", ip, credit);
        
        Ok(())
    }
    
    pub fn block_user(&self, ip: u32, blocked: bool) -> Result<()> {
        let block_list_map: &HashMap<_, u32, u8> = self.bpf.map("block_list")
            .ok_or_else(|| anyhow::anyhow!("block_list map not found"))?
            .try_into()?;
        
        let block_flag: u8 = if blocked { 1 } else { 0 };
        block_list_map.insert(&ip, &block_flag, 0)?;
        
        info!("{} IP {} in eBPF block list", if blocked { "Blocked" } else { "Unblocked" }, ip);
        Ok(())
    }
    
    pub async fn sync_to_redis(&self, redis_conn: &mut redis::aio::MultiplexedConnection) -> Result<()> {
        let stats = self.get_packet_stats()?;
        let stats_vec: Vec<(u32, u64)> = stats.iter().map(|s| (s.ip, s.bytes)).collect();
        RedisSyncer::sync_packet_stats(&stats_vec, redis_conn).await?;
        
        let credits = self.get_credit_info();
        if let Ok(credits) = credits {
            let credits_vec: Vec<(u32, i64)> = credits.iter().map(|c| (c.ip, c.credit)).collect();
            RedisSyncer::sync_credit_info(&credits_vec, redis_conn).await?;
        }
        
        debug!("Synced eBPF maps to Redis");
        Ok(())
    }
    
    pub async fn sync_from_redis(&self, redis_conn: &mut redis::aio::MultiplexedConnection) -> Result<()> {
        RedisSyncer::sync_credits_from_redis(redis_conn, |ip, credit| {
            self.update_user_credit(ip, credit)
        }).await?;
        
        RedisSyncer::sync_blocked_from_redis(redis_conn, |ip| {
            self.block_user(ip, true)
        }).await?;
        
        debug!("Synced data from Redis to eBPF maps");
        Ok(())
    }
    
    pub fn trigger_sync(&self) -> Result<()> {
        let sync_control_map: &HashMap<_, u32, u64> = self.bpf.map("sync_control")
            .ok_or_else(|| anyhow::anyhow!("sync_control map not found"))?
            .try_into()?;
        
        let key = 0u32;
        let sync_flag = 1u64;
        sync_control_map.insert(&key, &sync_flag, 0)?;
        
        debug!("Triggered eBPF sync function");
        Ok(())
    }
    
    pub fn read_sync_buffer(&self) -> Result<Vec<SyncEntry>> {
        let sync_buffer_map: &HashMap<_, u32, SyncEntry> = self.bpf.map("sync_buffer")
            .ok_or_else(|| anyhow::anyhow!("sync_buffer map not found"))?
            .try_into()?;
        
        let mut entries = Vec::new();
        
        let key = 0u32;
        if let Some(entry) = sync_buffer_map.get(&key, 0)? {
            if entry.valid == 1 {
                entries.push(*entry);
            }
        }
        
        Ok(entries)
    }
    
    pub fn sync_batch_to_redis(&self, redis_conn: &mut redis::aio::MultiplexedConnection) -> Result<()> {
        self.trigger_sync()?;
        std::thread::sleep(std::time::Duration::from_millis(10));
        
        let entries = self.read_sync_buffer()?;
        
        for entry in entries {
            RedisSyncer::sync_batch_entry(
                entry.ip,
                entry.bytes,
                entry.credit,
                entry.blocked,
                redis_conn,
            ).await?;
        }
        
        info!("Batch synced {} entries to Redis", entries.len());
        Ok(())
    }
    
    pub fn cleanup_maps(&self) -> Result<()> {
        if let Ok(packet_stats_map) = self.bpf.map("packet_stats").and_then(|m| m.try_into()) {
            let map: &HashMap<_, u32, u64> = packet_stats_map;
            let keys: Vec<u32> = map.iter()?.map(|(k, _)| *k).collect();
            for key in keys {
                let _: Result<(), _> = map.delete(&key);
            }
        }
        
        if let Ok(sync_control_map) = self.bpf.map("sync_control").and_then(|m| m.try_into()) {
            let map: &HashMap<_, u32, u64> = sync_control_map;
            let key = 0u32;
            let reset_flag = 0u64;
            let _: Result<(), _> = map.insert(&key, &reset_flag, 0);
        }
        
        info!("Cleared eBPF maps");
        Ok(())
    }
}

fn include_bytes_aligned(s: &str) -> &'static [u8] {
    aya::include_bytes_aligned!(s)
}
