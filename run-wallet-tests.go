// package main

// import (
// 	"fmt"
// 	"strings"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/users"

// 	payload := strings.NewReader("{\"userId\":\"123e4567-e89b-12d3-a456-426614174001\"}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"fmt"
// 	"strings"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/users"

// 	payload := strings.NewReader("{\"userId\":\"124e4567-e89b-12d3-a456-426614174001\"}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"fmt"
// 	"strings"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/user/initialize"

// 	payload := strings.NewReader("{\"idempotencyKey\":\"123a4567-e89b-12d3-a456-426614174001\",\"blockchains\":[\"MATIC-AMOY\"]}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")
// 	req.Header.Add("X-User-Token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRoTW9kZSI6IlBJTiIsImRldmVsb3BlckVudGl0eUVudmlyb25tZW50IjoiVEVTVCIsImVudGl0eUlkIjoiZGU5NWNiMTMtM2IzMC00YTJlLWJiYTEtMTBmYmI0YjBjYTkyIiwiZXhwIjoxNzYwOTQ2NjYxLCJpYXQiOjE3NjA5NDMwNjEsImludGVybmFsVXNlcklkIjoiZTNmNmY3OTMtN2ZhOS01ZDMxLTk2ZjYtODdmMzc0MzYxYjdmIiwiaXNzIjoiaHR0cHM6Ly9wcm9ncmFtbWFibGUtd2FsbGV0LmNpcmNsZS5jb20iLCJqdGkiOiIyMDk5OGZhOS1mNTBhLTRiMDMtYWMxMy1hZDg2ZGIzMjkwN2IiLCJzdWIiOiIxMjNlNDU2Ny1lODliLTEyZDMtYTQ1Ni00MjY2MTQxNzQwMDEifQ.QZ0aaFzRJ6dOtYSgVx4MK12MwtJpMBBMgnuPq6086_WDrGY6TLuFknwc4p4SIt1pukVF0Eg0NEn9JYL8N_IsUS8CGAKHK6MgALZmL3VwHZ-QXazNC1fztfhWdoSeQ1em8WNVp8FCxCOOWYRVu7_29Le9S9Nmsf-k4lMyeITe_M716bQWDA2r3sUogf7vGBh6Og2xjdfdWVp_pLMuROEp764sBJmNOc8bc4wyrueF1L-O2m7sA7m9Oi_G2LCiBJyKB1vZWaEBdnCrzarpUfZwkm5NbfLu2596HQ3HZ9ty-MbZmlPaKob3Wqs8dQVxHvut1wnl44KacQ4dwNxJVOO-zg")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"fmt"
// 	"strings"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/users/token"

// 	payload := strings.NewReader("{\"userId\":\"123e4567-e89b-12d3-a456-426614174001\"}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"fmt"
// 	"strings"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/user/wallets"

// 	payload := strings.NewReader("{\"idempotencyKey\":\"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11\",\"blockchains\":[\"MATIC-AMOY\"]}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")
// 	req.Header.Add("X-User-Token", "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRoTW9kZSI6IlBJTiIsImRldmVsb3BlckVudGl0eUVudmlyb25tZW50IjoiVEVTVCIsImVudGl0eUlkIjoiZGU5NWNiMTMtM2IzMC00YTJlLWJiYTEtMTBmYmI0YjBjYTkyIiwiZXhwIjoxNzYwOTQ2NjYxLCJpYXQiOjE3NjA5NDMwNjEsImludGVybmFsVXNlcklkIjoiZTNmNmY3OTMtN2ZhOS01ZDMxLTk2ZjYtODdmMzc0MzYxYjdmIiwiaXNzIjoiaHR0cHM6Ly9wcm9ncmFtbWFibGUtd2FsbGV0LmNpcmNsZS5jb20iLCJqdGkiOiIyMDk5OGZhOS1mNTBhLTRiMDMtYWMxMy1hZDg2ZGIzMjkwN2IiLCJzdWIiOiIxMjNlNDU2Ny1lODliLTEyZDMtYTQ1Ni00MjY2MTQxNzQwMDEifQ.QZ0aaFzRJ6dOtYSgVx4MK12MwtJpMBBMgnuPq6086_WDrGY6TLuFknwc4p4SIt1pukVF0Eg0NEn9JYL8N_IsUS8CGAKHK6MgALZmL3VwHZ-QXazNC1fztfhWdoSeQ1em8WNVp8FCxCOOWYRVu7_29Le9S9Nmsf-k4lMyeITe_M716bQWDA2r3sUogf7vGBh6Og2xjdfdWVp_pLMuROEp764sBJmNOc8bc4wyrueF1L-O2m7sA7m9Oi_G2LCiBJyKB1vZWaEBdnCrzarpUfZwkm5NbfLu2596HQ3HZ9ty-MbZmlPaKob3Wqs8dQVxHvut1wnl44KacQ4dwNxJVOO-zg")
// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"crypto/rand"
// 	"fmt"
// 	"io"
// )

