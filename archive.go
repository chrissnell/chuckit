package main

import (
	"archive/tar"
	"io"
	"log"
	"os"

	s3 "github.com/rlmcpherson/s3gof3r"
	"github.com/ulikunitz/xz"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func createAndStreamArchive(remoteName string, endpoint string, bucket string, useEncryption bool, encKey string) {
	fileName := os.Args[len(os.Args)-1]

	k, err := s3.EnvKeys() // get S3 keys from environment
	if err != nil {
		log.Fatalln(err)
	}
	// Open bucket to put file into
	s3 := s3.New("", k)

	s3.Domain = endpoint

	b := s3.Bucket(bucket)

	file, err := os.Stat(fileName)
	if err != nil {
		log.Fatalln("Could not stat file/directory", fileName, ":", err)
	}

	if remoteName == "" {
		if useEncryption {
			remoteName = file.Name() + ".tar.aes.xz"
		} else {
			remoteName = file.Name() + "tar.xz"
		}
	}

	// Open a PutWriter for upload
	w, err := b.PutWriter(remoteName, nil, nil)
	if err != nil {
		log.Fatalln(err)

	}

	xzw, err := xz.NewWriter(w)
	if err != nil {
		log.Fatalf("xz.NewWriter error %s", err)
	}

	var tw *tar.Writer
	var pgpw io.WriteCloser

	if useEncryption {

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

	if useEncryption {
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
