use std::process::Command;

fn main() {
    // Try to build the eBPF program, but skip if cargo-build-bpf is not installed
    let status = Command::new("cargo")
        .args(&["build-bpf", "--release"])
        .status();

    match status {
        Ok(status) if status.success() => {
            // eBPF build succeeded
        }
        Ok(_) => {
            // eBPF build failed, but continue without it
            println!("cargo:warning=Failed to build eBPF program, continuing without it");
        }
        Err(_) => {
            // cargo-build-bpf not installed, skip eBPF build
            println!("cargo:warning=cargo-build-bpf not installed, skipping eBPF build");
        }
    }
}
