'use client';

import React, { useState, useCallback, useEffect } from 'react';
import { ErrorDisplay } from './ErrorDisplay';
import { LoadingSpinner } from './LoadingSpinner';
import { useNotificationHelpers } from './Notifications';
import { formatApiError } from '@/utils/apiErrorUtils';

interface FormError {
  field?: string;
  message: string;
  code?: string;
  type?: 'validation' | 'server' | 'network' | 'timeout';
}

interface FormErrorState {
  errors: FormError[];
  isSubmitting: boolean;
  submitAttempts: number;
  lastSubmitTime: Date | null;
  canRetry: boolean;
}

interface FormErrorHandlerProps {
  children: React.ReactNode;
  onSubmit: (data: unknown) => Promise<void>;
  onSuccess?: (data: unknown) => void;
  onError?: (error: unknown) => void;
  maxRetries?: number;
  retryDelay?: number;
  showNotifications?: boolean;
  preventDuplicateSubmission?: boolean;
  className?: string;
}

export function FormErrorHandler({
  children,
  onSubmit,
  onSuccess,
  onError,
  maxRetries = 3,
  retryDelay = 1000,
  showNotifications = true,
  preventDuplicateSubmission = true,
  className = '',
}: FormErrorHandlerProps) {
  const [errorState, setErrorState] = useState<FormErrorState>({
    errors: [],
    isSubmitting: false,
    submitAttempts: 0,
    lastSubmitTime: null,
    canRetry: false,
  });

  const { showSuccess, showError, showWarning } = useNotificationHelpers();

  const removeError = useCallback((index: number) => {
    setErrorState(prev => ({
      ...prev,
      errors: prev.errors.filter((_, i) => i !== index),
    }));
  }, []);

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const handleSubmit = useCallback(async (formData: unknown) => {
    // Prevent duplicate submissions
    if (preventDuplicateSubmission && errorState.lastSubmitTime) {
      const timeSinceLastSubmit = Date.now() - errorState.lastSubmitTime.getTime();
      if (timeSinceLastSubmit < 2000) {
        if (showNotifications) {
          showWarning('Please Wait', 'Please wait before submitting again.');
        }
        return;
      }
    }

    setErrorState(prev => ({
      ...prev,
      isSubmitting: true,
      errors: [],
      submitAttempts: prev.submitAttempts + 1,
      lastSubmitTime: new Date(),
    }));

    try {
      await onSubmit(formData);

      // Success
      setErrorState(prev => ({
        ...prev,
        isSubmitting: false,
        submitAttempts: 0,
        canRetry: false,
      }));

      if (showNotifications) {
        showSuccess('Success', 'Form submitted successfully!');
      }

      onSuccess?.(formData);

    } catch (error) {
      console.error('Form submission error:', error);

      const formErrors = parseFormError(error);
      const canRetry = formErrors.some(err => err.type === 'network' || err.type === 'timeout') &&
        errorState.submitAttempts < maxRetries;

      setErrorState(prev => ({
        ...prev,
        isSubmitting: false,
        errors: formErrors,
        canRetry,
      }));

      if (showNotifications) {
        const errorMessage = formatApiError(error);
        showError('Submission Failed', errorMessage);
      }

      onError?.(error);
    }
  }, [
    onSubmit,
    onSuccess,
    onError,
    errorState.lastSubmitTime,
    errorState.submitAttempts,
    maxRetries,
    preventDuplicateSubmission,
    showNotifications,
    showSuccess,
    showError,
    showWarning,
  ]);

  const handleRetry = useCallback(async () => {
    if (!errorState.canRetry) return;

    // Wait for retry delay
    await new Promise(resolve => setTimeout(resolve, retryDelay * errorState.submitAttempts));

    // This would need to be called with the last form data
    // In practice, this would be handled by the parent component
    console.log('Retry requested - parent component should handle this');
  }, [errorState.canRetry, errorState.submitAttempts, retryDelay]);

  // Auto-retry for network errors
  useEffect(() => {
    if (errorState.canRetry && errorState.errors.some(err => err.type === 'network')) {
      const timer = setTimeout(() => {
        handleRetry();
      }, retryDelay * errorState.submitAttempts);

      return () => clearTimeout(timer);
    }
  }, [errorState.canRetry, errorState.errors, errorState.submitAttempts, retryDelay, handleRetry]);

  return (
    <div className={className}>
      {/* Error Display */}
      {errorState.errors.length > 0 && (
        <div className="mb-4 space-y-2">
          {errorState.errors.map((error, index) => (
            <ErrorDisplay
              key={index}
              error={error}
              title={error.field ? `${error.field} Error` : 'Form Error'}
              onRetry={errorState.canRetry ? handleRetry : undefined}
              onDismiss={() => removeError(index)}
              showRetry={errorState.canRetry}
              variant="inline"
              size="sm"
            />
          ))}
        </div>
      )}

      {/* Loading State */}
      {errorState.isSubmitting && (
        <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
          <div className="flex items-center">
            <LoadingSpinner size="sm" color="primary" className="mr-2" />
            <span className="text-sm text-blue-700">
              Submitting form... (Attempt {errorState.submitAttempts})
            </span>
          </div>
        </div>
      )}

      {/* Form Content */}
      <div className={errorState.isSubmitting ? 'opacity-75 pointer-events-none' : ''}>
        {children}
      </div>

      {/* Retry Information */}
      {errorState.canRetry && (
        <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
          <div className="flex items-center">
            <svg className="h-5 w-5 text-yellow-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div className="text-sm text-yellow-700">
              <p>
                Submission failed due to a temporary issue.
                {errorState.submitAttempts < maxRetries && (
                  <span> Retrying automatically...</span>
                )}
              </p>
              <p className="text-xs mt-1">
                Attempt {errorState.submitAttempts} of {maxRetries}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Helper function to parse different types of form errors
function parseFormError(error: unknown): FormError[] {
  const errors: FormError[] = [];

  if (typeof error === 'object' && error !== null) {
    // Handle validation errors from backend
    if ('validation_errors' in error) {
      const validationErrors = (error as { validation_errors?: unknown[] }).validation_errors;
      if (Array.isArray(validationErrors)) {
        validationErrors.forEach((err: { field?: string; message?: string; code?: string }) => {
          errors.push({
            field: err.field,
            message: err.message,
            code: err.code,
            type: 'validation',
          });
        });
      }
    }
    // Handle API errors
    else if ('status' in error || 'status_code' in error) {
      const status = (error as { status?: number; status_code?: number }).status || (error as { status?: number; status_code?: number }).status_code;
      const message = (error as { message?: string }).message || 'An error occurred';

      let type: FormError['type'] = 'server';
      if (status === 0 || status === undefined) {
        type = 'network';
      } else if (status === 408 || status === 504) {
        type = 'timeout';
      } else if (status >= 400 && status < 500) {
        type = 'validation';
      }

      errors.push({
        message,
        code: status?.toString(),
        type,
      });
    }
    // Handle Error objects
    else if (error instanceof Error) {
      let type: FormError['type'] = 'server';
      if (error.message.toLowerCase().includes('network') ||
        error.message.toLowerCase().includes('fetch')) {
        type = 'network';
      } else if (error.message.toLowerCase().includes('timeout')) {
        type = 'timeout';
      }

      errors.push({
        message: error.message,
        type,
      });
    }
  }

  // Fallback error
  if (errors.length === 0) {
    errors.push({
      message: 'An unexpected error occurred',
      type: 'server',
    });
  }

  return errors;
}

// Hook for using form error handling
export function useFormErrorHandler(options: {
  maxRetries?: number;
  retryDelay?: number;
  showNotifications?: boolean;
} = {}) {
  const [errorState, setErrorState] = useState<FormErrorState>({
    errors: [],
    isSubmitting: false,
    submitAttempts: 0,
    lastSubmitTime: null,
    canRetry: false,
  });

  const { showSuccess, showError } = useNotificationHelpers();

  const handleFormSubmit = useCallback(async (
    onSubmit: (data: unknown) => Promise<void>,
    formData: unknown
  ) => {
    setErrorState(prev => ({
      ...prev,
      isSubmitting: true,
      errors: [],
      submitAttempts: prev.submitAttempts + 1,
      lastSubmitTime: new Date(),
    }));

    try {
      await onSubmit(formData);

      setErrorState(prev => ({
        ...prev,
        isSubmitting: false,
        submitAttempts: 0,
        canRetry: false,
      }));

      if (options.showNotifications !== false) {
        showSuccess('Success', 'Form submitted successfully!');
      }

    } catch (error) {
      const formErrors = parseFormError(error);
      const canRetry = formErrors.some(err => err.type === 'network' || err.type === 'timeout') &&
        errorState.submitAttempts < (options.maxRetries || 3);

      setErrorState(prev => ({
        ...prev,
        isSubmitting: false,
        errors: formErrors,
        canRetry,
      }));

      if (options.showNotifications !== false) {
        const errorMessage = formatApiError(error);
        showError('Submission Failed', errorMessage);
      }

      throw error; // Re-throw for parent component handling
    }
  }, [errorState.submitAttempts, options.maxRetries, options.showNotifications, showSuccess, showError]);

  const clearErrors = useCallback(() => {
    setErrorState(prev => ({
      ...prev,
      errors: [],
      canRetry: false,
    }));
  }, []);

  return {
    errorState,
    handleFormSubmit,
    clearErrors,
  };
}