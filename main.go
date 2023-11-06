package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/waydago/promify/goss"
)

var format string
var textFilePath string
var promFileName string

func checkIfPiped() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if fi.Mode()&os.ModeNamedPipe != 0 {
		return true
	}

	fmt.Println("Program must be called as a pipe")
	return false

}

func loadPipedData() []byte {
	var dataPiped bytes.Buffer
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				dataPiped.WriteString(line)
				break
			} else {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		dataPiped.WriteString(line)
	}
	return dataPiped.Bytes()
}

type formatter interface {
	Unmarshal([]byte) error
	FormatPromFriendly(*os.File, string) error
}

// ValidateJSON checks if the provided input is a valid JSON
func ValidateJSON(data []byte, formatter formatter) (interface{}, error) {
	err := formatter.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON data: %w", err)
	}
	return formatter, nil
}

func unmarshalResultsJSON(data []byte, formatter formatter) error {
	return formatter.Unmarshal(data)
}

func writePromFileFriendly(formatter formatter, dotprom string, t string) error {
	f, err := os.Create(dotprom)
	if err != nil {
		return err
	}
	defer f.Close()

	err = formatter.FormatPromFriendly(f, t)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	flag.StringVar(&textFilePath, "path", "/var/lib/node_exporter/textfile_collector", "Where to store the .prom file")
	flag.StringVar(&promFileName, "name", "", "Name your .prom with the extension (required)")
	flag.StringVar(&format, "format", "goss", "Format of the input data")
	flag.Parse()

	if promFileName == "" {
		fmt.Println("name is required")
		os.Exit(1)
	}

	var formatter formatter
	switch format {
	case "goss":
		formatter = &goss.Formatter{}
	default:
		log.Fatalf("Unsupported format: %s", format)
	}

	if !checkIfPiped() {
		os.Exit(1)
	}

	data := loadPipedData()

	err := unmarshalResultsJSON(data, formatter)
	if err != nil {
		log.Fatal(err)
	}

	file := fmt.Sprintf("%v/%v", textFilePath, promFileName)
	err = writePromFileFriendly(formatter, file, promFileName)
	if err != nil {
		log.Fatal(err)
	}
}
