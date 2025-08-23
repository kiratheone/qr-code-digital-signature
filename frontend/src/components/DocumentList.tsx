/**
 * DocumentList Component
 * Presentation component for displaying a list of signed documents
 */

import React from 'react';
import { Button } from './ui/Button';
import type { Document } from '@/lib/types';

interface DocumentListProps {
  documents: Document[];
  onDelete: (documentId: string) => void;
  onDownload?: (documentId: string) => void;
  onDownloadQR?: (documentId: string) => void;
  onCopyVerificationUrl?: (documentId: string) => void;
  isLoading?: boolean;
  isDeletingDocument?: boolean;
  isDownloadingDocument?: boolean;
  formatDate?: (dateString: string) => string;
  formatFileSize?: (bytes: number) => string;
}

export function DocumentList({
  documents,
  onDelete,
  onDownload,
  onDownloadQR,
  onCopyVerificationUrl,
  isLoading = false,
  isDeletingDocument = false,
  isDownloadingDocument = false,
  formatDate = (date) => new Date(date).toLocaleDateString(),
  formatFileSize = (bytes) => `${Math.round(bytes / 1024)} KB`,
}: DocumentListProps) {
  if (isLoading) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            {[...Array(3)].map((_, i) => (
              <div key={i} className="h-16 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <div className="text-center py-8">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No documents</h3>
          <p className="mt-1 text-sm text-gray-500">
            Get started by signing your first PDF document.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-6 py-4 border-b border-gray-200">
        <h2 className="text-lg font-medium text-gray-900">Signed Documents</h2>
        <p className="mt-1 text-sm text-gray-600">
          {documents.length} document{documents.length !== 1 ? 's' : ''} signed
        </p>
      </div>

      <div className="divide-y divide-gray-200">
        {documents.map((document) => (
          <div key={document.id} className="px-6 py-4 hover:bg-gray-50">
            <div className="flex flex-col gap-4">
              {/* Document Info Row */}
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <svg
                    className="h-8 w-8 text-red-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z"
                      clipRule="evenodd"
                      />
                    </svg>
                  </div>
                  <div className="ml-4 flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {document.filename}
                    </p>
                    {document.title && (
                      <p className="text-sm text-gray-700 truncate mt-1">
                        {document.title}
                      </p>
                    )}
                    <div className="mt-1 flex items-center text-sm text-gray-500">
                      <span>Issued by {document.issuer}</span>
                      {document.letter_number && (
                        <>
                          <span className="mx-2">•</span>
                          <span>#{document.letter_number}</span>
                        </>
                      )}
                      <span className="mx-2">•</span>
                      <span>{formatDate(document.created_at)}</span>
                      <span className="mx-2">•</span>
                      <span>{formatFileSize(document.file_size)}</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Actions Row - Better spacing and layout */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-2 border-t border-gray-100">
                {/* Status Badge */}
                <div className="flex items-center">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      document.status === 'active'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {document.status === 'active' ? 'Active' : 'Inactive'}
                  </span>
                </div>

                {/* Download and Delete Actions */}
                <div className="flex items-center gap-2 flex-wrap">
                  {onDownload && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onDownload(document.id)}
                      disabled={isDownloadingDocument}
                      isLoading={isDownloadingDocument}
                      title="Download signed PDF"
                      className="flex-shrink-0"
                    >
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                      PDF
                    </Button>
                  )}

                  {onDownloadQR && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onDownloadQR(document.id)}
                      disabled={isDownloadingDocument}
                      title="Download QR code"
                      className="flex-shrink-0"
                    >
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
                      </svg>
                      QR
                    </Button>
                  )}

                  {onCopyVerificationUrl && (
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onCopyVerificationUrl(document.id)}
                      title="Copy verification URL"
                      className="flex-shrink-0"
                    >
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                      URL
                    </Button>
                  )}

                  {/* Delete Button - Separated with divider */}
                  <div className="border-l border-gray-300 pl-2 ml-1">
                    <Button
                      variant="danger"
                      size="sm"
                      onClick={() => {
                        if (window.confirm(`Are you sure you want to delete "${document.filename}"? This action cannot be undone.`)) {
                          onDelete(document.id);
                        }
                      }}
                      disabled={isDeletingDocument}
                      isLoading={isDeletingDocument}
                      className="flex-shrink-0"
                    >
                      Delete
                    </Button>
                  </div>
                </div>

                {/* Document metadata */}
                <div className="mt-3 text-xs text-gray-500 border-t border-gray-100 pt-2">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                    <div>
                      <span className="font-medium">Document ID:</span> {document.id}
                    </div>
                    <div>
                      <span className="font-medium">Hash:</span> {document.document_hash.substring(0, 16)}...
                    </div>
                  </div>
                </div>
              </div>
          </div>
        ))}
      </div>
    </div>
  );
}