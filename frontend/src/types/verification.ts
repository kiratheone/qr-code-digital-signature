export interface VerificationRequest {
  file: File;
}

export interface VerificationResponse {
  documentId: string;
  filename: string;
  issuer: string;
  createdAt: string;
  status: VerificationStatus;
  message: string;
  details?: string;
}

export type VerificationStatus = 'valid' | 'modified' | 'invalid' | 'pending';

export interface VerificationState {
  isVerifying: boolean;
  progress: number;
  error?: string;
  result?: VerificationResponse;
}