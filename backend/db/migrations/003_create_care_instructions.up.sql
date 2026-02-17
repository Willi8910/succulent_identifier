-- Create care_instructions table for caching LLM-generated care data
CREATE TABLE IF NOT EXISTS care_instructions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    genus VARCHAR(255) NOT NULL,
    species VARCHAR(255) NOT NULL,
    care_guide JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on genus+species combination
CREATE UNIQUE INDEX idx_care_instructions_genus_species ON care_instructions(genus, species);

-- Create index on created_at for tracking
CREATE INDEX idx_care_instructions_created_at ON care_instructions(created_at DESC);
