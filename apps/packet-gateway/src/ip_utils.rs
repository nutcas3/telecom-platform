use anyhow::{Context, Result};
use std::net::Ipv4Addr;

/// Converts a u32 IP address to an IPv4 string
pub fn u32_to_ipv4_string(ip: u32) -> String {
    let ip_addr = Ipv4Addr::from(ip);
    ip_addr.to_string()
}

/// Converts an IPv4 string to a u32 IP address
pub fn ipv4_string_to_u32(ip_str: &str) -> Result<u32> {
    let ip_addr: Ipv4Addr = ip_str.parse()
        .context("Invalid IP address format")?;
    Ok(u32::from(ip_addr))
}
