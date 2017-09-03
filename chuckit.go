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
	var key string
	var err error

	flag.Usage = printUsage

	uid, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}

	localOnly := flag.Bool("local-only", false, "Write files locally only (test mode)")
	bucket := flag.String("bucket", "chuckit", "Bucket name to upload to")
	remoteName := flag.String("remote-name", "", "Remote name of file")
	endpoint := flag.String("endpoint", "s3-us-west-2.amazonaws.com", "AWS S3 endpoint")
	useCompression := flag.Bool("use-compression", false, "Compress archive with XZ")
	encRecipient := flag.String("enc-recipient", "", "Email address of recipient of symmetric key that is uploaded alongside archive")
	keyringFile := flag.String("keyring-file", uid.HomeDir+"/.gnupg/pubring.gpg", "Path to GPG public keyring")

	flag.Parse()

	if *encRecipient != "" {
		useEncryption = true
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	if useEncryption {

		if *useCompression {
			log.Fatalln("Cannot use encryption + compression because encrypted files do not benefit from compression.")
		}

		key, err = createAndStreamKey(*remoteName, *endpoint, *bucket, *encRecipient, *keyringFile, localOnly)
		if err != nil {
			log.Fatalln(err)
		}
	}

	createAndStreamArchive(*remoteName, *endpoint, *bucket, key, useCompression, localOnly)

}

func printUsage() {
	fmt.Printf("Usage: %v <args> <file or directory>\n\n", os.Args[0])
	flag.PrintDefaults()
}
