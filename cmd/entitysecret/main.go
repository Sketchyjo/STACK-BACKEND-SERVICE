package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// IMPORTANT: This code implements idiomatic Go and Clean Architecture best practices for sensitive key handling and encryption.
// - Separation between key parsing and encryption logic
// - All errors wrapped and handled explicitly
// - Inputs and outputs validated
// - Modular, testable, and maintainable for future expansion

const (
	publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAxQsczKCXuMCgyGYff2tZ
xR+ZUW8MBvgwmbFkGTmyoenSC6X/5o5BPPkPZTIZs/oC8ouOdAKijOYsUP3+qdc+
mzjx2lIHnQN1TtNQ2Vm93Hk+G6vEFHDsYsb0nchk+7V5Pbki3ynOnfsV6LRbaFCf
cgTGxHSSmKbnItW3qAiVluPPoPBx4WbQNyeS5TREv0R1NC1U311rxLGbxl+bjb73
fFzlvSkGe2UyPs8tJnAYhqpvFOQv1SdXDvGbfwM5lBfqjCGMlkHkYYwsgLYl4R/R
x01ncZvYjgYwXAungJMRpD9aUBSt8f4pDDlUxoXq294y7hCSi6aNGoDPqDyAaqoN
2rSYbswGZmCz5ivJLHZNFP9qCwoKeL1l9+VlDrKs+nhRmrhCoXG0OOUdTbpkU4Ff
oUjh4SKR8YPq7TfSGyBe9q5VAF7bEici1FkH9I7+wf41YSq47dU3UOryjbF34fXZ
dQJ9xBEk1thTDUK8ZmIY8SQwqolSQIAKxsxOf2XoNdk3PiaXJHDTtfEiTtZFybKR
rWFG4h0GeRPLCy52KAe+nfJmpODKeGmrGgvlA0IVeHDpqv7WNsG/o3G4JBL3odWs
6qKoMrDhL1W/32EMPObdtUPTtAyTO3HxfXWsUavJ5KLHApoiwDx9Vn7aW5ytBvAV
6aAk60U2+xWaJJqFlWAx6a8CAwEAAQ==
-----END PUBLIC KEY-----`
	hexEncodedEntitySecret = "dcd90b5d7bfd4f17222283d14ac0e2ce0d814df1d4f030a37065868113437fdc"
)

func main() {
	// Decode the static hex-encoded 32-byte secret
	entitySecret, err := hex.DecodeString(hexEncodedEntitySecret)
	if err != nil {
		exitWithError(fmt.Errorf("failed to decode entity secret: %w", err))
	}
	if len(entitySecret) != 32 {
		exitWithError(errors.New("invalid entity secret length; must be 32 bytes"))
	}

	pubKey, err := parseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		exitWithError(fmt.Errorf("failed to parse RSA public key: %w", err))
	}

	ciphertext, err := encryptOAEP(pubKey, entitySecret)
	if err != nil {
		exitWithError(fmt.Errorf("encryption failed: %w", err))
	}

	fmt.Printf("print cyphertext %x\n", ciphertext)
	fmt.Printf("Hex encoded entity secret: %x\n", entitySecret)
	fmt.Printf("Entity secret ciphertext (base64): %s\n", base64.StdEncoding.EncodeToString(ciphertext))
}

// parseRSAPublicKeyFromPEM parses an RSA public key from PEM format.
func parseRSAPublicKeyFromPEM(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse public key DER: %w", err)
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key type parsed is not RSA")
	}
	return rsaPub, nil
}

// encryptOAEP performs RSA-OAEP encryption using SHA-256.
func encryptOAEP(pubKey *rsa.PublicKey, message []byte) ([]byte, error) {
	random := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), random, pubKey, message, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa.EncryptOAEP failed: %w", err)
	}
	return ciphertext, nil
}

// exitWithError prints the error and exits the program with exit code 1.
func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
