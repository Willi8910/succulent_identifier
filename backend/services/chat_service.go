package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"succulent-identifier-backend/db"
)

// ChatService handles LLM chat interactions
type ChatService struct {
	client *openai.Client
	model  string
}

// NewChatService creates a new chat service
func NewChatService(apiKey string) *ChatService {
	return &ChatService{
		client: openai.NewClient(apiKey),
		model:  openai.GPT4oMini, // Using GPT-4o-mini for cost efficiency
	}
}

// ChatRequest represents a chat request with context
type ChatRequest struct {
	UserMessage      string
	Identification   *db.Identification
	ConversationHistory []db.ChatMessage
}

// ChatResponse represents the LLM response
type ChatResponse struct {
	Message string
	Error   error
}

// Chat sends a message to OpenAI with plant identification context
func (s *ChatService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Build system prompt with plant context
	systemPrompt := s.buildSystemPrompt(req.Identification)

	// Build messages array
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// Add conversation history (last 10 messages for context)
	historyLimit := 10
	startIdx := 0
	if len(req.ConversationHistory) > historyLimit {
		startIdx = len(req.ConversationHistory) - historyLimit
	}

	for _, msg := range req.ConversationHistory[startIdx:] {
		role := openai.ChatMessageRoleUser
		if msg.Sender == "llm" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Message,
		})
	}

	// Add current user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.UserMessage,
	})

	// Call OpenAI API
	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       s.model,
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   500,
		},
	)

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		return nil, fmt.Errorf("failed to get response from LLM: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	return &ChatResponse{
		Message: resp.Choices[0].Message.Content,
	}, nil
}

// GenerateCareInstructions uses LLM to generate care instructions for a plant
func (s *ChatService) GenerateCareInstructions(ctx context.Context, genus, species string) (*db.CareGuide, error) {
	prompt := fmt.Sprintf(
		`Generate care instructions for the succulent plant: %s %s

Please provide specific care guidance in the following format (respond ONLY with valid JSON, no markdown formatting):

{
  "sunlight": "<detailed sunlight requirements>",
  "watering": "<detailed watering schedule and tips>",
  "soil": "<detailed soil requirements and recommendations>",
  "notes": "<additional care tips, growth patterns, or common issues>",
  "trivia": "<interesting facts, origin, cultural significance, or fun botanical trivia about this plant>"
}

Be specific, practical, and helpful. Include measurements and frequencies where relevant.`,
		genus,
		species,
	)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an expert botanist specializing in succulent plants. Provide accurate, detailed care instructions in JSON format.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	resp, err := s.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       s.model,
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   400,
		},
	)

	if err != nil {
		log.Printf("OpenAI API error while generating care instructions: %v", err)
		return nil, fmt.Errorf("failed to generate care instructions: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM for care instructions")
	}

	// Parse JSON response into CareGuide struct
	careGuide := &db.CareGuide{}
	content := resp.Choices[0].Message.Content

	// Try to parse JSON directly
	err = json.Unmarshal([]byte(content), careGuide)
	if err != nil {
		log.Printf("Failed to parse care instructions JSON: %v\nContent: %s", err, content)
		return nil, fmt.Errorf("failed to parse care instructions: %w", err)
	}

	log.Printf("Generated care instructions for %s %s", genus, species)
	return careGuide, nil
}

// buildSystemPrompt creates a system prompt with plant identification context
func (s *ChatService) buildSystemPrompt(identification *db.Identification) string {
	prompt := "You are a helpful succulent plant expert assistant. "

	if identification == nil {
		prompt += "Help the user with their questions about succulent plants."
		return prompt
	}

	prompt += fmt.Sprintf(
		"The user has identified a succulent plant. Here is the identification information:\n\n"+
			"Genus: %s\n",
		identification.Genus,
	)

	if identification.Species != "" {
		prompt += fmt.Sprintf("Species: %s\n", identification.Species)
	}

	prompt += fmt.Sprintf("Confidence: %.2f%%\n\n", identification.Confidence*100)

	if identification.CareGuide != nil {
		prompt += "Care Instructions:\n"
		if identification.CareGuide.Sunlight != "" {
			prompt += fmt.Sprintf("- Sunlight: %s\n", identification.CareGuide.Sunlight)
		}
		if identification.CareGuide.Watering != "" {
			prompt += fmt.Sprintf("- Watering: %s\n", identification.CareGuide.Watering)
		}
		if identification.CareGuide.Soil != "" {
			prompt += fmt.Sprintf("- Soil: %s\n", identification.CareGuide.Soil)
		}
		if identification.CareGuide.Notes != "" {
			prompt += fmt.Sprintf("- Notes: %s\n", identification.CareGuide.Notes)
		}
	}

	prompt += "\nAnswer the user's questions about this plant. Be concise, helpful, and friendly. " +
		"If asked about care, reference the care instructions provided above. " +
		"If you don't know something specific about this plant, be honest and provide general succulent care advice."

	return prompt
}
