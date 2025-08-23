/**
 * VerificationForm Component
 * Presentation component for document verification upload
 */

import React, { useState } from 'react';
import { Button } from './ui/Button';
import { FileInput } from './ui/FileInput';
import type { VerificationInfo } from '@/lib/types';

interface VerificationFormProps {
  documentInfo: VerificationInfo;
  onVerify: (file: File) => void;
  isLoading?: boolean;
  error?: string | null;
  onErrorDismiss?: () => void;
}

export function VerificationForm({
  documentInfo,
  onVerify,
  isLoading = false,
  error,
  onErrorDismiss,
}: VerificationFormProps) {
  const [file, setFile] = useState<File | null>(null);
  const [fileError, setFileError] = useState<string>('');

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!file) {
      setFileError('Please select a PDF file to verify');
      return;
    }

    onVerify(file);
  };

  const handleFileSelect = (selectedFile: File | null) => {
    setFile(selectedFile);
    if (fileError && selectedFile) {
      setFileError('');
    }
  };

  const formatDate = (dateString: string): string => {
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return 'Invalid date';
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* Document Information */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="mb-4">
          <h2 className="text-lg font-medium text-gray-900">Document Verification</h2>
          <p className="mt-1 text-sm text-gray-600">
            Verify the authenticity of this digitally signed document.
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <h3 className="text-sm font-medium text-gray-900 mb-3">Original Document Information</h3>
          <dl className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
            {documentInfo.title && (
              <div className="sm:col-span-2">
                <dt className="text-sm font-medium text-gray-500">Document Title</dt>
                <dd className="mt-1 text-sm text-gray-900 font-medium">{documentInfo.title}</dd>
              </div>
            )}
            <div>
              <dt className="text-sm font-medium text-gray-500">Filename</dt>
              <dd className="mt-1 text-sm text-gray-900">{documentInfo.filename}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Issuer</dt>
              <dd className="mt-1 text-sm text-gray-900">{documentInfo.issuer}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Letter Number</dt>
              <dd className="mt-1 text-sm text-gray-900">{documentInfo.letter_number && documentInfo.letter_number.trim() ? documentInfo.letter_number : 'Not provided'}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Signed Date</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatDate(documentInfo.created_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Document Hash</dt>
              <dd className="mt-1 text-sm text-gray-900 font-mono text-xs">
                {documentInfo.document_hash.substring(0, 32)}...
              </dd>
            </div>
          </dl>
        </div>
      </div>

      {/* Verification Upload */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="mb-6">
          <h3 className="text-lg font-medium text-gray-900">Upload Document for Verification</h3>
          <p className="mt-1 text-sm text-gray-600">
            Upload the PDF document you want to verify against the original signed version.
          </p>
        </div>

        {error && (
          <div className="mb-4 bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Verification failed
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error}</p>
                </div>
                {onErrorDismiss && (
                  <div className="mt-4">
                    <button
                      type="button"
                      onClick={onErrorDismiss}
                      className="text-sm font-medium text-red-800 hover:text-red-600"
                    >
                      Dismiss
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <FileInput
            label="PDF Document to Verify"
            onFileSelect={handleFileSelect}
            accept=".pdf"
            maxSize={50 * 1024 * 1024} // 50MB
            error={fileError}
            helperText="Select the PDF document you want to verify. This should be the same document that was originally signed."
            disabled={isLoading}
          />

          <div className="flex justify-end">
            <Button
              type="submit"
              variant="primary"
              isLoading={isLoading}
              disabled={!file || isLoading}
            >
              {isLoading ? 'Verifying Document...' : 'Verify Document'}
            </Button>
          </div>
        </form>

        <div className="mt-6 bg-blue-50 border border-blue-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">
                How verification works
              </h3>
              <div className="mt-2 text-sm text-blue-700">
                <p>
                  The system will calculate the hash of your uploaded document and compare it with the original signed document&apos;s hash. 
                  It will also verify the digital signature to ensure the document hasn&apos;t been tampered with.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}