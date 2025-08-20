/**
 * Document Service - Business Logic Layer
 * Handles all document-related operations including signing, management, and retrieval
 */

import { ApiClient } from '@/lib/api';
import type {
  Document,
  SignDocumentRequest,
  SignDocumentResponse,
  DocumentList,
} from '@/lib/types';

export class DocumentService {
  constructor(private apiClient: ApiClient) {}

  /**
   * Sign a PDF document with digital signature
   */
  async signDocument(file: File, issuer: string): Promise<SignDocumentResponse> {
    // Validate input
    if (!file) {
      throw new Error('File is required');
    }
    if (!issuer.trim()) {
      throw new Error('Issuer name is required');
    }
    if (file.type !== 'application/pdf') {
      throw new Error('Only PDF files are supported');
    }
    if (file.size > 50 * 1024 * 1024) { // 50MB limit
      throw new Error('File size must be less than 50MB');
    }

    // Create form data for file upload
    const formData = new FormData();
    formData.append('file', file);
    formData.append('issuer', issuer.trim());

    return this.apiClient.post<SignDocumentResponse>('/documents/sign', formData);
  }

  /**
   * Get list of signed documents with pagination
   */
  async getDocuments(page: number = 1, perPage: number = 10): Promise<DocumentList> {
    if (page < 1) {
      throw new Error('Page number must be greater than 0');
    }
    if (perPage < 1 || perPage > 100) {
      throw new Error('Per page must be between 1 and 100');
    }

    const params = new URLSearchParams({
      page: page.toString(),
      per_page: perPage.toString(),
    });

    return this.apiClient.get<DocumentList>(`/documents?${params.toString()}`);
  }

  /**
   * Get a specific document by ID
   */
  async getDocumentById(documentId: string): Promise<Document> {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    return this.apiClient.get<Document>(`/documents/${documentId}`);
  }

  /**
   * Delete a document by ID
   */
  async deleteDocument(documentId: string): Promise<void> {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    return this.apiClient.delete<void>(`/documents/${documentId}`);
  }

  /**
   * Download signed PDF document
   */
  async downloadDocument(documentId: string): Promise<Blob> {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    const response = await fetch(`${this.apiClient['baseURL']}/api/documents/${documentId}/download`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${this.apiClient.getToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to download document');
    }

    return response.blob();
  }

  /**
   * Download QR code image for a document
   */
  async downloadQRCode(documentId: string): Promise<Blob> {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    const response = await fetch(`${this.apiClient['baseURL']}/api/documents/${documentId}/qr-code`, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${this.apiClient.getToken()}`,
      },
    });

    if (!response.ok) {
      throw new Error('Failed to download QR code');
    }

    return response.blob();
  }

  /**
   * Generate verification URL for QR code
   */
  generateVerificationUrl(documentId: string): string {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    // Get the current origin or use a default
    const origin = typeof window !== 'undefined' ? window.location.origin : 'https://your-domain.com';
    return `${origin}/verify/${documentId}`;
  }

  /**
   * Download file helper
   */
  downloadFile(blob: Blob, filename: string): void {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }

  /**
   * Validate file before upload
   */
  validateFile(file: File): { isValid: boolean; error?: string } {
    if (!file) {
      return { isValid: false, error: 'File is required' };
    }

    if (file.type !== 'application/pdf') {
      return { isValid: false, error: 'Only PDF files are supported' };
    }

    if (file.size > 50 * 1024 * 1024) { // 50MB limit
      return { isValid: false, error: 'File size must be less than 50MB' };
    }

    if (file.size === 0) {
      return { isValid: false, error: 'File cannot be empty' };
    }

    return { isValid: true };
  }

  /**
   * Format file size for display
   */
  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Format date for display
   */
  formatDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) {
        return 'Invalid date';
      }
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return 'Invalid date';
    }
  }
}