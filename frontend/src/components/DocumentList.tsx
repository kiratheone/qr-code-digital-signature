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
            <div className="flex items-center justify-between">
              <div className="flex-1 min-w-0">
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
                    <div className="mt-1 flex items-center text-sm text-gray-500">
                      <span>Issued by {document.issuer}</span>
                      <span className="mx-2">•</span>
                      <span>{formatDate(document.created_at)}</span>
                      <span className="mx-2">•</span>
                      <span>{formatFileSize(document.file_size)}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="flex items-center space-x-2 ml-4">
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

                {onDownload && (
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={() => onDownload(document.id)}
                    disabled={isDownloadingDocument}
                    isLoading={isDownloadingDocument}
                  >
                    Download
                  </Button>
                )}

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
                >
                  Delete
                </Button>
              </div>
            </div>

            {/* Document metadata */}
            <div className="mt-3 text-xs text-gray-500">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="font-medium">Document ID:</span> {document.id}
                </div>
                <div>
                  <span className="font-medium">Hash:</span> {document.document_hash.substring(0, 16)}...
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}