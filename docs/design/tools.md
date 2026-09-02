# Tools

## Resources
- Ubuntu images for VMware/VirtualBox.
- Interactive prompts in Go (e.g. `c-bata/go-prompt`).

## References
- VS Code remote editing over SSH.

## Requirements

### Node.js installation
```
wget https://nodejs.org/dist/v10.13.0/node-v10.13.0-linux-x64.tar.xz
mkdir -p /opt/node && tar xf node-v10.13.0-linux-x64.tar.xz -C /opt/node --strip-components=1
```

### libpcap (for packet capture)
```
add-apt-repository ppa:ubuntu-toolchain-r/test
apt-get update
apt-get install g++-4.9 libpcap0.8-dev
export CC="gcc-4.9"; export CXX="gcc-4.9"
npm install pcap
```
