package p2p

import (
    "encoding/gob"
    "io"
)

// Decoder decodes messages from a reader
type Decoder interface {
    Decode(io.Reader, *RPC) error
}

// GOBDecoder uses gob encoding
type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
    return gob.NewDecoder(r).Decode(msg)
}

// DefaultDecoder handles simple message formats
type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
    peekBuf := make([]byte, 1)
    if _, err := r.Read(peekBuf); err != nil {
        return nil
    }

    // Check if this is a stream
    stream := peekBuf[0] == IncomingStream
    if stream {
        msg.Stream = true
        return nil
    }

    // Read regular message
    buf := make([]byte, 1028)
    n, err := r.Read(buf)
    if err != nil {
        return err
    }

    msg.Payload = buf[:n]

    return nil
}