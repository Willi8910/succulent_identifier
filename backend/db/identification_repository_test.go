package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIdentificationRepositoryCreate(t *testing.T) {
	// Create mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewIdentificationRepository(db)

	tests := []struct {
		name           string
		identification *Identification
		mockBehavior   func()
		expectError    bool
	}{
		{
			name: "Successful create",
			identification: &Identification{
				ID:         "test-uuid-1",
				Genus:      "Haworthia",
				Species:    "zebrina",
				Confidence: 0.95,
				ImagePath:  "/uploads/test.jpg",
				CareGuide: &CareGuide{
					Sunlight: "Bright indirect light",
					Watering: "Water when dry",
					Soil:     "Well-draining",
					Notes:    "Easy care",
				},
				CreatedAt: time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO identifications").
					WithArgs(
						sqlmock.AnyArg(), // id
						sqlmock.AnyArg(), // genus
						sqlmock.AnyArg(), // species
						sqlmock.AnyArg(), // confidence
						sqlmock.AnyArg(), // image_path
						sqlmock.AnyArg(), // care_guide JSON
						sqlmock.AnyArg(), // created_at
					).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
						AddRow("test-uuid-1", time.Now()))
			},
			expectError: false,
		},
		{
			name: "Create with null care guide",
			identification: &Identification{
				ID:         "test-uuid-2",
				Genus:      "Aloe",
				Species:    "vera",
				Confidence: 0.85,
				ImagePath:  "/uploads/aloe.jpg",
				CareGuide:  nil,
				CreatedAt:  time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO identifications").
					WithArgs(
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						[]byte("null"), // JSON null
						sqlmock.AnyArg(),
					).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
						AddRow("test-uuid-2", time.Now()))
			},
			expectError: false,
		},
		{
			name: "Database error",
			identification: &Identification{
				ID:         "test-uuid-3",
				Genus:      "Test",
				Species:    "test",
				Confidence: 0.5,
				ImagePath:  "/uploads/test.jpg",
				CareGuide:  &CareGuide{},
				CreatedAt:  time.Now(),
			},
			mockBehavior: func() {
				mock.ExpectQuery("INSERT INTO identifications").
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			err := repo.Create(tt.identification)

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

func TestIdentificationRepositoryGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewIdentificationRepository(db)

	tests := []struct {
		name         string
		id           string
		mockBehavior func()
		expectError  bool
		expectNil    bool
	}{
		{
			name: "Successful retrieval",
			id:   "test-uuid-1",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "genus", "species", "confidence", "image_path", "care_guide", "created_at",
				}).AddRow(
					"test-uuid-1",
					"Haworthia",
					"zebrina",
					0.95,
					"/uploads/test.jpg",
					[]byte(`{"sunlight":"Bright indirect light","watering":"Water when dry","soil":"Well-draining","notes":"Easy care"}`),
					time.Now(),
				)

				mock.ExpectQuery("SELECT (.+) FROM identifications WHERE id").
					WithArgs("test-uuid-1").
					WillReturnRows(rows)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name: "Not found",
			id:   "non-existent",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT (.+) FROM identifications WHERE id").
					WithArgs("non-existent").
					WillReturnError(sql.ErrNoRows)
			},
			expectError: true,
			expectNil:   false,
		},
		{
			name: "Database error",
			id:   "test-uuid-2",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT (.+) FROM identifications WHERE id").
					WithArgs("test-uuid-2").
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			result, err := repo.GetByID(tt.id)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectNil && result != nil {
				t.Error("Expected nil result")
			}

			if !tt.expectNil && !tt.expectError && result == nil {
				t.Error("Expected non-nil result")
			}

			// Verify data if successful
			if !tt.expectError && result != nil {
				if result.ID == "" {
					t.Error("Expected non-empty ID")
				}
				if result.Genus == "" {
					t.Error("Expected non-empty genus")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestIdentificationRepositoryGetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewIdentificationRepository(db)

	tests := []struct {
		name         string
		limit        int
		offset       int
		mockBehavior func()
		expectError  bool
		expectedLen  int
	}{
		{
			name:   "Successful retrieval with limit and offset",
			limit:  10,
			offset: 0,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "genus", "species", "confidence", "image_path", "care_guide", "created_at",
				}).
					AddRow("id1", "Haworthia", "zebrina", 0.95, "/uploads/1.jpg", []byte(`{"sunlight":"test"}`), time.Now()).
					AddRow("id2", "Aloe", "vera", 0.85, "/uploads/2.jpg", []byte(`{"sunlight":"test"}`), time.Now())

				mock.ExpectQuery("SELECT (.+) FROM identifications ORDER BY created_at DESC LIMIT (.+) OFFSET").
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 2,
		},
		{
			name:   "Empty result",
			limit:  10,
			offset: 100,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{
					"id", "genus", "species", "confidence", "image_path", "care_guide", "created_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM identifications ORDER BY created_at DESC LIMIT (.+) OFFSET").
					WithArgs(10, 100).
					WillReturnRows(rows)
			},
			expectError: false,
			expectedLen: 0,
		},
		{
			name:   "Database error",
			limit:  10,
			offset: 0,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT (.+) FROM identifications ORDER BY created_at DESC LIMIT (.+) OFFSET").
					WithArgs(10, 0).
					WillReturnError(sql.ErrConnDone)
			},
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			result, err := repo.GetAll(tt.limit, tt.offset)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d results, got %d", tt.expectedLen, len(result))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestIdentificationRepositoryCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	repo := NewIdentificationRepository(db)

	tests := []struct {
		name          string
		mockBehavior  func()
		expectError   bool
		expectedCount int
	}{
		{
			name: "Successful count",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(42)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM identifications").
					WillReturnRows(rows)
			},
			expectError:   false,
			expectedCount: 42,
		},
		{
			name: "Zero count",
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM identifications").
					WillReturnRows(rows)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name: "Database error",
			mockBehavior: func() {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM identifications").
					WillReturnError(sql.ErrConnDone)
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			count, err := repo.Count()

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
