/**
 * DocumentUploadForm Component
 * Presentation component for uploading and signing PDF documents
 */

import React, { useState } from 'react';
import { Button } from './ui/Button';
import { Input } from './ui/Input';
import { FileInput } from './ui/FileInput';
import { ErrorMessage } from './ui/ErrorMessage';
import { useFormValidation, commonRules } from '@/hooks/useFormValidation';
import { ApiClientError } from '@/lib/api';

interface DocumentUploadFormProps {
  onUpload: (file: File, issuer: string) => void;
  isLoading?: boolean;
  error?: Error | ApiClientError | string | null;
  onErrorDismiss?: () => void;
}

export function DocumentUploadForm({
  onUpload,
  isLoading = false,
  error,
  onErrorDismiss,
}: DocumentUploadFormProps) {
  const [file, setFile] = useState<File | null>(null);
  
  // Form validation
  const {
    values,
    errors,
    handleChange,
    handleBlur,
    validateAll,
    reset,
    getFieldError,
    hasFieldError,
  } = useFormValidation(
    { issuer: '' },
    {
      issuer: {
        required: true,
        minLength: 2,
        maxLength: 100,
      },
    }
  );

  const validateFile = (selectedFile: File | null): string | null => {
    if (!selectedFile) {
      return 'Please select a PDF file to sign';
    }
    
    if (selectedFile.type !== 'application/pdf') {
      return 'Only PDF files are supported';
    }
    
    if (selectedFile.size > 50 * 1024 * 1024) {
      return 'File size must be less than 50MB';
    }
    
    if (selectedFile.size === 0) {
      return 'File cannot be empty';
    }
    
    return null;
  };

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    
    // Validate form
    const formValidation = validateAll();
    const fileValidationError = validateFile(file);
    
    if (!formValidation.isValid || fileValidationError) {
      return;
    }

    if (file && values.issuer.trim()) {
      onUpload(file, values.issuer.trim());
    }
  };

  const handleFileSelect = (selectedFile: File | null) => {
    setFile(selectedFile);
  };

  const handleReset = () => {
    setFile(null);
    reset();
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

      <ErrorMessage
        error={error || null}
        className="mb-4"
        onDismiss={onErrorDismiss}
        showDetails={process.env.NODE_ENV === 'development'}
      />

      <form onSubmit={handleSubmit} className="space-y-6">
        <FileInput
          label="PDF Document"
          onFileSelect={handleFileSelect}
          accept=".pdf"
          maxSize={50 * 1024 * 1024} // 50MB
          error={validateFile(file) || undefined}
          helperText="Select a PDF file to digitally sign. Maximum file size: 50MB"
          disabled={isLoading}
        />

        <Input
          label="Issuer Name"
          type="text"
          value={values.issuer}
          onChange={(e) => handleChange('issuer', e.target.value)}
          onBlur={() => handleBlur('issuer')}
          placeholder="Enter your name or organization"
          error={getFieldError('issuer') || undefined}
          helperText="This will appear on the digital signature as the document issuer"
          disabled={isLoading}
          required
        />

        <div className="flex justify-between">
          <Button
            type="button"
            variant="secondary"
            onClick={handleReset}
            disabled={isLoading || (!file && !values.issuer)}
          >
            Reset Form
          </Button>
          
          <Button
            type="submit"
            variant="primary"
            isLoading={isLoading}
            disabled={!file || !values.issuer.trim() || isLoading || hasFieldError('issuer') || !!validateFile(file)}
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