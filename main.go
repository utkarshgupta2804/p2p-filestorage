package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/utkarshgupta2804/p2p-filestorage/p2p"
)

// sanitizeAddr removes invalid characters from address for use as directory name
func sanitizeAddr(addr string) string {
	// Remove colon and replace with underscore for Windows compatibility
	return strings.ReplaceAll(addr, ":", "")
}

// makeServer creates and configures a FileServer instance for P2P file sharing.
// listenAddr: The address this server will listen on
// nodes: Bootstrap nodes to connect to initially
func makeServer(listenAddr string, nodes ...string) *FileServer {
	// Configure TCP transport options
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,           // Address to listen on
		HandshakeFunc: p2p.NOPHandshakeFunc, // No-op handshake function
		Decoder:       p2p.DefaultDecoder{}, // Default message decoder
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	// Configure FileServer options
	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),                           // Generate encryption key for secure transfers
		StorageRoot:       sanitizeAddr(listenAddr) + "_network",        // Storage directory based on listen address (Windows-safe)
		PathTransformFunc: CASPathTransformFunc,                         // Content-addressable storage path function
		Transport:         tcpTransport,                                 // Network transport layer
		BootstrapNodes:    nodes,                                        // Initial nodes to connect to
	}

	// Create a new FileServer instance
	s := NewFileServer(fileServerOpts)

	// Assign OnPeer callback to handle new peer connections
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	// Create three FileServer instances:
	// s1 - First node with no bootstrap nodes
	// s2 - Second node with no bootstrap nodes
	// s3 - Third node that connects to s1 and s2
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", "")
	s3 := makeServer(":5000", ":3000", ":7000")

	// Start s1 and s2 concurrently
	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond) // Wait for s1 to start

	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(2 * time.Second) // Wait for s2 to start

	// Start s3 which will connect to s1 and s2
	go s3.Start()
	time.Sleep(2 * time.Second) // Wait for s3 to connect

	// Perform file operations on s3
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!")) // Simulated file data

		// Store the file in the network
		s3.Store(key, data)

		// Delete the file locally (but it remains in the network)
		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		// Retrieve the file from the network
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		// Read the retrieved file
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b)) // Print the file contents
	}
}