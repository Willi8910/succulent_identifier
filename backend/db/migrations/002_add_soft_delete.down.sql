-- Drop the deleted_at column and its index
DROP INDEX IF EXISTS idx_identifications_deleted_at;
ALTER TABLE identifications DROP COLUMN deleted_at;
