package goss

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type GossFormatter struct {
	GossResults
}

type GossResults struct {
	Tested  *[]Tested `json:"results,omitempty"`
	Summary *Summary  `json:"summary,omitempty"`
}

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

type Summary struct {
	FailedCount   int64 `json:"failed-count,omitempty"`
	TestCount     int64 `json:"test-count,omitempty"`
	TotalDuration int64 `json:"total-duration,omitempty"`
}

func (g *GossFormatter) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &g.GossResults)
}

func ValidateJSON(data []byte) (*GossResults, error) {
	var results GossResults
	err := json.Unmarshal(data, &results)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON data: %w", err)
	}
	return &results, nil
}

func (g *GossFormatter) FormatPromFriendly(f *os.File, t string) error {
	for _, result := range *g.Tested {
		var resourceId string

		switch result.ResourceType {
		case "Addr":
			resourceId = result.ResourceID
		case "Command":
			commandId := strings.Split(result.ResourceID, "|")
			resourceId = strings.TrimRight(strings.Replace(commandId[0], " -", "", -1), " ")
		case "Process":
			resourceId = strings.ReplaceAll(result.ResourceID, "/", "_")
		default:
			resourceId = result.ResourceID
		}

		f.WriteString(fmt.Sprintf("goss_result_%v{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceId, result.Property, result.Skipped, result.Result))
		f.WriteString(fmt.Sprintf("goss_result_%v_duration{property=\"%v\",resource=\"%v\",skipped=\"%t\"} %v\n",
			strings.ToLower(result.ResourceType), resourceId, result.Property, result.Skipped, result.Duration))
	}

	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"tested\"} %v\n", t, g.Summary.TestCount))
	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"failed\"} %v\n", t, g.Summary.FailedCount))
	f.WriteString(fmt.Sprintf("goss_results_summary{textfile=\"%v\",name=\"duration\"} %v\n", t, g.Summary.TotalDuration))

	return nil
}
