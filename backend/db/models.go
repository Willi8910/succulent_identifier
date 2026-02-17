package db

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrNotFound = errors.New("record not found")
)

// CareGuide represents plant care instructions
type CareGuide struct {
	Sunlight string `json:"sunlight"`
	Watering string `json:"watering"`
	Soil     string `json:"soil"`
	Notes    string `json:"notes,omitempty"`
}

// Identification represents a plant identification record
type Identification struct {
	ID         string     `json:"id"`
	Genus      string     `json:"genus"`
	Species    string     `json:"species"`
	Confidence float64    `json:"confidence"`
	ImagePath  string     `json:"image_path"`
	CareGuide  *CareGuide `json:"care_guide"` // Stored as JSONB in database
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"` // Soft delete timestamp
}

// ChatMessage represents a chat message in a conversation
type ChatMessage struct {
	ID               string    `json:"id"`
	IdentificationID string    `json:"identification_id"`
	Message          string    `json:"message"`
	Sender           string    `json:"sender"` // "user" or "llm"
	CreatedAt        time.Time `json:"created_at"`
}

// IdentificationWithChats represents an identification with its chat history
type IdentificationWithChats struct {
	Identification Identification `json:"identification"`
	ChatMessages   []ChatMessage  `json:"chat_messages"`
}

// CareInstructionsCache represents cached LLM-generated care instructions
type CareInstructionsCache struct {
	ID         string     `json:"id"`
	Genus      string     `json:"genus"`
	Species    string     `json:"species"`
	CareGuide  *CareGuide `json:"care_guide"` // Stored as JSONB in database
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
