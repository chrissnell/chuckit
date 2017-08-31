package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"

	s3 "github.com/rlmcpherson/s3gof3r"
	"github.com/ulikunitz/xz"
)

func main() {

	var useEncryption bool
	var remoteName string
	var endpoint string
	var encKey string

	flag.Usage = printUsage

	bucket := flag.String("bucket", "", "Bucket name to upload to")
	flag.StringVar(&remoteName, "remote-name", "", "Remote name of file")
	flag.StringVar(&endpoint, "endpoint", "s3-us-west-2.amazonaws.com", "AWS S3 endpoint")
	flag.StringVar(&encKey, "enc-key", "", "Symmetric encryption key")
	flag.Parse()

	if *bucket == "" {
		printUsage()
		os.Exit(1)
	}

	if encKey != "" {
		useEncryption = true
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	fileName := os.Args[len(os.Args)-1]

	k, err := s3.EnvKeys() // get S3 keys from environment
	if err != nil {
		log.Fatalln(err)
	}
	// Open bucket to put file into
	s3 := s3.New("", k)

	s3.Domain = endpoint

	b := s3.Bucket(*bucket)

	// open file to upload
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
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

	info, err := os.Stat(fileName)
	if err != nil {
		log.Fatalln("Could not stat file", fileName, err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(fileName)
	}

	err = filepath.Walk(fileName,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, fileName))
			}

			if err = tw.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			addfile, err := os.Open(path)
			if err != nil {
				return err
			}

			log.Println("Adding", path, "...")
			defer file.Close()
			_, err = io.Copy(tw, addfile)
			if err != nil {
				log.Fatalln("Error adding", addfile, "to tar archive:", err)
			}

			return nil
		})

	if err != nil {
		log.Fatalln("Error walking directory structure:", err)
	}

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

func printUsage() {
	fmt.Printf("Usage: %v <args> <file or directory>\n\n", os.Args[0])
	flag.PrintDefaults()
}
