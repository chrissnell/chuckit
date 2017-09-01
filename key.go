package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func generateKey() (string, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	k := make([]byte, 32)

	_, err := rand.Read(k)

	key := hex.EncodeToString(k)

	if err != nil {
		return "", err
	}
	return key, nil

}

func getRecipientEntity(email string, keyring string) (*openpgp.Entity, error) {
	keyringFileBuffer, err := os.Open(keyring)
	if err != nil {
		return nil, err
	}

	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		return nil, err
	}

	for e := range entityList {
		for id := range entityList[e].Identities {
			if strings.Compare(entityList[e].Identities[id].UserId.Email, email) == 0 {
				return entityList[e], nil
			}
		}

	}
	return nil, fmt.Errorf("Could not find key for %v in GPG keyring", email)
}

func encryptKey(key []byte, ent *openpgp.Entity) ([]byte, error) {

	hints := &openpgp.FileHints{
		IsBinary: false,
	}
	pktConfig := &packet.Config{
		DefaultCompressionAlgo: 0,
	}

	cryptoWriter := new(bytes.Buffer)

	plainwriter, err := openpgp.Encrypt(cryptoWriter, openpgp.EntityList{ent}, nil, hints, pktConfig)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	plainwriter.Write(key)

	plainwriter.Close()

	// Encode to base64
	b, err := ioutil.ReadAll(cryptoWriter)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	return b, nil

}
