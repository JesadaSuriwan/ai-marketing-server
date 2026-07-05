-- Add key_prefix column for safe display of API keys (first 8 chars of plaintext key)
-- key_value will now store SHA-256 hash of the plaintext key
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(20);

-- Backfill key_prefix from existing plaintext keys (first 8 chars)
UPDATE api_keys SET key_prefix = LEFT(key_value, 8) WHERE key_prefix IS NULL;
