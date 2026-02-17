package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// CareInstructionsRepository handles database operations for care instructions cache
type CareInstructionsRepository struct {
	db *sql.DB
}

// NewCareInstructionsRepository creates a new care instructions repository
func NewCareInstructionsRepository(db *sql.DB) *CareInstructionsRepository {
	return &CareInstructionsRepository{db: db}
}

// GetBySpecies retrieves cached care instructions for a specific genus and species
func (r *CareInstructionsRepository) GetBySpecies(genus, species string) (*CareInstructionsCache, error) {
	query := `
		SELECT id, genus, species, care_guide, created_at, updated_at
		FROM care_instructions
		WHERE genus = $1 AND species = $2
	`

	cache := &CareInstructionsCache{}
	var careGuideJSON []byte

	err := r.db.QueryRow(query, genus, species).Scan(
		&cache.ID,
		&cache.Genus,
		&cache.Species,
		&careGuideJSON,
		&cache.CreatedAt,
		&cache.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found in cache, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get care instructions: %w", err)
	}

	// Unmarshal care guide from JSON
	if len(careGuideJSON) > 0 {
		cache.CareGuide = &CareGuide{}
		if err := json.Unmarshal(careGuideJSON, cache.CareGuide); err != nil {
			return nil, fmt.Errorf("failed to unmarshal care guide: %w", err)
		}
	}

	return cache, nil
}

// Create saves new care instructions to the cache
func (r *CareInstructionsRepository) Create(cache *CareInstructionsCache) error {
	// Marshal care guide to JSON
	careGuideJSON, err := json.Marshal(cache.CareGuide)
	if err != nil {
		return fmt.Errorf("failed to marshal care guide: %w", err)
	}

	query := `
		INSERT INTO care_instructions (id, genus, species, care_guide, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (genus, species) DO UPDATE
		SET care_guide = EXCLUDED.care_guide,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		query,
		cache.ID,
		cache.Genus,
		cache.Species,
		careGuideJSON,
		cache.CreatedAt,
		cache.UpdatedAt,
	).Scan(&cache.ID, &cache.CreatedAt, &cache.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create care instructions: %w", err)
	}

	return nil
}

// Update updates existing care instructions in the cache
func (r *CareInstructionsRepository) Update(cache *CareInstructionsCache) error {
	// Marshal care guide to JSON
	careGuideJSON, err := json.Marshal(cache.CareGuide)
	if err != nil {
		return fmt.Errorf("failed to marshal care guide: %w", err)
	}

	query := `
		UPDATE care_instructions
		SET care_guide = $1, updated_at = $2
		WHERE genus = $3 AND species = $4
	`

	result, err := r.db.Exec(query, careGuideJSON, cache.UpdatedAt, cache.Genus, cache.Species)
	if err != nil {
		return fmt.Errorf("failed to update care instructions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("care instructions not found")
	}

	return nil
}
