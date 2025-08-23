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
  onUpload: (file: File, issuer: string, title: string, letterNumber: string) => void;
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
    { issuer: '', title: '', letterNumber: '' },
    {
      issuer: {
        required: true,
        minLength: 2,
        maxLength: 100,
      },
      title: {
        required: true,
        minLength: 1,
        maxLength: 200,
      },
      letterNumber: {
        required: true,
        minLength: 1,
        maxLength: 50,
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

    if (file && values.issuer.trim() && values.title.trim() && values.letterNumber.trim()) {
      onUpload(file, values.issuer.trim(), values.title.trim(), values.letterNumber.trim());
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
    <div className="bg-white shadow rounded-lg p-4">
      <div className="mb-4">
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

      <form onSubmit={handleSubmit} className="space-y-4">
        <FileInput
          label="PDF Document"
          onFileSelect={handleFileSelect}
          accept=".pdf"
          maxSize={50 * 1024 * 1024} // 50MB
          error={validateFile(file) || undefined}
          helperText="Click to upload or drag and drop PDF files up to 50MB"
          disabled={isLoading}
        />

        <div className="grid grid-cols-1 gap-4">
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

          <Input
            label="Document Title"
            type="text"
            value={values.title}
            onChange={(e) => handleChange('title', e.target.value)}
            onBlur={() => handleBlur('title')}
            placeholder="Enter document title"
            error={getFieldError('title') || undefined}
            helperText="Title of the document being signed"
            disabled={isLoading}
            required
          />

          <Input
            label="Letter Number"
            type="text"
            value={values.letterNumber}
            onChange={(e) => handleChange('letterNumber', e.target.value)}
            onBlur={() => handleBlur('letterNumber')}
            placeholder="Enter letter number (e.g., 001/2025)"
            error={getFieldError('letterNumber') || undefined}
            helperText="Document reference number for identification"
            disabled={isLoading}
            required
          />
        </div>

        <div className="flex flex-col sm:flex-row justify-between gap-3">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={handleReset}
            disabled={isLoading || (!file && !values.issuer && !values.title)}
          >
            Reset Form
          </Button>
          
          <Button
            type="submit"
            variant="primary"
            size="sm"
            isLoading={isLoading}
            disabled={!file || !values.issuer.trim() || !values.title.trim() || !values.letterNumber.trim() || isLoading || hasFieldError('issuer') || hasFieldError('title') || hasFieldError('letterNumber') || !!validateFile(file)}
          >
            {isLoading ? 'Signing Document...' : 'Sign Document'}
          </Button>
        </div>
      </form>

      <div className="mt-4 text-xs text-gray-500">
        <p>
          By signing this document, you confirm that you have the authority to digitally sign it.
          The signed document will include a QR code for verification purposes.
        </p>
      </div>
    </div>
  );
}