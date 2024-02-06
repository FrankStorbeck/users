package users

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
)

// testKey tests if the key has a correct length. If not, it returns an error.
func testKey(key []byte) (err error) {
	switch l := len(key); l {
	case 16, 24, 32:
	default:
		err = fmt.Errorf("key has a length of %d bytes, should be 16, 24 or 32", l)
	}
	return
}

// en encrypts a string with key and then encodes it into a
// string using a base32 format. The key must have a length of
// 16, 24, or 32 bytes
func en(s string, key []byte) (string, error) {
	if err := testKey(key); err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// the initial vector `iv` is put in front of the cipher text
	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewOFB(block, iv)
	cr := make([]byte, len(s))
	stream.XORKeyStream(cr, []byte(s))
	cr = append(iv, cr...)

	return base32.StdEncoding.EncodeToString(cr), nil
}

// de returns the decrypted string from a base32 encoded and
// with key encrypted string. The key must have a length of
// 16, 24, or 32 bytes.
func de(s string, key []byte) (string, error) {
	if err := testKey(key); err != nil {
		return "", err
	}

	cr, err := base32.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	if len(cr)-aes.BlockSize < 0 {
		return "", errors.New("wrong initial vector for decryption")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plain := make([]byte, len(cr)-aes.BlockSize)
	stream := cipher.NewOFB(block, cr[:aes.BlockSize])
	stream.XORKeyStream(plain, cr[aes.BlockSize:])

	return string(plain), nil
}
