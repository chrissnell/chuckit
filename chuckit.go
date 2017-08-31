package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	var useEncryption bool
	// var remoteName string
	// var endpoint string
	// var encKey string

	flag.Usage = printUsage

	bucket := flag.String("bucket", "", "Bucket name to upload to")
	remoteName := flag.String("remote-name", "", "Remote name of file")
	endpoint := flag.String("endpoint", "s3-us-west-2.amazonaws.com", "AWS S3 endpoint")
	encKey := flag.String("enc-key", "", "Symmetric encryption key")

	// flag.StringVar(&remoteName, "remote-name", "", "Remote name of file")
	// flag.StringVar(&endpoint, "endpoint", "s3-us-west-2.amazonaws.com", "AWS S3 endpoint")
	// flag.StringVar(&encKey, "enc-key", "", "Symmetric encryption key")
	flag.Parse()

	if *bucket == "" {
		printUsage()
		os.Exit(1)
	}

	if *encKey != "" {
		useEncryption = true
	}

	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}

	createAndStreamArchive(*remoteName, *endpoint, *bucket, useEncryption, *encKey)

}

func printUsage() {
	fmt.Printf("Usage: %v <args> <file or directory>\n\n", os.Args[0])
	flag.PrintDefaults()
}
