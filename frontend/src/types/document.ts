export interface DocumentUploadRequest {
  file: File;
  issuer: string;
  description?: string;
  position?: QRCodePosition;
}

export interface QRCodePosition {
  page?: number; // Default: last page
  x?: number;    // Default: center
  y?: number;    // Default: bottom
}

export interface DocumentUploadResponse {
  id: string;
  filename: string;
  issuer: string;
  documentHash: string;
  createdAt: string;
  status: string;
}

export interface Document {
  id: string;
  filename: string;
  issuer: string;
  documentHash: string;
  createdAt: string;
  updatedAt: string;
  fileSize: number;
  status: string;
  qrCodeData?: string;
}

export interface DocumentListResponse {
  documents: Document[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface UploadProgressState {
  isUploading: boolean;
  progress: number;
  error?: string;
  success?: DocumentUploadResponse;
}
