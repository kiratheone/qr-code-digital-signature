/**
 * Error Message Component
 * Displays user-friendly error messages with different severity levels
 */

import React from 'react';
import { ApiClientError } from '@/lib/api';

interface ErrorMessageProps {
  error: Error | ApiClientError | string | null;
  className?: string;
  showDetails?: boolean;
  onRetry?: () => void;
  onDismiss?: () => void;
}

export function ErrorMessage({ 
  error, 
  className = '', 
  showDetails = false, 
  onRetry, 
  onDismiss 
}: ErrorMessageProps) {
  if (!error) return null;

  const errorInfo = getErrorInfo(error);

  return (
    <div role="alert" className={`rounded-md p-4 ${getErrorStyles(errorInfo.severity)} ${className}`}>
      <div className="flex">
        <div className="flex-shrink-0">
          {getErrorIcon(errorInfo.severity)}
        </div>
        <div className="ml-3 flex-1">
          <h3 className={`text-sm font-medium ${getTextColor(errorInfo.severity)}`}>
            {errorInfo.title}
          </h3>
          <div className={`mt-2 text-sm ${getTextColor(errorInfo.severity, true)}`}>
            <p>{errorInfo.message}</p>
            {showDetails && errorInfo.details && (
              <details className="mt-2">
                <summary className="cursor-pointer font-medium">
                  Technical Details
                </summary>
                <pre className="mt-1 text-xs bg-gray-100 p-2 rounded overflow-auto">
                  {errorInfo.details}
                </pre>
              </details>
            )}
          </div>
          {(onRetry || onDismiss) && (
            <div className="mt-4 flex space-x-2">
              {onRetry && (
                <button
                  type="button"
                  onClick={onRetry}
                  className={`text-sm font-medium ${getButtonColor(errorInfo.severity)} hover:underline`}
                >
                  Try Again
                </button>
              )}
              {onDismiss && (
                <button
                  type="button"
                  onClick={onDismiss}
                  className="text-sm font-medium text-gray-600 hover:text-gray-500 hover:underline"
                >
                  Dismiss
                </button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

interface ErrorInfo {
  title: string;
  message: string;
  details?: string;
  severity: 'error' | 'warning' | 'info';
}

function getErrorInfo(error: Error | ApiClientError | string): ErrorInfo {
  if (typeof error === 'string') {
    return {
      title: 'Error',
      message: error,
      severity: 'error',
    };
  }

  if (error instanceof ApiClientError) {
    return {
      title: getApiErrorTitle(error.code),
      message: getApiErrorMessage(error),
      details: error.details,
      severity: getApiErrorSeverity(error.code),
    };
  }

  // Generic Error
  return {
    title: 'Unexpected Error',
    message: error.message || 'An unexpected error occurred',
    details: error.stack,
    severity: 'error',
  };
}

function getApiErrorTitle(code: string): string {
  switch (code) {
    case 'NETWORK_ERROR':
      return 'Connection Problem';
    case 'UNAUTHORIZED':
      return 'Authentication Required';
    case 'FORBIDDEN':
      return 'Access Denied';
    case 'NOT_FOUND':
      return 'Not Found';
    case 'VALIDATION_FAILED':
      return 'Invalid Input';
    case 'RATE_LIMIT_EXCEEDED':
      return 'Too Many Requests';
    case 'FILE_TOO_LARGE':
      return 'File Too Large';
    case 'INVALID_FILE':
      return 'Invalid File';
    case 'SIGNATURE_FAILED':
      return 'Signature Error';
    case 'VERIFICATION_FAILED':
      return 'Verification Error';
    default:
      return 'Error';
  }
}

function getApiErrorMessage(error: ApiClientError): string {
  switch (error.code) {
    case 'NETWORK_ERROR':
      return 'Unable to connect to the server. Please check your internet connection and try again.';
    case 'UNAUTHORIZED':
      return 'Please log in to continue.';
    case 'FORBIDDEN':
      return 'You don\'t have permission to perform this action.';
    case 'NOT_FOUND':
      return 'The requested resource could not be found.';
    case 'VALIDATION_FAILED':
      return error.message || 'Please check your input and try again.';
    case 'RATE_LIMIT_EXCEEDED':
      return 'You\'re making requests too quickly. Please wait a moment and try again.';
    case 'FILE_TOO_LARGE':
      return 'The file you\'re trying to upload is too large. Please choose a smaller file.';
    case 'INVALID_FILE':
      return 'The file format is not supported. Please upload a PDF file.';
    case 'SIGNATURE_FAILED':
      return 'Failed to sign the document. Please try again.';
    case 'VERIFICATION_FAILED':
      return 'Document verification failed. The document may have been modified.';
    default:
      return error.message || 'An error occurred while processing your request.';
  }
}

function getApiErrorSeverity(code: string): 'error' | 'warning' | 'info' {
  switch (code) {
    case 'NETWORK_ERROR':
    case 'SIGNATURE_FAILED':
    case 'VERIFICATION_FAILED':
      return 'error';
    case 'UNAUTHORIZED':
    case 'FORBIDDEN':
    case 'NOT_FOUND':
      return 'warning';
    case 'VALIDATION_FAILED':
    case 'RATE_LIMIT_EXCEEDED':
    case 'FILE_TOO_LARGE':
    case 'INVALID_FILE':
      return 'info';
    default:
      return 'error';
  }
}

function getErrorStyles(severity: 'error' | 'warning' | 'info'): string {
  switch (severity) {
    case 'error':
      return 'bg-red-50 border border-red-200';
    case 'warning':
      return 'bg-yellow-50 border border-yellow-200';
    case 'info':
      return 'bg-blue-50 border border-blue-200';
  }
}

function getTextColor(severity: 'error' | 'warning' | 'info', secondary = false): string {
  if (secondary) {
    switch (severity) {
      case 'error':
        return 'text-red-700';
      case 'warning':
        return 'text-yellow-700';
      case 'info':
        return 'text-blue-700';
    }
  }

  switch (severity) {
    case 'error':
      return 'text-red-800';
    case 'warning':
      return 'text-yellow-800';
    case 'info':
      return 'text-blue-800';
  }
}

function getButtonColor(severity: 'error' | 'warning' | 'info'): string {
  switch (severity) {
    case 'error':
      return 'text-red-800';
    case 'warning':
      return 'text-yellow-800';
    case 'info':
      return 'text-blue-800';
  }
}

function getErrorIcon(severity: 'error' | 'warning' | 'info') {
  const baseClasses = "h-5 w-5";
  
  switch (severity) {
    case 'error':
      return (
        <svg className={`${baseClasses} text-red-400`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      );
    case 'warning':
      return (
        <svg className={`${baseClasses} text-yellow-400`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z" />
        </svg>
      );
    case 'info':
      return (
        <svg className={`${baseClasses} text-blue-400`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      );
  }
}