// // generateRandomHex generates a new 32-byte cryptographically secure random secret.
// // Returns the secret as a byte slice or panics on error.
// func generateRandomHex() []byte {
// 	mainBuff := make([]byte, 32)
// 	_, err := io.ReadFull(rand.Reader, mainBuff)
// 	if err != nil {
// 		panic("reading from crypto/rand failed: " + err.Error())
// 	}
// 	return mainBuff
// }

// // The following sample codes generate a distinct hex encoded entity secret with each execution
// // The generation of entity secret only need to be executed once unless you need to rotate entity secret.
// func main() {
// 	entitySecret := generateRandomHex()
// 	fmt.Printf("Hex encoded entity secret: %x\n", entitySecret)
// }

// package main

// import (
// 	"fmt"
// 	"net/http"
// 	"io"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/config/entity/publicKey"

// 	req, _ := http.NewRequest("GET", url, nil)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

// package main

// import (
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"crypto/sha256"
// 	"crypto/x509"
// 	"encoding/base64"
// 	"encoding/hex"
// 	"encoding/pem"
// 	"errors"
// 	"fmt"
// 	"os"
// )

// // IMPORTANT: This code implements idiomatic Go and Clean Architecture best practices for sensitive key handling and encryption.
// // - Separation between key parsing and encryption logic
// // - All errors wrapped and handled explicitly
// // - Inputs and outputs validated
// // - Modular, testable, and maintainable for future expansion

// const (
// 	publicKeyPEM = `-----BEGIN PUBLIC KEY-----
// MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAxQsczKCXuMCgyGYff2tZ
// xR+ZUW8MBvgwmbFkGTmyoenSC6X/5o5BPPkPZTIZs/oC8ouOdAKijOYsUP3+qdc+
// mzjx2lIHnQN1TtNQ2Vm93Hk+G6vEFHDsYsb0nchk+7V5Pbki3ynOnfsV6LRbaFCf
// cgTGxHSSmKbnItW3qAiVluPPoPBx4WbQNyeS5TREv0R1NC1U311rxLGbxl+bjb73
// fFzlvSkGe2UyPs8tJnAYhqpvFOQv1SdXDvGbfwM5lBfqjCGMlkHkYYwsgLYl4R/R
// x01ncZvYjgYwXAungJMRpD9aUBSt8f4pDDlUxoXq294y7hCSi6aNGoDPqDyAaqoN
// 2rSYbswGZmCz5ivJLHZNFP9qCwoKeL1l9+VlDrKs+nhRmrhCoXG0OOUdTbpkU4Ff
// oUjh4SKR8YPq7TfSGyBe9q5VAF7bEici1FkH9I7+wf41YSq47dU3UOryjbF34fXZ
// dQJ9xBEk1thTDUK8ZmIY8SQwqolSQIAKxsxOf2XoNdk3PiaXJHDTtfEiTtZFybKR
// rWFG4h0GeRPLCy52KAe+nfJmpODKeGmrGgvlA0IVeHDpqv7WNsG/o3G4JBL3odWs
// 6qKoMrDhL1W/32EMPObdtUPTtAyTO3HxfXWsUavJ5KLHApoiwDx9Vn7aW5ytBvAV
// 6aAk60U2+xWaJJqFlWAx6a8CAwEAAQ==
// -----END PUBLIC KEY-----`
// 	hexEncodedEntitySecret = "dcd90b5d7bfd4f17222283d14ac0e2ce0d814df1d4f030a37065868113437fdc"
// )

// func main() {
// 	// Decode the static hex-encoded 32-byte secret
// 	entitySecret, err := hex.DecodeString(hexEncodedEntitySecret)
// 	if err != nil {
// 		exitWithError(fmt.Errorf("failed to decode entity secret: %w", err))
// 	}
// 	if len(entitySecret) != 32 {
// 		exitWithError(errors.New("invalid entity secret length; must be 32 bytes"))
// 	}

// 	pubKey, err := parseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
// 	if err != nil {
// 		exitWithError(fmt.Errorf("failed to parse RSA public key: %w", err))
// 	}

// 	ciphertext, err := encryptOAEP(pubKey, entitySecret)
// 	if err != nil {
// 		exitWithError(fmt.Errorf("encryption failed: %w", err))
// 	}

// 	fmt.Printf("Hex encoded entity secret: %x\n", entitySecret)
// 	fmt.Printf("Entity secret ciphertext (base64): %s\n", base64.StdEncoding.EncodeToString(ciphertext))
// }

// // parseRSAPublicKeyFromPEM parses an RSA public key from PEM format.
// func parseRSAPublicKeyFromPEM(pubPEM []byte) (*rsa.PublicKey, error) {
// 	block, _ := pem.Decode(pubPEM)
// 	if block == nil {
// 		return nil, errors.New("failed to parse PEM block containing the key")
// 	}
// 	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to parse public key DER: %w", err)
// 	}
// 	rsaPub, ok := pub.(*rsa.PublicKey)
// 	if !ok {
// 		return nil, errors.New("key type parsed is not RSA")
// 	}
// 	return rsaPub, nil
// }

