-- Create identifications table
CREATE TABLE IF NOT EXISTS identifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    genus VARCHAR(255) NOT NULL,
    species VARCHAR(255),
    confidence DECIMAL(5, 4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    image_path TEXT NOT NULL,
    care_guide JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on created_at for faster queries
CREATE INDEX idx_identifications_created_at ON identifications(created_at DESC);

-- Create chat_messages table
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identification_id UUID NOT NULL REFERENCES identifications(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    sender VARCHAR(10) NOT NULL CHECK (sender IN ('user', 'llm')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for faster queries
CREATE INDEX idx_chat_messages_identification_id ON chat_messages(identification_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);

-- Create composite index for common query pattern
CREATE INDEX idx_chat_messages_id_created ON chat_messages(identification_id, created_at);
