package main

import (
    "crypto/sha1"
    "encoding/hex"
    "errors"
    "fmt"
    "io"
    "log"
    "os"
    "strings"
)

const defaultRootFolderName = "p2pnetwork" // Default storage directory

// CASPathTransformFunc creates a content-addressable storage path
func CASPathTransformFunc(key string) PathKey {
    hash := sha1.Sum([]byte(key))           // SHA1 hash of key
    hashStr := hex.EncodeToString(hash[:]) // Hex encoded hash

    blocksize := 5
    sliceLen := len(hashStr) / blocksize
    paths := make([]string, sliceLen)

    // Split hash into blocks for directory structure
    for i := 0; i < sliceLen; i++ {
        from, to := i*blocksize, (i*blocksize)+blocksize
        paths[i] = hashStr[from:to]
    }

    return PathKey{
        PathName: strings.Join(paths, "/"), // Directory path
        Filename: hashStr,                 // Filename
    }
}

// PathTransformFunc defines how to transform keys to storage paths
type PathTransformFunc func(string) PathKey

// PathKey represents a storage path and filename
type PathKey struct {
    PathName string // Directory path components
    Filename string // Final filename
}

// FirstPathName returns the first component of the path
func (p PathKey) FirstPathName() string {
    paths := strings.Split(p.PathName, "/")
    if len(paths) == 0 {
        return ""
    }
    return paths[0]
}

// FullPath returns the complete path including filename
func (p PathKey) FullPath() string {
    return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

// StoreOpts contains storage configuration
type StoreOpts struct {
    Root              string            // Root storage directory
    PathTransformFunc PathTransformFunc // Function to transform keys to paths
}

// DefaultPathTransformFunc is a simple path transform that uses the key directly
var DefaultPathTransformFunc = func(key string) PathKey {
    return PathKey{
        PathName: key,
        Filename: key,
    }
}

// Store manages file storage operations
type Store struct {
    StoreOpts
}

// NewStore creates a new Store instance
func NewStore(opts StoreOpts) *Store {
    if opts.PathTransformFunc == nil {
        opts.PathTransformFunc = DefaultPathTransformFunc
    }
    if len(opts.Root) == 0 {
        opts.Root = defaultRootFolderName
    }

    return &Store{
        StoreOpts: opts,
    }
}

// Has checks if a file exists for the given ID and key
func (s *Store) Has(id string, key string) bool {
    pathKey := s.PathTransformFunc(key)
    fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

    _, err := os.Stat(fullPathWithRoot)
    return !errors.Is(err, os.ErrNotExist)
}

// Clear removes all stored files
func (s *Store) Clear() error {
    return os.RemoveAll(s.Root)
}

// Delete removes a file by ID and key
func (s *Store) Delete(id string, key string) error {
    pathKey := s.PathTransformFunc(key)

    defer func() {
        log.Printf("deleted [%s] from disk", pathKey.Filename)
    }()

    firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())

    return os.RemoveAll(firstPathNameWithRoot)
}

// Write stores data for the given ID and key
func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
    return s.writeStream(id, key, r)
}

// WriteDecrypt writes and decrypts data using the provided encryption key
func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
    f, err := s.openFileForWriting(id, key)
    if err != nil {
        return 0, err
    }
    n, err := copyDecrypt(encKey, r, f)
    return int64(n), err
}

// openFileForWriting prepares a file for writing
func (s Store) openFileForWriting(id string, key string) (*os.File, error) {
    pathKey := s.PathTransformFunc(key)
    pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)
    if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
        return nil, err
    }

    fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

    return os.Create(fullPathWithRoot)
}

// writeStream handles the actual file writing
func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
    f, err := s.openFileForWriting(id, key)
    if err != nil {
        return 0, err
    }
    return io.Copy(f, r)
}

// Read retrieves a file by ID and key
func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
    return s.readStream(id, key)
}

// readStream handles the actual file reading
func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error) {
    pathKey := s.PathTransformFunc(key)
    fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

    file, err := os.Open(fullPathWithRoot)
    if err != nil {
        return 0, nil, err
    }

    fi, err := file.Stat()
    if err != nil {
        return 0, nil, err
    }

    return fi.Size(), file, nil
}