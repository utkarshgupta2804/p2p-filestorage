
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


## ⚙️ Requirements

- [Go](https://go.dev/) **1.21+**

---

## 🏃‍♂️ How to Run

1️⃣ **Clone the repository**

```bash
git clone https://github.com/your-username/p2p-filestorage.git
cd p2p-filestorage

go run . 
```

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
