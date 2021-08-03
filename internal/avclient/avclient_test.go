package avclient

import (
	"fmt"
	"net/http"
	"testing"
)

type TestKong struct {
}

func (k TestKong) Debug(args ...interface{}) error {
	_, err := fmt.Printf("Debug:%v\n", args)
	return err
}

func (k TestKong) Info(args ...interface{}) error {
	_, err := fmt.Printf("Info:%v\n", args)
	return err
}

func (k TestKong) Warn(args ...interface{}) error {
	_, err := fmt.Printf("Warn:%v\n", args)
	return err
}

func (k TestKong) Err(args ...interface{}) error {
	_, err := fmt.Printf("Err:%v\n", args)
	return err
}

func (k TestKong) ResponseSetHeader(key string, v string) error {
	fmt.Printf("ResponseSetHeader:(%v, %v)\n", key, v)
	return nil
}

func (k TestKong) ResponseExit(status int, body string, headers map[string][]string) {
	fmt.Printf("ResponseSetHeader:(%v, %v, %v)\n", status, body, headers)
}

func TestDetermineIsInfectedMultipleHeaders(t *testing.T) {
	h := make(http.Header, 0)
	h.Add("X-Virus-ID", "no threats")
	h.Add("X-FSecure-Scan-Result", "infected")

	actualResult := determineIsInfected(TestKong{}, h)
	expectedResult := true
	if actualResult != expectedResult {
		t.Errorf("Actual:%v, Expected:%v", actualResult, expectedResult)
	}
}

func TestDetermineIsInfectedSingleHeader(t *testing.T) {
	cases := []struct {
		name           string
		key            string
		value          string
		expectedResult bool
	}{
		{
			name:           "no headers",
			key:            "foo",
			value:          "not empty",
			expectedResult: false,
		},
		{
			name:           "X-Infection-Found is set",
			key:            "X-Infection-Found",
			value:          "not empty",
			expectedResult: true,
		},
		{
			name:           "X-Virus-ID is set and OK",
			key:            "X-Virus-ID",
			value:          "no threats",
			expectedResult: false,
		},
		{
			name:           "X-Virus-ID is set",
			key:            "X-Virus-ID",
			value:          "not empty",
			expectedResult: true,
		},
		{
			name:           "X-FSecure-Scan-Result is misc",
			key:            "X-FSecure-Scan-Result",
			value:          "not empty",
			expectedResult: false,
		},
		{
			name:           "X-FSecure-Scan-Result is set",
			key:            "X-FSecure-Scan-Result",
			value:          "infected",
			expectedResult: true,
		},
		{
			name:           "X-FSecure-Scan-Result is empty",
			key:            "X-FSecure-Scan-Result",
			value:          "",
			expectedResult: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := make(http.Header, 0)
			h.Add(tc.key, tc.value)

			kong := TestKong{}
			actualResult := determineIsInfected(kong, h)
			expectedResult := tc.expectedResult
			if actualResult != expectedResult {
				t.Errorf("Actual:%v, Expected:%v", actualResult, expectedResult)
			}
		})
	}
}
