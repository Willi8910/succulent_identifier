package utils

import (
	"strings"
)

// ParseLabel extracts genus and species from a label
// Label format: "genus_species" (e.g., "echeveria_elegans")
func ParseLabel(label string) (genus string, species string) {
	parts := strings.Split(label, "_")

	if len(parts) >= 1 {
		genus = parts[0]
	}

	if len(parts) >= 2 {
		species = label // Full label represents species
	}

	return genus, species
}

// FormatGenus formats genus name for display (capitalize first letter)
func FormatGenus(genus string) string {
	if genus == "" {
		return ""
	}
	return strings.ToUpper(string(genus[0])) + genus[1:]
}

// FormatSpecies formats species name for display
// Converts "genus_species" to "Genus species"
func FormatSpecies(label string) string {
	parts := strings.Split(label, "_")
	if len(parts) < 2 {
		return ""
	}

	genus := FormatGenus(parts[0])
	species := strings.Join(parts[1:], " ")

	return genus + " " + species
}
