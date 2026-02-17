-- Drop care_instructions table and its indexes
DROP INDEX IF EXISTS idx_care_instructions_created_at;
DROP INDEX IF EXISTS idx_care_instructions_genus_species;
DROP TABLE IF EXISTS care_instructions;
