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

yaml
Copy
Edit

---

## ⚙️ Requirements

- [Go](https://go.dev/) **1.21+**

---

## 🏃‍♂️ How to Run

1️⃣ **Clone the repository**

```bash
git clone https://github.com/your-username/p2p-file-storage.git
cd p2p-file-storage
2️⃣ Run peers on different ports

Open multiple terminals and run:

bash
Copy
Edit
# Terminal 1
go run . --port 3000

# Terminal 2
go run . --port 7000

# Terminal 3 (connects to peers)
go run . --port 5000 --bootstrap 3000,7000
⚡️ Command-line Flags
Flag	Description	Example
--port	Port for this peer’s TCP server	--port 5000
--bootstrap	Comma-separated list of other peer ports	--bootstrap 3000,7000

✅ Example Output
plaintext
Copy
Edit
[:3000] starting fileserver...
2025/06/16 00:17:54 TCP transport listening on port: :3000

[:7000] starting fileserver...
2025/06/16 00:17:54 TCP transport listening on port: :7000

[:5000] starting fileserver...
2025/06/16 00:17:56 TCP transport listening on port: :5000
[:5000] attempting to connect with remote :7000
[:5000] attempting to connect with remote :3000
2025/06/16 00:17:56 connected with remote 127.0.0.1:3000
2025/06/16 00:17:56 connected with remote 127.0.0.1:7000
