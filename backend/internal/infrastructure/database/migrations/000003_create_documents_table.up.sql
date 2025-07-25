CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
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

CREATE INDEX idx_documents_hash ON documents(document_hash);
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_issuer ON documents(issuer);
CREATE INDEX idx_documents_created_at ON documents(created_at);