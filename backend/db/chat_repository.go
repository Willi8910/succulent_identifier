package db

import (
	"database/sql"
	"fmt"
)

// ChatRepository handles database operations for chat messages
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// Create saves a new chat message to the database
func (r *ChatRepository) Create(message *ChatMessage) error {
	query := `
		INSERT INTO chat_messages (id, identification_id, message, sender, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		message.ID,
		message.IdentificationID,
		message.Message,
		message.Sender,
		message.CreatedAt,
	).Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create chat message: %w", err)
	}

	return nil
}

// GetByIdentificationID retrieves all chat messages for a specific identification
func (r *ChatRepository) GetByIdentificationID(identificationID string) ([]ChatMessage, error) {
	query := `
		SELECT id, identification_id, message, sender, created_at
		FROM chat_messages
		WHERE identification_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, identificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var message ChatMessage
		err := rows.Scan(
			&message.ID,
			&message.IdentificationID,
			&message.Message,
			&message.Sender,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat messages: %w", err)
	}

	return messages, nil
}

// GetLatestMessages retrieves the N most recent messages for an identification
func (r *ChatRepository) GetLatestMessages(identificationID string, limit int) ([]ChatMessage, error) {
	query := `
		SELECT id, identification_id, message, sender, created_at
		FROM chat_messages
		WHERE identification_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, identificationID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest chat messages: %w", err)
	}
	defer rows.Close()

	messages := []ChatMessage{}
	for rows.Next() {
		var message ChatMessage
		err := rows.Scan(
			&message.ID,
			&message.IdentificationID,
			&message.Message,
			&message.Sender,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat messages: %w", err)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// CountByIdentificationID returns the number of messages for an identification
func (r *ChatRepository) CountByIdentificationID(identificationID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM chat_messages WHERE identification_id = $1`
	err := r.db.QueryRow(query, identificationID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count chat messages: %w", err)
	}
	return count, nil
}
