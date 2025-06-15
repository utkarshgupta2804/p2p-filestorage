
# 📁 P2P File Storage System

A simple **Peer-to-Peer (P2P) file storage system** written in **Go**.  
This project demonstrates a minimal decentralized file sharing network where peers store, share, and retrieve file chunks over TCP connections using content-addressable storage.

---

## 🚀 Features

- ✅ TCP transport for peer communication  
- ✅ Bootstrapped peer discovery  
- ✅ Content-addressable chunk storage  
- ✅ Automatic peer connections  
- ✅ Simple CLI for starting multiple peers  

---

## 📂 Project Structure

.
├── cmd/ # Main entry point (main.go)
├── internal/
│ ├── transport/ # TCP transport implementation
│ ├── storage/ # Content-addressable storage logic
│ ├── protocol/ # Protocol handlers for file transfer
│ └── config.go # Node configuration
├── go.mod
├── go.sum
└── README.md
