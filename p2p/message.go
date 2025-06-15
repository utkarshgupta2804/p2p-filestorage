package p2p

const (
    IncomingMessage = 0x1 // Regular message
    IncomingStream  = 0x2 // Stream message
)

// RPC represents a remote procedure call
type RPC struct {
    From    string // Sender address
    Payload []byte // Message content
    Stream  bool   // Whether this is a stream
}