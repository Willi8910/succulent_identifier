-- Drop indexes first
DROP INDEX IF EXISTS idx_chat_messages_id_created;
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_chat_messages_identification_id;
DROP INDEX IF EXISTS idx_identifications_created_at;

-- Drop tables
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS identifications;
