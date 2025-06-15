package main

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "crypto/rand"
    "encoding/hex"
    "io"
)

// generateID creates a random 32-byte identifier
func generateID() string {
    buf := make([]byte, 32)
    io.ReadFull(rand.Reader, buf)
    return hex.EncodeToString(buf)
}

// hashKey creates an MD5 hash of a key
func hashKey(key string) string {
    hash := md5.Sum([]byte(key))
    return hex.EncodeToString(hash[:])
}

// newEncryptionKey generates a random 32-byte encryption key
func newEncryptionKey() []byte {
    keyBuf := make([]byte, 32)
    io.ReadFull(rand.Reader, keyBuf)
    return keyBuf
}

// copyStream handles the actual encryption/decryption stream copying
func copyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
    var (
        buf = make([]byte, 32*1024) // 32KB buffer
        nw  = blockSize
    )
    for {
        n, err := src.Read(buf)
        if n > 0 {
            stream.XORKeyStream(buf, buf[:n]) // Encrypt/decrypt
            nn, err := dst.Write(buf[:n])
            if err != nil {
                return 0, err
            }
            nw += nn
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return 0, err
        }
    }
    return nw, nil
}

// copyDecrypt decrypts data from src to dst using the provided key
func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
    block, err := aes.NewCipher(key) // AES cipher
    if err != nil {
        return 0, err
    }

    // Read initialization vector
    iv := make([]byte, block.BlockSize())
    if _, err := src.Read(iv); err != nil {
        return 0, err
    }

    stream := cipher.NewCTR(block, iv) // CTR mode stream
    return copyStream(stream, block.BlockSize(), src, dst)
}

// copyEncrypt encrypts data from src to dst using the provided key
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
    block, err := aes.NewCipher(key) // AES cipher
    if err != nil {
        return 0, err
    }

    // Generate random initialization vector
    iv := make([]byte, block.BlockSize()) // 16 bytes
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return 0, err
    }

    // Write IV first
    if _, err := dst.Write(iv); err != nil {
        return 0, err
    }

    stream := cipher.NewCTR(block, iv) // CTR mode stream
    return copyStream(stream, block.BlockSize(), src, dst)
}