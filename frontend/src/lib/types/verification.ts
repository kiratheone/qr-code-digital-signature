/**
 * Verification-related TypeScript interfaces
 */

export interface VerificationInfo {
  document_id: string;
  filename: string;
  issuer: string;
  // optional title of the document
  title?: string;
  // optional letter/issue number associated with the signed document
  letter_number?: string;
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
  // optional title of the document
  title?: string | null;
  // letter number included with the verification details when available
  letter_number?: string | null;
  };
  verified_at: string;
}

export type VerificationStatus = 'valid' | 'invalid' | 'modified';