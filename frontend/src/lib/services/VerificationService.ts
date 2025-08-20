/**
 * Verification Service - Business Logic Layer
 * Handles all document verification operations including QR code validation and document integrity checks
 */

import { ApiClient } from '@/lib/api';
import type {
  VerificationInfo,
  VerifyDocumentRequest,
  VerificationResult,
  VerificationStatus,
} from '@/lib/types';

export class VerificationService {
  constructor(private apiClient: ApiClient) {}

  /**
   * Get verification information for a document by ID
   */
  async getVerificationInfo(documentId: string): Promise<VerificationInfo> {
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }

    const response = await this.apiClient.get<{ verification_info: VerificationInfo }>(`/verify/${documentId}`);
    return response.verification_info || response as any;
  }

  /**
   * Verify a document by uploading it for comparison
   */
  async verifyDocument(documentId: string, file: File): Promise<VerificationResult> {
    // Validate input
    if (!documentId.trim()) {
      throw new Error('Document ID is required');
    }
    if (!file) {
      throw new Error('File is required for verification');
    }
    if (file.type !== 'application/pdf') {
      throw new Error('Only PDF files can be verified');
    }
    if (file.size > 50 * 1024 * 1024) { // 50MB limit
      throw new Error('File size must be less than 50MB');
    }

    // Create form data for file upload
    const formData = new FormData();
    formData.append('file', file);

    const response = await this.apiClient.post<{ verification_result: VerificationResult }>(`/verify/${documentId}/upload`, formData);
    return response.verification_result || response as any;
  }

  /**
   * Parse QR code data to extract document ID
   * This would typically be used with a QR code scanner library
   */
  parseQRCodeData(qrData: string): { documentId: string; isValid: boolean } {
    try {
      // Expected QR format: JSON with document metadata
      const data = JSON.parse(qrData);
      
      if (data.doc_id && typeof data.doc_id === 'string') {
        return {
          documentId: data.doc_id,
          isValid: true,
        };
      }

      return { documentId: '', isValid: false };
    } catch {
      // Try to extract document ID from URL format
      const urlMatch = qrData.match(/\/verify\/([a-f0-9-]+)/);
      if (urlMatch && urlMatch[1]) {
        return {
          documentId: urlMatch[1],
          isValid: true,
        };
      }

      return { documentId: '', isValid: false };
    }
  }

  /**
   * Get verification status display information
   */
  getVerificationStatusInfo(result: VerificationResult): {
    icon: string;
    color: string;
    title: string;
    description: string;
  } {
    switch (result.status) {
      case 'valid':
        return {
          icon: '✅',
          color: 'text-green-600',
          title: 'Document is Valid',
          description: 'The document is authentic and has not been modified since signing.',
        };
      
      case 'modified':
        return {
          icon: '⚠️',
          color: 'text-yellow-600',
          title: 'Document Modified',
          description: 'The QR code is valid, but the document content has been changed since signing.',
        };
      
      case 'invalid':
        return {
          icon: '❌',
          color: 'text-red-600',
          title: 'Document Invalid',
          description: 'The document signature is invalid or the QR code is corrupted.',
        };
      
      default:
        return {
          icon: '❓',
          color: 'text-gray-600',
          title: 'Unknown Status',
          description: 'Unable to determine document verification status.',
        };
    }
  }

  /**
   * Validate file before verification
   */
  validateFileForVerification(file: File): { isValid: boolean; error?: string } {
    if (!file) {
      return { isValid: false, error: 'File is required for verification' };
    }

    if (file.type !== 'application/pdf') {
      return { isValid: false, error: 'Only PDF files can be verified' };
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
   * Format verification date for display
   */
  formatVerificationDate(dateString: string): string {
    try {
      const date = new Date(dateString);
      return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return 'Invalid date';
    }
  }

  /**
   * Generate verification report summary
   */
  generateVerificationSummary(result: VerificationResult): {
    summary: string;
    details: Array<{ label: string; value: string; status: 'success' | 'warning' | 'error' }>;
  } {
    const details: Array<{ label: string; value: string; status: 'success' | 'warning' | 'error' }> = [
      {
        label: 'QR Code Validity',
        value: result.details.qr_valid ? 'Valid' : 'Invalid',
        status: result.details.qr_valid ? 'success' : 'error',
      },
      {
        label: 'Document Hash Match',
        value: result.details.hash_matches ? 'Matches' : 'Does not match',
        status: result.details.hash_matches ? 'success' : 'warning',
      },
      {
        label: 'Digital Signature',
        value: result.details.signature_valid ? 'Valid' : 'Invalid',
        status: result.details.signature_valid ? 'success' : 'error',
      },
    ];

    let summary: string;
    if (result.status === 'valid') {
      summary = 'Document verification successful. The document is authentic and unmodified.';
    } else if (result.status === 'modified') {
      summary = 'Document has been modified since signing. The signature is valid but content has changed.';
    } else {
      summary = 'Document verification failed. The document may be forged or corrupted.';
    }

    return { summary, details };
  }

  /**
   * Check if document ID format is valid (UUID format)
   */
  isValidDocumentId(documentId: string): boolean {
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    return uuidRegex.test(documentId);
  }

  /**
   * Create verification URL for sharing
   */
  createVerificationUrl(documentId: string, baseUrl?: string): string {
    const base = baseUrl || (typeof window !== 'undefined' ? window.location.origin : '');
    return `${base}/verify/${documentId}`;
  }
}