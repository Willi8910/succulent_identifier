package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// IdentificationRepository handles database operations for identifications
type IdentificationRepository struct {
	db *sql.DB
}

// NewIdentificationRepository creates a new identification repository
func NewIdentificationRepository(db *sql.DB) *IdentificationRepository {
	return &IdentificationRepository{db: db}
}

// Create saves a new identification to the database
func (r *IdentificationRepository) Create(identification *Identification) error {
	// Marshal care guide to JSON
	careGuideJSON, err := json.Marshal(identification.CareGuide)
	if err != nil {
		return fmt.Errorf("failed to marshal care guide: %w", err)
	}

	query := `
		INSERT INTO identifications (id, genus, species, confidence, image_path, care_guide, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err = r.db.QueryRow(
		query,
		identification.ID,
		identification.Genus,
		identification.Species,
		identification.Confidence,
		identification.ImagePath,
		careGuideJSON,
		identification.CreatedAt,
	).Scan(&identification.ID, &identification.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create identification: %w", err)
	}

	return nil
}

// GetByID retrieves an identification by ID (excludes soft-deleted records)
func (r *IdentificationRepository) GetByID(id string) (*Identification, error) {
	query := `
		SELECT id, genus, species, confidence, image_path, care_guide, created_at
		FROM identifications
		WHERE id = $1 AND deleted_at IS NULL
	`

	identification := &Identification{}
	var careGuideJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&identification.ID,
		&identification.Genus,
		&identification.Species,
		&identification.Confidence,
		&identification.ImagePath,
		&careGuideJSON,
		&identification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("identification not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get identification: %w", err)
	}

	// Unmarshal care guide from JSON
	if len(careGuideJSON) > 0 {
		identification.CareGuide = &CareGuide{}
		if err := json.Unmarshal(careGuideJSON, identification.CareGuide); err != nil {
			return nil, fmt.Errorf("failed to unmarshal care guide: %w", err)
		}
	}

	return identification, nil
}

// GetAll retrieves all identifications ordered by creation date (newest first)
// Excludes soft-deleted records
func (r *IdentificationRepository) GetAll(limit, offset int) ([]Identification, error) {
	query := `
		SELECT id, genus, species, confidence, image_path, care_guide, created_at
		FROM identifications
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get identifications: %w", err)
	}
	defer rows.Close()

	identifications := []Identification{}
	for rows.Next() {
		var identification Identification
		var careGuideJSON []byte

		err := rows.Scan(
			&identification.ID,
			&identification.Genus,
			&identification.Species,
			&identification.Confidence,
			&identification.ImagePath,
			&careGuideJSON,
			&identification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan identification: %w", err)
		}

		// Unmarshal care guide from JSON
		if len(careGuideJSON) > 0 {
			identification.CareGuide = &CareGuide{}
			if err := json.Unmarshal(careGuideJSON, identification.CareGuide); err != nil {
				return nil, fmt.Errorf("failed to unmarshal care guide: %w", err)
			}
		}

		identifications = append(identifications, identification)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating identifications: %w", err)
	}

	return identifications, nil
}

// Count returns the total number of non-deleted identifications
func (r *IdentificationRepository) Count() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM identifications WHERE deleted_at IS NULL`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count identifications: %w", err)
	}
	return count, nil
}

// Delete performs a soft delete by setting deleted_at timestamp
func (r *IdentificationRepository) Delete(id string) error {
	query := `
		UPDATE identifications
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete identification: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("identification not found")
	}

	return nil
}
