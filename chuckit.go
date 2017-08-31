package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
)

func main() {

	var useEncryption bool
	var key []byte
	var err error

	flag.Usage = printUsage

	uid, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	bucket := flag.String("bucket", "", "Bucket name to upload to")
	remoteName := flag.String("remote-name", "", "Remote name of file")
	endpoint := flag.String("endpoint", "s3-us-west-2.amazonaws.com", "AWS S3 endpoint")
	encRecipient := flag.String("enc-recipient", "", "Email address of recipient of symmetric key that is uploaded alongside archive")
	keyringFile := flag.String("keyring-file", uid.HomeDir+"/.gnupg/pubring.gpg", "Path to GPG public keyring")

	flag.Parse()

	if *bucket == "" {
		printUsage()
		os.Exit(1)
	}

	if *encRecipient != "" {
		useEncryption = true
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	if useEncryption {
		key, err = createAndStreamKey(*remoteName, *endpoint, *bucket, *encRecipient, *keyringFile)
		if err != nil {
			log.Fatalln(err)
		}
	}

	createAndStreamArchive(*remoteName, *endpoint, *bucket, key)

}

func printUsage() {
	fmt.Printf("Usage: %v <args> <file or directory>\n\n", os.Args[0])
	flag.PrintDefaults()
}
