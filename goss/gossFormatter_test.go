package goss

import (
	"log"
	"os"
	"testing"
)

func createTestData() ([]byte, *Formatter) {
	data := []byte(`{
		"results": [{
			"duration": 10,
			"expected": ["yes"],
			"found": ["yes"],
			"property": "test_property",
			"resource-id": "test_id",
			"resource-type": "test_type",
			"result": 1, "tested": [],
			"skipped": false,
			"successful": true,
			"test-type": 1
		}],
		"summary": {
			"failed-count": 0,
			"test-count": 1,
			"total-duration": 10
		}
	}`)

	formatter := &Formatter{}

	return data, formatter
}

func TestGossFormatter_Unmarshal(t *testing.T) {
	data, formatter := createTestData()

	err := formatter.Unmarshal(data)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	if len(*formatter.Tested) != 1 {
		t.Errorf("Expected 1 test result, but got %v", len(*formatter.Tested))
	}

	tested := (*formatter.Tested)[0]
	if tested.Property != "test_property" {
		t.Errorf("Expected property 'test_property', but got %s", tested.Property)
	}
}

var formatter = &Formatter{
	Results: Results{
		Tested: &[]Tested{
			{
				Duration:     10,
				Expected:     []string{"yes"},
				Found:        []string{"yes"},
				Property:     "test_property",
				ResourceID:   "test_id",
				ResourceType: "test_type",
				Result:       1,
				Skipped:      false,
				Successful:   true,
				TestType:     1,
			},
		},
		Summary: &Summary{
			FailedCount:   0,
			TestCount:     1,
			TotalDuration: 10,
		},
	},
}

var expectedContent = `# HELP test_property Goss test_property
test_property{resource_id="test_id",resource_type="test_type"} 1
goss_test_duration{resource_id="test_id",resource_type="test_type"} 10
goss_test_result{resource_id="test_id",resource_type="test_type"} 1
goss_test_skipped{resource_id="test_id",resource_type="test_type"} 0
goss_test_successful{resource_id="test_id",resource_type="test_type"} 1
goss_test_type{resource_id="test_id",resource_type="test_type"} 1
goss_test_expected{resource_id="test_id",resource_type="test_type"} 1
goss_test_found{resource_id="test_id",resource_type="test_type"} 1
goss_test_count 1
goss_test_failed_count 0
goss_test_total_duration 10
goss_test_duration_seconds{resource_id="test_id",resource_type="test_type"} 10
goss_test_total_duration_seconds 10
goss_test_expected_count{resource_id="test_id",resource_type="test_type"} 1
goss_test_found_count{resource_id="test_id",resource_type="test_type"} 1
goss_test_expected_count_percent{resource_id="test_id",resource_type="test_type"} 100
goss_test_found_count_percent{resource_id="test_id",resource_type="test_type"} 100
goss_test_expected_count_percent_total{resource_id="test_id",resource_type="test_type"} 100
goss_test_found_count_percent_total{resource_id="test_id",resource_type="test_type"} 100
`

func TestGossFormatter_FormatPromFriendly(t *testing.T) {

	tmpfile, err := os.CreateTemp("", "example.*.prom")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	err = formatter.FormatPromFriendly(tmpfile, "test")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())

		content, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if string(content) != expectedContent {
			t.Errorf("Expected content:\n%s\n\nBut got:\n%s", expectedContent, string(content))
		}
	}
}
