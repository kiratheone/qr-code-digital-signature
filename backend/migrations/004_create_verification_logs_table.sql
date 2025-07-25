-- Create verification_logs table
CREATE TABLE IF NOT EXISTS verification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID REFERENCES documents(id) ON DELETE CASCADE,
    verification_result VARCHAR(50) NOT NULL,
    verified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verifier_ip INET,
    details JSONB
);

-- Create indexes for optimal query performance
CREATE INDEX IF NOT EXISTS idx_verification_logs_document_id ON verification_logs(document_id);
CREATE INDEX IF NOT EXISTS idx_verification_logs_verified_at ON verification_logs(verified_at);
CREATE INDEX IF NOT EXISTS idx_verification_logs_result ON verification_logs(verification_result);
CREATE INDEX IF NOT EXISTS idx_verification_logs_verifier_ip ON verification_logs(verifier_ip);

-- Create composite index for document verification history
CREATE INDEX IF NOT EXISTS idx_verification_logs_doc_date ON verification_logs(document_id, verified_at);

-- Create GIN index for JSONB details column for efficient querying
CREATE INDEX IF NOT EXISTS idx_verification_logs_details ON verification_logs USING GIN (details);