package decryptonite

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"crypto/aes"
	"crypto/cipher"

	"golang.org/x/crypto/scrypt"
    "golang.org/x/term"
)

type Secrets map[string]string

func GetTokenFromSecrets() (string, error) {
	secretsFile := os.Getenv("HOME") + "/.secrets.json"
f, err := os.Open(secretsFile)
if err != nil {
	return "", err
}
defer func() {
	if cerr := f.Close(); cerr != nil {
		fmt.Println("Warning: failed to close file:", cerr)
	}
}()
secrets := Secrets{}
if err := json.NewDecoder(f).Decode(&secrets); err != nil {
	return "", fmt.Errorf("error decoding secrets file: %w", err)
}

	fmt.Print("Enter token name to decrypt (e.g. github, aws, stripe): ")
	tokenName, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	tokenName = strings.TrimSpace(tokenName)

	encryptedToken, ok := secrets[tokenName]
	if !ok {
		return "", fmt.Errorf("token %q not found in secrets", tokenName)
	}

	fmt.Print("Enter passphrase to decrypt token: ")
	passBytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	passphrase := strings.TrimSpace(string(passBytes))

	token, err := Decrypt(encryptedToken, passphrase)
	if err != nil {
		return "", err
	}

	// Confirm before displaying
	fmt.Printf("Are you sure you want to display the decrypted token for %q? (y/n): ", tokenName)
	confirm, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		return "", fmt.Errorf("user aborted display of token")
	}

	return token, nil
}

func Decrypt(encryptedB64, passphrase string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", err
	}

	if len(data) < 16+12 {
		return "", errors.New("encrypted data too short")
	}

	salt := data[:16]
	nonce := data[16 : 16+12]
	ciphertext := data[16+12:]

	key, err := scrypt.Key([]byte(passphrase), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}