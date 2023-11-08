package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/waydago/promify/goss"
)

var format string
var filePath string
var fileName string

func getFromPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println(err.Error())
		return false
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
				log.Fatal(err)
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

func tidyFileName(name string) (string, error) {
	switch {
	case name == "" || name == " ":
		return "", errors.New("name error: argument is empty")
	case strings.Contains(name, "/"):
		return "", errors.New("name error: cannot contain '/'")
	case !strings.HasSuffix(name, ".prom"):
		return "", errors.New("name error: '.prom' extension is required")
	default:
		name = strings.TrimRight(name, " \t\n\r\x0b\x0c")
		name = strings.TrimLeft(name, " \t\n\r\x0b\x0c")
	}

	return name, nil
}

func tidyFilePath(path string) (string, error) {
	switch {
	case len(path) == 0 || path == " " || path == "":
		return "", errors.New("path error: argument is empty")
	case strings.HasPrefix(path, "..."):
		return "", errors.New("path error: too many dots '...'")
	case strings.Contains(path, "//"):
		return "", errors.New("path error: too many slashes '//'")
	}

	switch {
	case strings.HasPrefix(path, "~") || strings.HasPrefix(path, "$HOME"):
		path = strings.Replace(path, "~", os.Getenv("HOME"), 1)
	case strings.HasPrefix(path, "../"):
		path = filepath.Dir(os.Getenv("PWD"))
	case strings.HasPrefix(path, ".") || !strings.HasPrefix(path, "/"):
		path = os.Getenv("PWD")
	default:
		path = filepath.Clean(path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("path error: directory does not exist")
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	return path, nil
}

func main() {

	if !getFromPipe() {
		log.Fatal("Program must be called as a pipe")
	}

	flag.StringVar(&filePath, "path", "/var/lib/node_exporter/textfile_collector", "Path to store the .prom file")
	flag.StringVar(&fileName, "name", "", "Name your .prom with the extension (required)")
	flag.StringVar(&format, "format", "goss", "Format of the input data")
	flag.Parse()

	fileName, err := tidyFileName(fileName)
	if err != nil {
		log.Fatal(err)
	}

	filePath, err := tidyFilePath(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var formatter formatter
	switch format {
	case "goss":
		formatter = &goss.Formatter{}
	default:
		log.Fatalf("Unsupported format: %s", format)
	}

	data := loadPipedData()

	err = unmarshalResultsJSON(data, formatter)
	if err != nil {
		log.Fatal(err)
	}

	file := path.Join(filePath, fileName)

	err = writePromFileFriendly(formatter, file, fileName)
	if err != nil {
		log.Fatal(err)
	}
}
