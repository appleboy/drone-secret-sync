package main

import (
	"os"
	"testing"
)

func TestGetGlobalValue(t *testing.T) {
	// Simulate environment variables
	os.Setenv("KEY1", "value1")
	os.Setenv("PLUGIN_KEY2", "value2")

	// Test case 1: Testing "key2"
	// Expected result: "value2"
	expected := "value2"
	actual := getGlobalValue("key2")

	if actual != expected {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}

	// Test case 2: Testing case conversion to uppercase
	// Expected result: "value1"
	expected = "value1"
	actual = getGlobalValue("Key1")

	if actual != expected {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}

	// Test case 3: Testing a non-existent environment variable
	// Expected result: ""
	expected = ""
	actual = getGlobalValue("key3")

	if actual != expected {
		t.Errorf("Expected %s, but got %s", expected, actual)
	}
}
