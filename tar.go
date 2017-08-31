package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func addFilesToTarArchive(path string, tw *tar.Writer) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Could not stat file %v: %v", path, err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(path)
	}

	err = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("Error walking file/directory: %v", err)
			}

			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, path))
			}

			if err = tw.WriteHeader(header); err != nil {
				return fmt.Errorf("Error writing tar header: %v", err)
			}

			if info.IsDir() {
				return nil
			}

			addfile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Could not open file %v: %v", path, err)
			}

			log.Println("Adding", path, "...")
			defer addfile.Close()
			_, err = io.Copy(tw, addfile)
			if err != nil {
				return fmt.Errorf("Error copying %v to tar archive: %v", path, err)
			}

			return nil
		})

	return err
}
