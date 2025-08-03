/**
 * Verification-related TypeScript interfaces
 */

export interface VerificationInfo {
  document_id: string;
  filename: string;
  issuer: string;
  created_at: string;
  document_hash: string;
}

export interface VerifyDocumentRequest {
  document_id: string;
  file: File;
}

export interface VerificationResult {
  status: 'valid' | 'invalid' | 'modified';
  message: string;
  details: {
    qr_valid: boolean;
    hash_matches: boolean;
    signature_valid: boolean;
    original_hash: string;
    uploaded_hash: string;
  };
  verified_at: string;
}

export type VerificationStatus = 'valid' | 'invalid' | 'modified';