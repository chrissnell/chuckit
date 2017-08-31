package main

import (
	"archive/tar"
	"io"
	"log"
	"os"

	"github.com/ulikunitz/xz"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func createAndStreamKey(remoteName string, endpoint string, bucket string, recipient string, keyring string) ([]byte, error) {

	fileName := os.Args[len(os.Args)-1]

	file, err := os.Stat(fileName)
	if err != nil {
		log.Fatalln("Could not stat file/directory", fileName, ":", err)
	}

	if remoteName == "" {
		remoteName = file.Name() + ".key.aes"
	} else {
		remoteName = remoteName + ".key.aes"
	}

	key, err := generateKey()
	if err != nil {
		return nil, err
	}

	ent, err := getRecipientEntity(recipient, keyring)
	if err != nil {
		return nil, err
	}

	encKey, err := encryptKey(key, ent)
	if err != nil {
		return nil, err
	}

	w, err := newS3Writer(endpoint, bucket, remoteName)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = w.Write(encKey)
	if err != nil {
		log.Fatalln(err)
	}

	if err = w.Close(); err != nil {
		log.Fatalln(err)
	}

	return key, nil
}

func createAndStreamArchive(remoteName string, endpoint string, bucket string, encKey []byte) {

	fileName := os.Args[len(os.Args)-1]

	file, err := os.Stat(fileName)
	if err != nil {
		log.Fatalln("Could not stat file/directory", fileName, ":", err)
	}

	if remoteName == "" {
		if len(encKey) == 0 {
			remoteName = file.Name() + ".tar.xz"
		} else {
			remoteName = file.Name() + ".tar.aes.xz"
		}
	}

	// Get a new PutWriter for upload
	w, err := newS3Writer(endpoint, bucket, remoteName)
	if err != nil {
		log.Fatalln(err)

	}

	xzw, err := xz.NewWriter(w)
	if err != nil {
		log.Fatalf("xz.NewWriter error %s", err)
	}

	var tw *tar.Writer
	var pgpw io.WriteCloser

	if len(encKey) == 0 {

		hints := &openpgp.FileHints{
			IsBinary: false,
		}
		config := &packet.Config{
			DefaultCompressionAlgo: 0,
		}

		pgpw, err = openpgp.SymmetricallyEncrypt(xzw, []byte(encKey), hints, config)
		if err != nil {
			log.Fatalln("Could not start OpenPGP encryption:", err)
		}

		tw = tar.NewWriter(pgpw)

	} else {
		tw = tar.NewWriter(xzw)
	}

	addFilesToTarArchive(fileName, tw)

	// Make sure to check the error on Close.
	if err = tw.Close(); err != nil {
		log.Fatalln(err)
	}

	if len(encKey) == 0 {
		if err = pgpw.Close(); err != nil {
			log.Fatalln(err)
		}
	}

	if err = xzw.Close(); err != nil {
		log.Println("Error closing XZ writer", err)
	}

	if err = w.Close(); err != nil {
		log.Fatalln(err)
	}

}
