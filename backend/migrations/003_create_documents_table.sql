-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    issuer VARCHAR(255) NOT NULL,
    document_hash VARCHAR(64) NOT NULL,
    signature_data TEXT NOT NULL,
    qr_code_data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    file_size BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'active'
);

-- Create indexes for optimal query performance
CREATE INDEX IF NOT EXISTS idx_documents_hash ON documents(document_hash);
CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_created_at ON documents(created_at);
CREATE INDEX IF NOT EXISTS idx_documents_issuer ON documents(issuer);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_filename ON documents(filename);

-- Create composite index for user documents with status
CREATE INDEX IF NOT EXISTS idx_documents_user_status ON documents(user_id, status);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();