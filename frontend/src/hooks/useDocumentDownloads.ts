/**
 * Document Download Operations Hook
 * Handles downloading signed PDFs and QR codes
 */

import { useState } from 'react';
import { DocumentService } from '@/lib/services/DocumentService';
import { ApiClient } from '@/lib/api';

export function useDocumentDownloads() {
  const [isDownloading, setIsDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState<string | null>(null);

  const apiClient = new ApiClient();
  const documentService = new DocumentService(apiClient);

  const downloadPDF = async (documentId: string, filename: string) => {
    setIsDownloading(true);
    setDownloadError(null);

    try {
      const blob = await documentService.downloadDocument(documentId);
      const downloadFilename = filename.endsWith('.pdf') ? `signed_${filename}` : `signed_${filename}.pdf`;
      documentService.downloadFile(blob, downloadFilename);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to download PDF';
      setDownloadError(errorMessage);
      throw error;
    } finally {
      setIsDownloading(false);
    }
  };

  const downloadQRCode = async (documentId: string, filename: string) => {
    setIsDownloading(true);
    setDownloadError(null);

    try {
      const blob = await documentService.downloadQRCode(documentId);
      const baseName = filename.replace(/\.pdf$/i, '');
      const downloadFilename = `${baseName}_qr_code.png`;
      documentService.downloadFile(blob, downloadFilename);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to download QR code';
      setDownloadError(errorMessage);
      throw error;
    } finally {
      setIsDownloading(false);
    }
  };

  const getVerificationUrl = (documentId: string): string => {
    return documentService.generateVerificationUrl(documentId);
  };

  const copyVerificationUrl = async (documentId: string): Promise<void> => {
    const url = getVerificationUrl(documentId);
    
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(url);
    } else {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = url;
      textArea.style.position = 'fixed';
      textArea.style.left = '-999999px';
      textArea.style.top = '-999999px';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      document.execCommand('copy');
      textArea.remove();
    }
  };

  return {
    downloadPDF,
    downloadQRCode,
    getVerificationUrl,
    copyVerificationUrl,
    isDownloading,
    downloadError,
  };
}