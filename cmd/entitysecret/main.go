package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func generateRandomHex() []byte {
	mainBuff := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

// fetchPublicKeyFromCircle fetches the public key from Circle's API
func fetchPublicKeyFromCircle() (string, error) {
	req, err := http.NewRequest("GET", circleAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", "Bearer "+getCircleAPIKey())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var response CirclePublicKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return response.Data.PublicKey, nil
}

// Circle API configuration
var circleAPIURL = "https://api.circle.com/v1/w3s/config/entity/publicKey"

// getCircleAPIKey returns the API key from environment variable or uses the default
func getCircleAPIKey() string {
	if apiKey := os.Getenv("CIRCLE_API_KEY"); apiKey != "" {
		return apiKey
	}
	// Fallback to default test key
	return "TEST_API_KEY:23387683755f3f749392263317fd3968:9f76b6e4170b99acae094ebcdc1886af"
}

// Circle API response structure
type CirclePublicKeyResponse struct {
	Data struct {
		PublicKey string `json:"publicKey"`
	} `json:"data"`
}

// The following sample codes generate a distinct entity secret and encrypt it in one execution
func main() {
	// Fetch public key from Circle API
	fmt.Println("Fetching public key from Circle API...")
	publicKeyString, err := fetchPublicKeyFromCircle()
	if err != nil {
		panic(fmt.Errorf("failed to fetch public key: %w", err))
	}
	fmt.Println("Public key fetched successfully!")

	// Generate a new entity secret (32 bytes = 64 hex characters)
	entitySecret := generateRandomHex()

	// Parse the public key
	pubKey, err := ParseRsaPublicKeyFromPem([]byte(publicKeyString))
	if err != nil {
		panic(err)
	}

	// Encrypt the entity secret
	cipher, err := EncryptOAEP(pubKey, entitySecret)
	if err != nil {
		panic(err)
	}

	// Output both the hex encoded secret and the encrypted ciphertext
	fmt.Printf("\n=== Entity Secret Generation Complete ===\n")
	fmt.Printf("Hex encoded entity secret: %x\n", entitySecret)
	fmt.Printf("Entity secret ciphertext: %s\n", base64.StdEncoding.EncodeToString(cipher))
	fmt.Printf("\nNote: You can set CIRCLE_API_KEY environment variable to use a different API key.\n")
}

// ParseRsaPublicKeyFromPem parse rsa public key from pem.
func ParseRsaPublicKeyFromPem(pubPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
	}
	return nil, errors.New("key type is not rsa")
}

// EncryptOAEP rsa encrypt oaep.
func EncryptOAEP(pubKey *rsa.PublicKey, message []byte) (ciphertext []byte, err error) {
	random := rand.Reader
	ciphertext, err = rsa.EncryptOAEP(sha256.New(), random, pubKey, message, nil)
	if err != nil {
		return nil, err
	}
	return
}
