CREATE TABLE IF NOT EXISTS verification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    verification_result VARCHAR(50) NOT NULL,
    verified_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    verifier_ip INET,
    details JSONB
);

CREATE INDEX idx_verification_logs_document_id ON verification_logs(document_id);
CREATE INDEX idx_verification_logs_verified_at ON verification_logs(verified_at);
CREATE INDEX idx_verification_logs_result ON verification_logs(verification_result);