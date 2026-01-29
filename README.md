<!-- requirement:
- ubuntu 22.04
- go 1.20
- tcconfig (sudo pip install tcconfig)
- mininet-wifi 
        git clone https://github.com/intrig-unicamp/mininet-wifi
        cd mininet-wifi
        mininet-wifi$ sudo util/install.sh -Wlnfv6

serverdrl:
- flask (pip install Flask)
- pytorch (pip3 install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu118) -->


# SAC-MPQUIC simulation framework

This repository provides a research-oriented simulation framework for **Multipath QUIC (MP-QUIC)** scheduling based on Deep Reinforcement Learning (Soft Actor-Critic – SAC).  
The framework is designed to evaluate intelligent path scheduling strategies in heterogeneous wireless networks using **Mininet-WiFi**, **MP-QUIC**, and **PyTorch**.

---

## Repository Structure

```
.
├── aes12/                    # Utility hashing functions
├── fnv128a/                  # Utility hashing functions
├── quic-clients/             # QUIC client implementations
├── quic-go/                  # Modified quic-go with MPQUIC support
├── quic-go-certificates/     # Certificates for QUIC/HTTP3
├── serverdrl/                # DRL-based scheduler
├── simulation/               # Mininet-WiFi simulation scenarios
├── simulationFileServer/     # File server for download experiments
```

---

## System Overview

The framework consists of three main components:

1. **Network Emulator**
   - Built on **Mininet-WiFi**
   - Supports heterogeneous paths (e.g., Wi-Fi, LTE/5G-like links)
   - Allows control over bandwidth, delay, loss, and mobility

2. **MPQUIC Stack**
   - Based on a modified version of `quic-go`
   - Supports multiple concurrent paths
   - Enables custom scheduling decisions from an external controller

3. **DRL Scheduler (SAC)**
   - Implemented in **PyTorch**
   - Learns optimal path selection policies
   - Interacts with the network via runtime feedback (RTT, CWND, loss, throughput)

---

## Requirements

### Operating System
- Ubuntu 22.04 LTS

### Core Dependencies
- Go ≥ 1.20  
- Python ≥ 3.8

### Network Emulation

Install **Mininet-WiFi**:
```bash
git clone https://github.com/intrig-unicamp/mininet-wifi
cd mininet-wifi
sudo util/install.sh -Wlnfv6
```

Install **tcconfig**:
```bash
sudo pip install tcconfig
```

---

## Python Dependencies (DRL Scheduler)

Install required Python packages:
```bash
pip install Flask
pip3 install torch torchvision torchaudio \
  --index-url https://download.pytorch.org/whl/cu118
```
---

## Key Features

- MPQUIC multi-path transmission
- Deep Reinforcement Learning-based scheduling (Soft Actor-Critic)
- Configurable wireless heterogeneity
- Support for file download and streaming scenarios
- Reproducible and research-oriented experiments

---

## Notes

- This repository is intended for research and experimental use only
- Some components are legacy implementations retained for reference
