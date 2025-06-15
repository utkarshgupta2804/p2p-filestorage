package p2p

import "net"

// Peer represents a remote node in the network
type Peer interface {
    net.Conn          // Embedded connection interface
    Send([]byte) error // Send data to peer
    CloseStream()     // Close an active stream
}

// Transport handles communication between nodes
type Transport interface {
    Addr() string            // Listening address
    Dial(string) error       // Connect to remote address
    ListenAndAccept() error  // Start listening
    Consume() <-chan RPC     // Channel for incoming messages
    Close() error            // Shutdown transport
}