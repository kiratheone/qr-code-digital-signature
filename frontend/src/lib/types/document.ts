export interface Document {
  id: string;
  user_id: string;
  filename: string;
  issuer: string;
  title?: string;
  letter_number?: string;
  document_hash: string;
  signature_data: string;
  qr_code_data: string;
  created_at: string;
  updated_at: string;
  file_size: number;
  status: string;
}

export interface SignDocumentRequest {
  file: File;
  issuer: string;
  title: string; // Required for new documents
  letterNumber: string; // Required for new documents
}

export interface SignDocumentResponse {
  document: Document;
  download_url: string;
}

export interface DocumentList {
  documents: Document[];
  total: number;
  page: number;
  per_page: number;
}