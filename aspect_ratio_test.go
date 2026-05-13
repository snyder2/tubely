package main

import (
	"fmt"
	"testing"
)

func TestGetAspectRatio(t *testing.T) {
	cases := []struct {
		filePath string
		expected string
	}{
		{
			filePath: "./samples/boots-video-horizontal.mp4",
			expected: "16:9",
		},
		{
			filePath: "./samples/boots-video-vertical.mp4",
			expected: "9:16",
		},
	}

	for _, c := range cases {
		result, err := getVideoAspectRatio(c.filePath)
		if err != nil {
			t.Errorf("Error: %s", err)
		}

		fmt.Println("Input file path:", c.filePath)
		fmt.Println("Expected result:", c.expected)
		fmt.Println("Actual result:", result)

		if c.expected != result {
			t.Errorf("Result does not match expected output")
		}
	}
}
