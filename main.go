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

var Format, TextFilePath, PromFileName string

func LoadPipedData() []byte {
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

type Formatter interface {
	Unmarshal([]byte) error
	FormatPromFriendly(*os.File, string) error
}

func ValidateJSON(data []byte, formatter Formatter) (interface{}, error) {
	err := formatter.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON data: %w", err)
	}
	return formatter, nil
}

func UnmarshalResultsJSON(data []byte, formatter Formatter) error {
	return formatter.Unmarshal(data)
}

func WritePromFileFriendly(formatter Formatter, dotprom string, t string) error {
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
	flag.StringVar(&TextFilePath, "path", "/var/lib/node_exporter/textfile_collector", "Where to store the .prom file")
	flag.StringVar(&PromFileName, "name", "", "Name your .prom with the extension")
	flag.StringVar(&Format, "format", "goss", "Format of the input data (goss or debugvarz)")
	flag.Parse()

	var formatter Formatter
	switch Format {
	case "goss":
		formatter = &goss.GossFormatter{}
	default:
		log.Fatalf("Unsupported format: %s", Format)
	}

	data := LoadPipedData()

	err := UnmarshalResultsJSON(data, formatter)
	if err != nil {
		log.Fatal(err)
	}

	File := fmt.Sprintf("%v/%v", TextFilePath, PromFileName)
	err = WritePromFileFriendly(formatter, File, PromFileName)
	if err != nil {
		log.Fatal(err)
	}
}
