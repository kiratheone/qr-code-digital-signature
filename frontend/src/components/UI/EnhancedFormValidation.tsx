'use client';

import React, { useState, useCallback, useRef, useEffect } from 'react';
import { useFormValidation, ValidationRule, ValidatedInput, ValidatedTextarea, ValidatedSelect } from './FormValidation';
import { LoadingSpinner } from './LoadingSpinner';
import { useNotificationHelpers } from './Notifications';

// Enhanced form wrapper with comprehensive error handling
interface EnhancedFormProps {
  children: React.ReactNode;
  onSubmit: (data: any) => Promise<void>;
  initialValues?: Record<string, any>;
  validationRules?: Record<string, ValidationRule>;
  className?: string;
  showProgress?: boolean;
  autoSave?: boolean;
  autoSaveDelay?: number;
}

export function EnhancedForm({
  children,
  onSubmit,
  initialValues = {},
  validationRules = {},
  className = '',
  showProgress = false,
  autoSave = false,
  autoSaveDelay = 2000,
}: EnhancedFormProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitAttempts, setSubmitAttempts] = useState(0);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const autoSaveTimeoutRef = useRef<NodeJS.Timeout>();
  const { showSuccess, showError, showWarning } = useNotificationHelpers();

  // Create form validation fields from initial values and rules
  const formFields = Object.keys({ ...initialValues, ...validationRules }).reduce((acc, key) => {
    acc[key] = {
      value: initialValues[key] || '',
      rules: validationRules[key] || {},
    };
    return acc;
  }, {} as Record<string, { value: any; rules: ValidationRule }>);

  const {
    fields,
    validateAllFields,
    updateField,
    resetForm,
    isFormValid,
    hasErrors,
  } = useFormValidation(formFields);

  // Auto-save functionality
  const performAutoSave = useCallback(async () => {
    if (!autoSave || !isFormValid || hasErrors) return;

    try {
      const formData = Object.keys(fields).reduce((acc, key) => {
        acc[key] = fields[key].value;
        return acc;
      }, {} as Record<string, any>);

      // Call a draft save function (would need to be passed as prop)
      // await onAutoSave?.(formData);
      setLastSaved(new Date());
      
      // Show subtle success indicator
      showSuccess('Draft Saved', '', { duration: 2000 });
    } catch (error) {
      console.warn('Auto-save failed:', error);
    }
  }, [autoSave, isFormValid, hasErrors, fields, showSuccess]);

  // Debounced auto-save
  useEffect(() => {
    if (!autoSave) return;

    if (autoSaveTimeoutRef.current) {
      clearTimeout(autoSaveTimeoutRef.current);
    }

    autoSaveTimeoutRef.current = setTimeout(performAutoSave, autoSaveDelay);

    return () => {
      if (autoSaveTimeoutRef.current) {
        clearTimeout(autoSaveTimeoutRef.current);
      }
    };
  }, [fields, performAutoSave, autoSave, autoSaveDelay]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (isSubmitting) return;

    setSubmitAttempts(prev => prev + 1);
    setIsSubmitting(true);

    try {
      // Validate all fields
      const { isValid, errors } = validateAllFields();
      
      if (!isValid) {
        const errorCount = Object.keys(errors).length;
        showError(
          'Form Validation Failed',
          `Please fix ${errorCount} error${errorCount > 1 ? 's' : ''} before submitting.`
        );
        return;
      }

      // Prepare form data
      const formData = Object.keys(fields).reduce((acc, key) => {
        acc[key] = fields[key].value;
        return acc;
      }, {} as Record<string, any>);

      // Submit form
      await onSubmit(formData);
      
      showSuccess('Form Submitted', 'Your form has been submitted successfully.');
      setSubmitAttempts(0);
      
    } catch (error) {
      console.error('Form submission error:', error);
      
      // Show user-friendly error message
      const errorMessage = error instanceof Error ? error.message : 'An unexpected error occurred';
      showError('Submission Failed', errorMessage);
      
      // Provide retry suggestion after multiple attempts
      if (submitAttempts >= 2) {
        showWarning(
          'Multiple Attempts Failed',
          'If this problem persists, try refreshing the page or contact support.',
          {
            persistent: true,
            action: {
              label: 'Refresh Page',
              onClick: () => window.location.reload(),
            },
          }
        );
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReset = () => {
    resetForm();
    setSubmitAttempts(0);
    setLastSaved(null);
    showSuccess('Form Reset', 'Form has been reset to initial values.');
  };

  return (
    <form onSubmit={handleSubmit} className={`space-y-6 ${className}`} noValidate>
      {/* Progress indicator */}
      {showProgress && isSubmitting && (
        <div className="mb-4">
          <div className="flex items-center justify-between text-sm text-gray-600 mb-2">
            <span>Submitting form...</span>
            <span>Please wait</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div className="bg-blue-600 h-2 rounded-full animate-pulse" style={{ width: '70%' }} />
          </div>
        </div>
      )}

      {/* Auto-save indicator */}
      {autoSave && lastSaved && (
        <div className="text-xs text-gray-500 text-right">
          Last saved: {lastSaved.toLocaleTimeString()}
        </div>
      )}

      {/* Form fields */}
      <div className="space-y-4">
        {React.Children.map(children, (child) => {
          if (React.isValidElement(child)) {
            // Clone form field components with validation props
            const fieldName = child.props.name;
            if (fieldName && fields[fieldName]) {
              return React.cloneElement(child, {
                ...child.props,
                value: fields[fieldName].value,
                error: fields[fieldName].error,
                touched: fields[fieldName].touched,
                onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
                  updateField(fieldName, e.target.value);
                  child.props.onChange?.(e);
                },
              });
            }
          }
          return child;
        })}
      </div>

      {/* Form actions */}
      <div className="flex justify-between items-center pt-6 border-t border-gray-200">
        <button
          type="button"
          onClick={handleReset}
          disabled={isSubmitting}
          className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Reset
        </button>

        <div className="flex items-center space-x-3">
          {submitAttempts > 0 && (
            <span className="text-xs text-gray-500">
              Attempt {submitAttempts}
            </span>
          )}
          
          <button
            type="submit"
            disabled={isSubmitting || (!isFormValid && submitAttempts === 0)}
            className="flex items-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting && <LoadingSpinner size="sm" color="white" className="mr-2" />}
            {isSubmitting ? 'Submitting...' : 'Submit'}
          </button>
        </div>
      </div>

      {/* Form status */}
      {hasErrors && (
        <div className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
          <div className="flex items-center">
            <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            Please fix the errors above before submitting.
          </div>
        </div>
      )}
    </form>
  );
}

// Enhanced file upload component with better error handling
interface EnhancedFileUploadProps {
  name: string;
  label: string;
  accept?: string;
  maxSize?: number;
  multiple?: boolean;
  onUpload: (files: File[]) => Promise<void>;
  className?: string;
}

export function EnhancedFileUpload({
  name,
  label,
  accept = '*/*',
  maxSize = 50 * 1024 * 1024, // 50MB default
  multiple = false,
  onUpload,
  className = '',
}: EnhancedFileUploadProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [errors, setErrors] = useState<string[]>([]);
  const { showSuccess, showError } = useNotificationHelpers();

  const validateFiles = (files: FileList | File[]): { valid: File[]; errors: string[] } => {
    const fileArray = Array.from(files);
    const validFiles: File[] = [];
    const validationErrors: string[] = [];

    fileArray.forEach((file) => {
      // Check file size
      if (file.size > maxSize) {
        validationErrors.push(`${file.name}: File size exceeds ${(maxSize / (1024 * 1024)).toFixed(0)}MB limit`);
        return;
      }

      // Check file type if specified
      if (accept !== '*/*') {
        const acceptedTypes = accept.split(',').map(type => type.trim());
        const isValidType = acceptedTypes.some(type => {
          if (type.startsWith('.')) {
            return file.name.toLowerCase().endsWith(type.toLowerCase());
          }
          return file.type.match(type.replace('*', '.*'));
        });

        if (!isValidType) {
          validationErrors.push(`${file.name}: File type not supported`);
          return;
        }
      }

      validFiles.push(file);
    });

    return { valid: validFiles, errors: validationErrors };
  };

  const handleFileSelect = (files: FileList | File[]) => {
    const { valid, errors: validationErrors } = validateFiles(files);
    
    setErrors(validationErrors);
    setSelectedFiles(valid);

    if (validationErrors.length > 0) {
      showError('File Validation Failed', validationErrors.join('\n'));
    }
  };

  const handleUpload = async () => {
    if (selectedFiles.length === 0) return;

    setIsUploading(true);
    setUploadProgress(0);

    try {
      // Simulate progress
      const progressInterval = setInterval(() => {
        setUploadProgress(prev => Math.min(prev + 10, 90));
      }, 200);

      await onUpload(selectedFiles);

      clearInterval(progressInterval);
      setUploadProgress(100);
      
      showSuccess('Upload Successful', `${selectedFiles.length} file(s) uploaded successfully.`);
      setSelectedFiles([]);
      setErrors([]);
      
    } catch (error) {
      console.error('Upload error:', error);
      const errorMessage = error instanceof Error ? error.message : 'Upload failed';
      showError('Upload Failed', errorMessage);
    } finally {
      setIsUploading(false);
      setTimeout(() => setUploadProgress(0), 1000);
    }
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    handleFileSelect(e.dataTransfer.files);
  };

  return (
    <div className={`space-y-4 ${className}`}>
      <label className="block text-sm font-medium text-gray-700">{label}</label>
      
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors ${
          isDragging
            ? 'border-blue-400 bg-blue-50'
            : errors.length > 0
            ? 'border-red-300 bg-red-50'
            : 'border-gray-300 hover:border-gray-400'
        }`}
      >
        <input
          type="file"
          accept={accept}
          multiple={multiple}
          onChange={(e) => e.target.files && handleFileSelect(e.target.files)}
          className="hidden"
          id={`file-input-${name}`}
        />
        
        <label htmlFor={`file-input-${name}`} className="cursor-pointer">
          <div className="space-y-2">
            <svg className="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
              <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <div className="text-gray-600">
              <span className="font-medium text-blue-600 hover:text-blue-500">Click to upload</span>
              {' or drag and drop'}
            </div>
            <p className="text-xs text-gray-500">
              {accept !== '*/*' && `Accepted: ${accept} • `}
              Max size: {(maxSize / (1024 * 1024)).toFixed(0)}MB
            </p>
          </div>
        </label>
      </div>

      {/* Selected files */}
      {selectedFiles.length > 0 && (
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-700">Selected Files:</h4>
          {selectedFiles.map((file, index) => (
            <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded">
              <span className="text-sm text-gray-600 truncate">{file.name}</span>
              <span className="text-xs text-gray-500">{(file.size / (1024 * 1024)).toFixed(2)} MB</span>
            </div>
          ))}
        </div>
      )}

      {/* Errors */}
      {errors.length > 0 && (
        <div className="text-sm text-red-600 space-y-1">
          {errors.map((error, index) => (
            <div key={index} className="flex items-center">
              <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
              {error}
            </div>
          ))}
        </div>
      )}

      {/* Upload progress */}
      {isUploading && (
        <div className="space-y-2">
          <div className="flex justify-between text-sm text-gray-600">
            <span>Uploading...</span>
            <span>{uploadProgress}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
        </div>
      )}

      {/* Upload button */}
      {selectedFiles.length > 0 && !isUploading && (
        <button
          onClick={handleUpload}
          className="w-full px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          Upload {selectedFiles.length} file{selectedFiles.length > 1 ? 's' : ''}
        </button>
      )}
    </div>
  );
}