package db

import (
	"database/sql"
	"fmt"
	"log"
)

// RunMigrations executes the database migrations
func RunMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Create identifications table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS identifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			genus VARCHAR(255) NOT NULL,
			species VARCHAR(255),
			confidence DECIMAL(5, 4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
			image_path TEXT NOT NULL,
			care_guide JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create identifications table: %w", err)
	}

	// Create index on identifications
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_identifications_created_at
		ON identifications(created_at DESC)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index on identifications: %w", err)
	}

	// Create chat_messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS chat_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			identification_id UUID NOT NULL REFERENCES identifications(id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			sender VARCHAR(10) NOT NULL CHECK (sender IN ('user', 'llm')),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create chat_messages table: %w", err)
	}

	// Create indexes on chat_messages
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_chat_messages_identification_id
		ON chat_messages(identification_id)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index on chat_messages: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at
		ON chat_messages(created_at)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index on chat_messages: %w", err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_chat_messages_id_created
		ON chat_messages(identification_id, created_at)
	`)
	if err != nil {
		return fmt.Errorf("failed to create composite index on chat_messages: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
