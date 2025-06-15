package p2p

// HandshakeFunc performs a handshake with a peer
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc is a no-operation handshake
func NOPHandshakeFunc(Peer) error { return nil }