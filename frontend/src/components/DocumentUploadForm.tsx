/**
 * DocumentUploadForm Component
 * Presentation component for uploading and signing PDF documents
 */

import React, { useState } from 'react';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { FileInput } from './ui/FileInput';

interface DocumentUploadFormProps {
  onUpload: (file: File, issuer: string) => void;
  isLoading?: boolean;
  error?: string | null;
  onErrorDismiss?: () => void;
}

export function DocumentUploadForm({
  onUpload,
  isLoading = false,
  error,
  onErrorDismiss,
}: DocumentUploadFormProps) {
  const [file, setFile] = useState<File | null>(null);
  const [issuer, setIssuer] = useState('');
  const [fileError, setFileError] = useState<string>('');
  const [issuerError, setIssuerError] = useState<string>('');

  const validateForm = (): boolean => {
    let isValid = true;
    
    // Reset errors
    setFileError('');
    setIssuerError('');

    // Validate file
    if (!file) {
      setFileError('Please select a PDF file to sign');
      isValid = false;
    }

    // Validate issuer
    if (!issuer.trim()) {
      setIssuerError('Issuer name is required');
      isValid = false;
    } else if (issuer.trim().length < 2) {
      setIssuerError('Issuer name must be at least 2 characters');
      isValid = false;
    }

    return isValid;
  };

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    if (file && issuer.trim()) {
      onUpload(file, issuer.trim());
    }
  };

  const handleFileSelect = (selectedFile: File | null) => {
    setFile(selectedFile);
    if (fileError && selectedFile) {
      setFileError('');
    }
  };

  const handleIssuerChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setIssuer(value);
    if (issuerError && value.trim()) {
      setIssuerError('');
    }
  };

  const handleReset = () => {
    setFile(null);
    setIssuer('');
    setFileError('');
    setIssuerError('');
    if (onErrorDismiss) {
      onErrorDismiss();
    }
  };

  return (
    <div className="bg-white shadow rounded-lg p-6">
      <div className="mb-6">
        <h2 className="text-lg font-medium text-gray-900">Sign PDF Document</h2>
        <p className="mt-1 text-sm text-gray-600">
          Upload a PDF document to create a digitally signed version with QR code verification.
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
                Document signing failed
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
          label="PDF Document"
          onFileSelect={handleFileSelect}
          accept=".pdf"
          maxSize={50 * 1024 * 1024} // 50MB
          error={fileError}
          helperText="Select a PDF file to digitally sign. Maximum file size: 50MB"
          disabled={isLoading}
        />

        <Input
          label="Issuer Name"
          type="text"
          value={issuer}
          onChange={handleIssuerChange}
          placeholder="Enter your name or organization"
          error={issuerError}
          helperText="This will appear on the digital signature as the document issuer"
          disabled={isLoading}
          required
        />

        <div className="flex justify-between">
          <Button
            type="button"
            variant="secondary"
            onClick={handleReset}
            disabled={isLoading || (!file && !issuer)}
          >
            Reset Form
          </Button>
          
          <Button
            type="submit"
            variant="primary"
            isLoading={isLoading}
            disabled={!file || !issuer.trim() || isLoading}
          >
            {isLoading ? 'Signing Document...' : 'Sign Document'}
          </Button>
        </div>
      </form>

      <div className="mt-6 text-xs text-gray-500">
        <p>
          By signing this document, you confirm that you have the authority to digitally sign it.
          The signed document will include a QR code for verification purposes.
        </p>
      </div>
    </div>
  );
}