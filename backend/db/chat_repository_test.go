package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestChatRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	tests := []struct {
		name         string
		message      *ChatMessage
		mockBehavior func()
		expectError  bool
	}{
		{
			name: "Successful create - user message",
			message: &ChatMessage{
				ID:               "chat-id-1",
				IdentificationID: "plant-id-1",
				Message:          "What is the best watering schedule?",
				Sender:           "user",
				CreatedAt:        time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO chat_messages").
					WithArgs(
						sqlmock.AnyArg(), // id
						sqlmock.AnyArg(), // identification_id
						sqlmock.AnyArg(), // message
						sqlmock.AnyArg(), // sender
						sqlmock.AnyArg(), // created_at
					).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
						AddRow("chat-id-1", time.Now()))
			},
			expectError: false,
		},
		{
			name: "Successful create - llm message",
			message: &ChatMessage{
				ID:               "chat-id-2",
				IdentificationID: "plant-id-1",
				Message:          "Water once every 2 weeks in summer.",
				Sender:           "llm",
				CreatedAt:        time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO chat_messages").
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
						AddRow("chat-id-2", time.Now()))
			},
			expectError: false,
		},
		{
			name: "Database error",
			message: &ChatMessage{
				ID:               "chat-id-3",
				IdentificationID: "plant-id-1",
				Message:          "Test message",
				Sender:           "user",
				CreatedAt:        time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO chat_messages").
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			err := repo.Create(tt.message)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestChatRepositoryGetByIdentificationID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	tests := []struct {
		name         string
		plantID      string
		mockBehavior func()
		expectError  bool
		expectedLen  int
	}{
		{
			name:    "Successful retrieval with multiple messages",
			plantID: "plant-id-1",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "identification_id", "message", "sender", "created_at",
				}).
					AddRow("chat-1", "plant-id-1", "User question", "user", time.Now()).
					AddRow("chat-2", "plant-id-1", "LLM response", "llm", time.Now()).
					AddRow("chat-3", "plant-id-1", "Follow-up question", "user", time.Now())

				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-1").
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 3,
		},
		{
			name:    "No messages found",
			plantID: "plant-id-2",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "identification_id", "message", "sender", "created_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-2").
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 0,
		},
		{
			name:    "Database error",
			plantID: "plant-id-3",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-3").
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			result, err := repo.GetByIdentificationID(tt.plantID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d messages, got %d", tt.expectedLen, len(result))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestChatRepositoryGetLatestMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	tests := []struct {
		name         string
		plantID      string
		limit        int
		mockBehavior func()
		expectError  bool
		expectedLen  int
	}{
		{
			name:    "Get latest 5 messages",
			plantID: "plant-id-1",
			limit:   5,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "identification_id", "message", "sender", "created_at",
				}).
					AddRow("chat-5", "plant-id-1", "Latest", "user", time.Now()).
					AddRow("chat-4", "plant-id-1", "Message 4", "llm", time.Now()).
					AddRow("chat-3", "plant-id-1", "Message 3", "user", time.Now()).
					AddRow("chat-2", "plant-id-1", "Message 2", "llm", time.Now()).
					AddRow("chat-1", "plant-id-1", "Message 1", "user", time.Now())

				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id (.+) ORDER BY created_at DESC LIMIT").
					WithArgs("plant-id-1", 5).
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 5,
		},
		{
			name:    "Get latest 10 when only 3 exist",
			plantID: "plant-id-2",
			limit:   10,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "identification_id", "message", "sender", "created_at",
				}).
					AddRow("chat-3", "plant-id-2", "Message 3", "user", time.Now()).
					AddRow("chat-2", "plant-id-2", "Message 2", "llm", time.Now()).
					AddRow("chat-1", "plant-id-2", "Message 1", "user", time.Now())

				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id (.+) ORDER BY created_at DESC LIMIT").
					WithArgs("plant-id-2", 10).
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 3,
		},
		{
			name:    "Database error",
			plantID: "plant-id-3",
			limit:   5,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT (.+) FROM chat_messages WHERE identification_id (.+) ORDER BY created_at DESC LIMIT").
					WithArgs("plant-id-3", 5).
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			result, err := repo.GetLatestMessages(tt.plantID, tt.limit)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d messages, got %d", tt.expectedLen, len(result))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestChatRepositoryCountByIdentificationID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewChatRepository(db)

	tests := []struct {
		name          string
		plantID       string
		mockBehavior  func()
		expectError   bool
		expectedCount int
	}{
		{
			name:    "Count messages for existing chat",
			plantID: "plant-id-1",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(15)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-1").
					WillReturnRows(rows)
			},
			expectError:   false,
			expectedCount: 15,
		},
		{
			name:    "Count zero for new chat",
			plantID: "plant-id-2",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-2").
					WillReturnRows(rows)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:    "Database error",
			plantID: "plant-id-3",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM chat_messages WHERE identification_id").
					WithArgs("plant-id-3").
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			count, err := repo.CountByIdentificationID(tt.plantID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}
