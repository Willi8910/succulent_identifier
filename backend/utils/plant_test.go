package utils

import (
	"testing"
)

func TestParseLabel(t *testing.T) {
	tests := []struct {
		name           string
		label          string
		expectedGenus  string
		expectedSpecies string
	}{
		{
			name:           "Valid label with genus and species",
			label:          "echeveria_elegans",
			expectedGenus:  "echeveria",
			expectedSpecies: "echeveria_elegans",
		},
		{
			name:           "Valid label with multiple underscores",
			label:          "haworthia_zebra_plant",
			expectedGenus:  "haworthia",
			expectedSpecies: "haworthia_zebra_plant",
		},
		{
			name:           "Label with genus only",
			label:          "echeveria",
			expectedGenus:  "echeveria",
			expectedSpecies: "",
		},
		{
			name:           "Empty label",
			label:          "",
			expectedGenus:  "",
			expectedSpecies: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			genus, species := ParseLabel(tt.label)

			if genus != tt.expectedGenus {
				t.Errorf("ParseLabel() genus = %v, expected %v", genus, tt.expectedGenus)
			}

			if species != tt.expectedSpecies {
				t.Errorf("ParseLabel() species = %v, expected %v", species, tt.expectedSpecies)
			}
		})
	}
}

func TestFormatGenus(t *testing.T) {
	tests := []struct {
		name     string
		genus    string
		expected string
	}{
		{
			name:     "Lowercase genus",
			genus:    "echeveria",
			expected: "Echeveria",
		},
		{
			name:     "Already capitalized",
			genus:    "Echeveria",
			expected: "Echeveria",
		},
		{
			name:     "Single letter",
			genus:    "e",
			expected: "E",
		},
		{
			name:     "Empty string",
			genus:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatGenus(tt.genus)
			if result != tt.expected {
				t.Errorf("FormatGenus() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFormatSpecies(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		expected string
	}{
		{
			name:     "Valid species label",
			label:    "echeveria_elegans",
			expected: "Echeveria elegans",
		},
		{
			name:     "Species with multiple words",
			label:    "haworthia_zebra_plant",
			expected: "Haworthia zebra plant",
		},
		{
			name:     "Genus only (no species)",
			label:    "echeveria",
			expected: "",
		},
		{
			name:     "Empty label",
			label:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSpecies(tt.label)
			if result != tt.expected {
				t.Errorf("FormatSpecies() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
