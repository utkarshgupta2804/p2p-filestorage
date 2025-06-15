package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/utkarshgupta2804/p2p-filestorage/p2p"
)

// FileServerOpts contains configuration options for FileServer
type FileServerOpts struct {
	ID                string            // Server identifier
	EncKey            []byte            // Encryption key
	StorageRoot       string            // Root storage directory
	PathTransformFunc PathTransformFunc // Path transformation function
	Transport         p2p.Transport     // Network transport
	BootstrapNodes    []string          // Initial nodes to connect to
}

// FileServer implements the P2P file storage server
type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex          // Protects peers map
	peers    map[string]p2p.Peer // Connected peers

	store  *Store        // Storage backend
	quitch chan struct{} // Channel for graceful shutdown
}

// NewFileServer creates a new FileServer instance
func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	if len(opts.ID) == 0 {
		opts.ID = generateID() // Generate unique ID if not provided
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

// broadcast sends a message to all connected peers
func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage}) // Message type indicator
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// Message represents a network message
type Message struct {
	Payload any // Can be any serializable type
}

// MessageStoreFile contains file storage information
type MessageStoreFile struct {
	ID   string // File owner ID
	Key  string // File key
	Size int64  // File size
}

// MessageGetFile contains file retrieval information
type MessageGetFile struct {
	ID  string // File owner ID
	Key string // File key
}

// Get retrieves a file by key, either locally or from the network
func (s *FileServer) Get(key string) (io.Reader, error) {
	// Check local storage first
	if s.store.Has(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}

	fmt.Printf("[%s] dont have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)

	// Create and broadcast get request
	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: hashKey(key),
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	time.Sleep(time.Millisecond * 500) // Wait for responses

	// Receive file from peers
	for _, peer := range s.peers {
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		// Write and decrypt the received file
		n, err := s.store.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] received (%d) bytes over the network from (%s)", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	// Return the now locally stored file
	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

// Store saves a file locally and propagates it to the network
func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer) // Tee reader to write and keep data
	)

	// Store locally first
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}

	// Broadcast storage info to network
	msg := Message{
		Payload: MessageStoreFile{
			ID:   s.ID,
			Key:  hashKey(key),
			Size: size + 16, // Account for IV in encrypted data
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 5)

	// Stream file to all peers
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream}) // Stream type indicator
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

	return nil
}

// Stop shuts down the file server
func (s *FileServer) Stop() {
	close(s.quitch)
}

// OnPeer handles new peer connections
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

// loop is the main event loop for the file server
func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}

		case <-s.quitch:
			return
		}
	}
}

// handleMessage processes incoming messages
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

// handleMessageGetFile processes file retrieval requests
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	if rc, ok := r.(io.ReadCloser); ok {
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	// Send stream indicator and file size
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

// handleMessageStoreFile processes file storage notifications
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	// Write the incoming file data
	n, err := s.store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Addr(), n)

	peer.CloseStream()

	return nil
}

// bootstrapNetwork connects to initial nodes
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

// Start begins the file server operation
func (s *FileServer) Start() error {
	fmt.Printf("[%s] starting fileserver...\n", s.Transport.Addr())

	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork() // Connect to initial nodes

	s.loop() // Start main event loop

	return nil
}

func init() {
	// Register message types for gob encoding
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
