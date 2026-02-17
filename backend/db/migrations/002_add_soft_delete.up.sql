-- Add deleted_at column to identifications table for soft delete
ALTER TABLE identifications ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index on deleted_at for faster queries (filtering non-deleted records)
CREATE INDEX idx_identifications_deleted_at ON identifications(deleted_at);
