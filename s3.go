package main

import (
	"fmt"
	"io"
	"log"

	s3 "github.com/rlmcpherson/s3gof3r"
)

func newS3Writer(endpoint string, bucket string, remoteName string) (io.WriteCloser, error) {
	if remoteName == "" {
		return nil, fmt.Errorf("Error: Could not create S3 writer.  Remote filename cannot be nil.")
	}

	k, err := s3.EnvKeys() // get S3 keys from environment
	if err != nil {
		return nil, err
	}

	// Open bucket to put file into
	s3 := s3.New("", k)

	s3.Domain = endpoint

	b := s3.Bucket(bucket)

	// Open a PutWriter for upload
	w, err := b.PutWriter(remoteName, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	return w, nil

}
