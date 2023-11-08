package goss

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Formatter holds the results of a goss test
type Formatter struct {
	Results
}

// Results separates the summary from the individual test results
type Results struct {
	Tested  *[]Tested `json:"results,omitempty"`
	Summary *Summary  `json:"summary,omitempty"`
}

// Tested holds the individual test results of each goss test
type Tested struct {
	Duration     int64    `json:"duration,omitempty"`
	Expected     []string `json:"expected,omitempty"`
	Found        []string `json:"found,omitempty"`
	Property     string   `json:"property,omitempty"`
	ResourceID   string   `json:"resource-id,omitempty"`
	ResourceType string   `json:"resource-type,omitempty"`
	Result       int64    `json:"result,omitempty"`
	Skipped      bool     `json:"skipped,omitempty"`
	Successful   bool     `json:"successful,omitempty"`
	TestType     int64    `json:"test-type,omitempty"`
}

// Summary holds the summary of a goss test
type Summary struct {
	FailedCount   int64 `json:"failed-count,omitempty"`
	TestCount     int64 `json:"test-count,omitempty"`
	TotalDuration int64 `json:"total-duration,omitempty"`
}

func errorCheck(err error) {
	if err != nil {
		fmt.Println("error writing to file: %w", err)
	}
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by g.Results
func (g *Formatter) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &g.Results)
}

// FormatPromFriendly writes the results of a goss test to a file in a format that Prometheus can read
func (g *Formatter) FormatPromFriendly(f *os.File, t string) error {
	for _, result := range *g.Tested {
		var resourceID string

		switch result.ResourceType {
		case "Addr":
			resourceID = result.ResourceID
		case "Command":
			commandID := strings.Split(result.ResourceID, "|")
			resourceID = strings.TrimRight(strings.Replace(commandID[0], " -", "", -1), " ")
		case "Process":
			resourceID = strings.ReplaceAll(result.ResourceID, "/", "_")
		default:
			resourceID = result.ResourceID
		}

		_, err := f.WriteString(fmt.Sprintf("goss_result_%v{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceID, result.Property, result.Skipped, result.Result))
		errorCheck(err)
		_, err = f.WriteString(fmt.Sprintf("goss_result_%v_duration{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceID, result.Property, result.Skipped, result.Duration))
		errorCheck(err)
	}

	_, err := f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"tested\"} %v\n", t, g.Summary.TestCount))
	errorCheck(err)
	_, err = f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"failed\"} %v\n", t, g.Summary.FailedCount))
	errorCheck(err)
	_, err = f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"duration\"} %v\n", t, g.Summary.TotalDuration))
	errorCheck(err)

	return nil
}
