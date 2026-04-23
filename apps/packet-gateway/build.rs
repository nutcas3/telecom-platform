use std::process::Command;

fn main() {
    // Build the eBPF program
    let status = Command::new("cargo")
        .args(&["build-bpf", "--release"])
        .status()
        .expect("Failed to execute cargo build-bpf");

    if !status.success() {
        panic!("Failed to build eBPF program");
    }
}
