use thiserror::Error;

#[derive(Debug, Error)]
pub enum EbpfError {
    #[error("Failed to load eBPF program: {0}")]
    LoadError(String),
    
    #[error("Failed to attach XDP program to interface {interface}: {message}")]
    AttachError { interface: String, message: String },
    
    #[error("eBPF map '{map_name}' not found")]
    MapNotFound { map_name: String },
    
    #[error("Failed to access eBPF map '{map_name}': {message}")]
    MapAccessError { map_name: String, message: String },
    
    #[error("Failed to sync data to Redis: {0}")]
    RedisSyncError(String),
    
    #[error("Failed to sync data from Redis: {0}")]
    RedisFetchError(String),
    
    #[error("eBPF program detached or not attached")]
    NotAttached,
    
    #[error("eBPF operation timeout: {0}")]
    Timeout(String),
    
    #[error("Invalid eBPF data: {0}")]
    InvalidData(String),
}

impl EbpfError {
    pub fn is_recoverable(&self) -> bool {
        match self {
            EbpfError::MapAccessError { .. } => true,
            EbpfError::RedisSyncError(_) => true,
            EbpfError::RedisFetchError(_) => true,
            EbpfError::Timeout(_) => true,
            _ => false,
        }
    }
}
