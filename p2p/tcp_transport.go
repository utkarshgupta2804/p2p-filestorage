package p2p

import (
    "errors"
    "fmt"
    "log"
    "net"
    "sync"
)

// TCPPeer represents a remote node over a TCP connection
type TCPPeer struct {
    net.Conn       // Embedded net.Conn interface
    outbound bool  // True if we dialed the connection, false if we accepted it
    wg       *sync.WaitGroup // WaitGroup for stream synchronization
}

// NewTCPPeer creates a new TCPPeer instance
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
    return &TCPPeer{
        Conn:     conn,
        outbound: outbound,
        wg:       &sync.WaitGroup{},
    }
}

// CloseStream signals the end of a stream
func (p *TCPPeer) CloseStream() {
    p.wg.Done()
}

// Send writes data to the peer connection
func (p *TCPPeer) Send(b []byte) error {
    _, err := p.Conn.Write(b)
    return err
}

// TCPTransportOpts contains configuration options for TCPTransport
type TCPTransportOpts struct {
    ListenAddr    string        // Address to listen on
    HandshakeFunc HandshakeFunc // Function to perform handshake
    Decoder       Decoder       // Message decoder
    OnPeer        func(Peer) error // Callback when new peer connects
}

// TCPTransport implements the Transport interface using TCP
type TCPTransport struct {
    TCPTransportOpts            // Embedded options
    listener      net.Listener  // TCP listener
    rpcch         chan RPC     // Channel for incoming RPC messages
}

// NewTCPTransport creates a new TCPTransport instance
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
    return &TCPTransport{
        TCPTransportOpts: opts,
        rpcch:            make(chan RPC, 1024), // Buffered channel for RPCs
    }
}

// Addr returns the listen address
func (t *TCPTransport) Addr() string {
    return t.ListenAddr
}

// Consume returns a read-only channel for incoming RPC messages
func (t *TCPTransport) Consume() <-chan RPC {
    return t.rpcch
}

// Close shuts down the transport
func (t *TCPTransport) Close() error {
    return t.listener.Close()
}

// Dial connects to a remote peer
func (t *TCPTransport) Dial(addr string) error {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return err
    }

    go t.handleConn(conn, true) // Handle outbound connection

    return nil
}

// ListenAndAccept starts listening for incoming connections
func (t *TCPTransport) ListenAndAccept() error {
    var err error

    t.listener, err = net.Listen("tcp", t.ListenAddr)
    if err != nil {
        return err
    }

    go t.startAcceptLoop() // Start accepting connections in a goroutine

    log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)

    return nil
}

// startAcceptLoop continuously accepts new connections
func (t *TCPTransport) startAcceptLoop() {
    for {
        conn, err := t.listener.Accept()
        if errors.Is(err, net.ErrClosed) {
            return
        }

        if err != nil {
            fmt.Printf("TCP accept error: %s\n", err)
        }

        go t.handleConn(conn, false) // Handle inbound connection
    }
}

// handleConn manages an established connection
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
    var err error

    defer func() {
        fmt.Printf("dropping peer connection: %s", err)
        conn.Close()
    }()

    peer := NewTCPPeer(conn, outbound)

    // Perform handshake
    if err = t.HandshakeFunc(peer); err != nil {
        return
    }

    // Notify about new peer
    if t.OnPeer != nil {
        if err = t.OnPeer(peer); err != nil {
            return
        }
    }

    // Read loop for incoming messages
    for {
        rpc := RPC{}
        err = t.Decoder.Decode(conn, &rpc)
        if err != nil {
            return
        }

        rpc.From = conn.RemoteAddr().String() // Set message source

        if rpc.Stream {
            // Handle streaming data
            peer.wg.Add(1)
            fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
            peer.wg.Wait()
            fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
            continue
        }

        t.rpcch <- rpc // Send RPC to consumer channel
    }
}