package spartan_go

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
)

// encoding == "" to use default hex encoding
func Hash(o interface{}, encoding string) string {
	if len(encoding) == 0 {
		encoding = "hex"
	}
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", o)))
	bytes := hash.Sum(nil)
	if encoding == "base64" {
		return base64.StdEncoding.EncodeToString(bytes)
	} else if encoding == "hex" {
		return hex.EncodeToString(bytes)
	} else {
		return ""
	}
}

// equivalent to generateKeypair() because private key object contains public key
func GenerateKey() *rsa.PrivateKey {
	keypair, _ := rsa.GenerateKey(rand.Reader, 2048)
	return keypair
}

func Sign(privKey *rsa.PrivateKey, msg string) (string, error) {
	msgHash := sha256.New()
	msgHash.Write([]byte(msg))
	hashedMsg := msgHash.Sum(nil)
	sig, err := rsa.SignPSS(rand.Reader, privKey, crypto.SHA256, hashedMsg, nil)
	if err != nil {
		log.Println("Could not sign message!")
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func VerifySignature(pubKey rsa.PublicKey, msg string, sig string) bool {
	msgHash := sha256.New()
	msgHash.Write([]byte(msg))
	hashedMsg := msgHash.Sum(nil)
	bytesSig, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	err = rsa.VerifyPSS(&pubKey, crypto.SHA256, hashedMsg, bytesSig, nil)
	if err != nil {
		return false
	}
	return true
}

func CalcAddress(pubKey rsa.PublicKey) string {
	return Hash(pubKey, "base64")
}

func AddressMatchesKey(addr string, pubKey rsa.PublicKey) bool {
	return addr == CalcAddress(pubKey)
}
