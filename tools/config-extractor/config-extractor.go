package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/itchio/lzma"
	"io"
	"os"
	"strings"
)

func main() {
	iName := flag.String("i", "", "log file")
	oName := flag.String("o", "", "output file (optional)")
	flag.Parse()

	if *iName == "" {
		flag.Usage()
		os.Exit(1)
	}

	compressed, err := extractConfigString(*iName)
	if err != nil {
		panic(fmt.Errorf("failed to extract string: [%w]", err))
	}

	decompressed, err := getDecompressedString(compressed)
	if err != nil {
		panic(fmt.Errorf("failed to decompress config string: [%w]", err))
	}

	if *oName != "" {
		err = os.WriteFile(*oName, []byte(decompressed), 0655)
		if err != nil {
			panic(fmt.Errorf("failed to write to file: [%w]", err))
		}
	} else {
		fmt.Println(decompressed)
	}
}

func extractConfigString(logPath string) (string, error) {
	iFile, err := os.Open(logPath)
	if err != nil {
		return "", err
	}

	defer iFile.Close()

	scanner := bufio.NewScanner(iFile)

	buf := make([]byte, 0, 10*1024*1024) // being generous
	scanner.Buffer(buf, cap(buf))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Current config: ") {
			return strings.Split(line, "Current config: ")[1], nil
		}
	}

	return "", errors.New("config string not found")
}

func getDecompressedString(in string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return "", err
	}

	reader := lzma.NewReader(bytes.NewReader(decoded))
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}
