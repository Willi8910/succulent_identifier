package models

// MLPrediction represents a single prediction from the ML service
type MLPrediction struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
}

// MLInferenceRequest represents the request to ML service
type MLInferenceRequest struct {
	ImagePath string `json:"image_path"`
}

// MLInferenceResponse represents the response from ML service
type MLInferenceResponse struct {
	Predictions []MLPrediction `json:"predictions"`
}

// CareInstructions represents plant care information
type CareInstructions struct {
	Sunlight string `json:"sunlight"`
	Watering string `json:"watering"`
	Soil     string `json:"soil"`
	Notes    string `json:"notes"`
}

// PlantInfo represents identified plant information
type PlantInfo struct {
	Genus      string  `json:"genus"`
	Species    string  `json:"species,omitempty"`
	Confidence float64 `json:"confidence"`
}

// IdentifyResponse represents the response to the client
type IdentifyResponse struct {
	ID    string           `json:"id"`
	Plant PlantInfo        `json:"plant"`
	Care  CareInstructions `json:"care"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
