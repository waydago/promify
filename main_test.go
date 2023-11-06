package main

import (
	"flag"
	"os"
	"testing"

	"github.com/waydago/promify/goss"
)

func parseFlags() {
	flag.StringVar(&textFilePath, "path", "", "Path to the text file")
	flag.StringVar(&promFileName, "name", "", "Name of the prom file")
	flag.Parse()
}

func TestMainFlags(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"cmd", "-path", "/test/path", "-name", "test_name"}

	parseFlags()

	// Check if the flags were correctly parsed
	if textFilePath != "/test/path" {
		t.Errorf("Expected path '/test/path', but got '%s'", textFilePath)
	}
	if promFileName != "test_name" {
		t.Errorf("Expected name 'test_name', but got '%s'", promFileName)
	}
}
func TestLoadPipedData(t *testing.T) {
	content := "Hello\nWorld\n"
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()
	os.Stdin = r
	go func() {
		defer w.Close()
		_, err := w.WriteString(content)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
	}()

	result := loadPipedData()
	if string(result) != content {
		t.Errorf("Expected '%s', but got '%s'", content, string(result))
	}
}

func TestValidateJSON(t *testing.T) {
	content := `{"test": "test"}`
	result, err := ValidateJSON([]byte(content), &goss.Formatter{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if result == nil {
		t.Errorf("Expected result, but got nil")
	}
}

func TestUnmarshalResultsJSON(t *testing.T) {
	content := `{"test": "test"}`
	err := unmarshalResultsJSON([]byte(content), &goss.Formatter{})
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
}
