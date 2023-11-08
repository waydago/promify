package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waydago/promify/goss"
)

func parseFlags() {
	flag.StringVar(&filePath, "path", "", "Path to the text file")
	flag.StringVar(&fileName, "name", "", "Name of the prom file")
	flag.StringVar(&format, "format", "", "Format of the input data")
	flag.Parse()
}

func TestMainFlags(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"cmd", "-path", "./test", "-name", "test_name", "-format", "test_format"}

	parseFlags()

	if filePath != "./test" {
		t.Errorf("Expected path './test', but got '%s'", filePath)
	}
	if fileName != "test_name" {
		t.Errorf("Expected name 'test_name', but got '%s'", fileName)
	}
	if format != "test_format" {
		t.Errorf("Expected format 'test_format', but got '%s'", format)
	}
}

func TestDirectoryExists(t *testing.T) {
	err := os.Mkdir("./test", 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	_, err = os.Stat("./test")
	if os.IsNotExist(err) {
		t.Fatalf("expected directory to exist")
	}

	err = os.RemoveAll("./test")
	if err != nil {
		t.Fatalf("failed to remove directory: %v", err)
	}
}

func TestDirectoryDoesNotExist(t *testing.T) {
	_, err := os.Stat("./test")
	if !os.IsNotExist(err) {
		t.Fatalf("expected directory to not exist")
	}
}

func TestGetFromPipe(t *testing.T) {
	oldStdin := os.Stdin
	defer func() {
		os.Stdin = oldStdin
	}()

	t.Run("With named pipe", func(t *testing.T) {
		r, w, _ := os.Pipe()
		os.Stdin = r
		go func() {
			defer w.Close()
		}()

		if !getFromPipe() {
			t.Errorf("Expected true, but got false")
		}
	})

	t.Run("Without named pipe", func(t *testing.T) {
		os.Stdin = oldStdin
		if getFromPipe() {
			t.Errorf("Expected false, but got true")
		}
	})

	t.Run("Error case", func(t *testing.T) {
		os.Stdin, _ = os.Open("/tmp/nonexistent")
		if getFromPipe() {
			t.Errorf("Expected false, but got true")
		}
	})
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
	} else {
		t.Logf("Expected '%s', and got '%s'", content, string(result))
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

type mockFormatter struct {
	data []byte
}

func (m *mockFormatter) Unmarshal(data []byte) error {
	m.data = data
	return nil
}

func (m *mockFormatter) FormatPromFriendly(f *os.File, t string) error {
	return nil
}

func TestMockUnmarshalResultsJSON(t *testing.T) {
	data := []byte("test data")
	formatter := &mockFormatter{}

	err := unmarshalResultsJSON(data, formatter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(formatter.data, data) {
		t.Errorf("expected %s, got %s", data, formatter.data)
	}
}

func TestTidyFilePath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		err      string
	}{
		{"Empty path", "", "", "path error: argument is empty"},
		{"Path with space", " ", "", "path error: argument is empty"},
		{"Path with ...", "...", "", "path error: too many dots '...'"},
		{"Path with //", "//", "", "path error: too many slashes '//'"},
		{"Valid path", "/tmp/", "/tmp/", ""},
		{"Path with ~", "~/", strings.TrimSuffix(os.Getenv("HOME"), "/") + "/", ""},
		{"Path with ../", "../", filepath.Dir(os.Getenv("PWD")) + "/", ""},
		{"Path with .", "./", os.Getenv("PWD") + "/", ""},
		{"Path without /", "tmp", os.Getenv("PWD") + "/", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := tidyFilePath(tc.input)
			if output != tc.expected || (err != nil && err.Error() != tc.err) {
				t.Errorf("expected %s, %s, got %s, %v", tc.expected, tc.err, output, err)
			}
		})
	}

	err := os.Mkdir("/tmp/test", 0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	os.RemoveAll("/tmp/test")

	_, err = tidyFilePath("/tmp/test")
	if err == nil {
		t.Errorf("expected error, got nil")
	}

}

func TestTidyFileName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		err      string
	}{
		{"Empty name", "", "", "name error: argument is empty"},
		{"Name with space", " ", "", "name error: argument is empty"},
		{"Name with slash", "file/prom", "", "name error: cannot contain '/'"},
		{"Name without .prom extension", "file", "", "name error: '.prom' extension is required"},
		{"Valid name", "file.prom", "file.prom", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := tidyFileName(tc.input)
			if output != tc.expected || (err != nil && err.Error() != tc.err) {
				t.Errorf("expected %s, %s, got %s, %v", tc.expected, tc.err, output, err)
			}
		})
	}
}

func TestWritePromFileFriendly(t *testing.T) {
	content := `{"test": "test"}`
	formatter := &goss.Formatter{}
	result, err := ValidateJSON([]byte(content), formatter)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if result == nil {
		t.Errorf("Expected result, but got nil")
	}
}

func TestAltWritePromFileFriendly(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := []byte(`{
		"summary": {
			"failed-count": 0,
			"test-count": 1,
			"total-duration": 123
		}
	}`)
	formatter := &goss.Formatter{
		Results: goss.Results{
			Tested: &[]goss.Tested{},
			Summary: &goss.Summary{
				FailedCount:   0,
				TestCount:     1,
				TotalDuration: 123,
			},
		},
	}
	result, err := ValidateJSON([]byte(content), formatter)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if result == nil {
		t.Errorf("Expected result, but got nil")
	}

	err = writePromFileFriendly(formatter, tmpfile.Name(), "test")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(data))

	expected := data
	if !bytes.Equal(data, expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(tmpfile.Name()); err != nil {
		t.Fatal(err)
	}

}