// // encryptOAEP performs RSA-OAEP encryption using SHA-256.
// func encryptOAEP(pubKey *rsa.PublicKey, message []byte) ([]byte, error) {
// 	random := rand.Reader
// 	ciphertext, err := rsa.EncryptOAEP(sha256.New(), random, pubKey, message, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("rsa.EncryptOAEP failed: %w", err)
// 	}
// 	return ciphertext, nil
// }

// // exitWithError prints the error and exits the program with exit code 1.
// func exitWithError(err error) {
// 	fmt.Fprintln(os.Stderr, "Error:", err)
// 	os.Exit(1)
// }

// package main

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"
// )

// func main() {

// 	url := "https://api.circle.com/v1/w3s/developer/walletSets"

// 	payload := strings.NewReader("{\"idempotencyKey\":\"323e4567-e89b-12d2-a556-426614174001\",\"entitySecretCipherText\":\"dLkBWoeDbT7zVMw0ajZw9vNsS0g2i+val6N4LMu0jhaBSff+bmC+N7wocI/HZEqJKaWbCUDZ/Ri1jr1MPUCKpWYvtzK3U//k9gL4Up2C+7mm8+voQP4a4pY0HtLBQJbJkLRqVJGchkLl0m0KjAFVosJ+WoMIzuy8FJ0JFMgRmoP9HYM+9/U2+45/Thwj1t408rGXDMlJEiHVkHVqJf6nWfTLQMCaTd2YEATZHapeJqiu/5XCK8o1uy4TEuF+hUBPcSBzdOV3y1hYS/GTAOo84z9KuDvN9QZ4x48t5/Z0kc6pT+oU3lq2pXwD2hyEIGRo39Hihr4i7AIA6HXQpgn08D8yJwpTAvYVQXZJgQ1GnD4IoLie5w2Pnuqcesb1Br2bPok46aPXyS7xvCpUGH6T0vorHzbRBK4/zVckpJD7hV3y5C4GmzMfsabxXBXriS5M138N3r4aBdMp2/VdYIhjIVQAi9QbqhFPpzdXzAqImoRaoAV1dV0pQtduwq0qoabKLZ6Tr9/wZsO5Wdce/yTSVqziPorhnGQrXjHttU9n33GNGSejViU9IsDseb70YB8KhO7jzFWVs9WbWwW+9WP9P+VAQTsiSnK78x2r55AXRKHY+N7spjSsdPNcYPA0gVMR61sUpjgqwWvDG+hduvDcZ2e5s+OeQObktYXt0GYAZuM=\",\"name\":\"stacky\"}")

// 	req, _ := http.NewRequest("POST", url, payload)

// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, _ := io.ReadAll(res.Body)

// 	fmt.Println(string(body))

// }

package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {

	url := "https://api.circle.com/v1/w3s/developer/wallets"

	payload := strings.NewReader("{\"idempotencyKey\":\"923e4567-e89b-12d2-a656-426614174001\",\"entitySecretCipherText\":\"oy+hSuc+/mr+lLRKiUNv5+byqr4Q3n+jJUfO3h0KIYLPhFebcUc+9MmSXVPvRSZZhYgkZdfvLP8kaXuUNafZgI+qekxE4VMtAcPnpSFc12aXy78ar0vVVHXeP3dpARpalQbwCmOSBqjvA/5j2Q/YLBwg3h0gAVLte29lH9LLj4jaSks1/lhvDhIMdolbQa8g9hRX5UvEwgIX+dE6fryWaaSWJWeEw1seEzqunjhDBAH6Za06vJrvOTZYLSptcmil+2KdpRc5DXbh9uC9+hJTlFoICniz9MOek1qOinJUN71y/PlZmSzoi+0UeUX6KNXhtZIa/3LwAHgBPY2buQGLBr0DDFjQl4RcaKLGiBa7lJPI4QuVKeQCdq0SEqE9zODDkFG31IgKdWMHTbF3l4bpGC4fnzXxxWy4xmML300gPJzdXdNWzKxi8o8l42t2BaDT+obXJGfKU6pPlGQR0mzd/JAeflxTAr4d4VwFukkejQSFjiWB7V9VeGKn8Cbj4Tm/o2rrShbaL3xURmWqXE1d8J0QL/7K2Usc8TJKw6qRZ4GsAjCk91WzSAfGW2on9rl+hn9T4HoejPRruIKfaXKoKZp+YYtEZF8S/654VUyZvxi5/q4oU/KlExNKbBuFLRwbQrVgxDmGOspDsYx8HBPUY04q47GVjN8vKoBOHyv6df4=\",\"blockchains\":[\"MATIC-AMOY\"],\"count\":1,\"accountType\":\"EOA\",\"walletSetId\":\"e2618c1c-dda5-560a-b4f8-d91b1e64efe6\"}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer TEST_API_KEY:0558a886ad0f03f57767ab3bd998ae76:a62be17299c9398a615928e9053e03cb")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

}
