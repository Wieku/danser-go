package files

import (
	"fmt"
	"io"
	"os"
)

// MoveFile allows moving the file across drives compared to os.Rename
func MoveFile(source, destination string) error {
	inputFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	defer inputFile.Close()

	outputFile, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}

	defer outputFile.Close()

	if _, err = io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("couldn't copy contents: %s", err)
	}

	inputFile.Close()

	if err = os.Remove(source); err != nil {
		return fmt.Errorf("couldn't remove source file: %s", err)
	}

	return nil
}
