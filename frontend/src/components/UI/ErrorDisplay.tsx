'use client';

import React, { useState } from 'react';
import { formatApiError, isRetryableError, isNetworkError } from '@/utils/apiErrorUtils';
import { LoadingButton } from './LoadingSpinner';
import { useNotificationHelpers } from './Notifications';

export interface ErrorDisplayProps {
  error: unknown;
  title?: string;
  onRetry?: () => Promise<void>;
  onDismiss?: () => void;
  showRetry?: boolean;
  showDetails?: boolean;
  className?: string;
  variant?: 'inline' | 'modal' | 'banner';
  size?: 'sm' | 'md' | 'lg';
}

export function ErrorDisplay({
  error,
  title,
  onRetry,
  onDismiss,
  showRetry = true,
  showDetails = false,
  className = '',
  variant = 'inline',
  size = 'md',
}: ErrorDisplayProps) {
  const [isRetrying, setIsRetrying] = useState(false);
  const [showFullDetails, setShowFullDetails] = useState(false);
  const { showSuccess, showError } = useNotificationHelpers();

  const errorMessage = formatApiError(error);
  const isRetryable = showRetry && isRetryableError(error);
  const isOffline = isNetworkError(error);

  const handleRetry = async () => {
    if (!onRetry || isRetrying) return;

    setIsRetrying(true);
    try {
      await onRetry();
      showSuccess('Retry Successful', 'The operation completed successfully.');
    } catch (retryError) {
      showError('Retry Failed', 'The retry attempt was unsuccessful. Please try again.');
    } finally {
      setIsRetrying(false);
    }
  };

  const getVariantClasses = () => {
    const baseClasses = 'rounded-lg border';
    
    switch (variant) {
      case 'modal':
        return `${baseClasses} bg-white shadow-lg p-6`;
      case 'banner':
        return `${baseClasses} bg-red-50 border-red-200 p-4`;
      case 'inline':
      default:
        return `${baseClasses} bg-red-50 border-red-200 p-4`;
    }
  };

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'text-sm';
      case 'lg':
        return 'text-base';
      case 'md':
      default:
        return 'text-sm';
    }
  };

  const getIcon = () => {
    if (isOffline) {
      return (
        <svg className="h-5 w-5 text-orange-400" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
        </svg>
      );
    }

    return (
      <svg className="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
      </svg>
    );
  };

  const getErrorTitle = () => {
    if (title) return title;
    if (isOffline) return 'Connection Problem';
    return 'Error';
  };

  const getHelpText = () => {
    if (isOffline) {
      return 'Please check your internet connection and try again.';
    }
    if (isRetryable) {
      return 'This error might be temporary. You can try again.';
    }
    return 'If this problem persists, please contact support.';
  };

  return (
    <div className={`${getVariantClasses()} ${getSizeClasses()} ${className}`} role="alert">
      <div className="flex">
        <div className="flex-shrink-0">
          {getIcon()}
        </div>
        <div className="ml-3 flex-1">
          <h3 className="font-medium text-red-800">
            {getErrorTitle()}
          </h3>
          <div className="mt-2 text-red-700">
            <p>{errorMessage}</p>
            {!showFullDetails && (
              <p className="mt-1 text-sm text-red-600">{getHelpText()}</p>
            )}
          </div>

          {/* Action buttons */}
          <div className="mt-4 flex flex-wrap gap-2">
            {isRetryable && (
              <LoadingButton
                onClick={handleRetry}
                isLoading={isRetrying}
                loadingText="Retrying..."
                variant="secondary"
                size="sm"
                className="bg-red-50 text-red-800 border-red-300 hover:bg-red-100"
              >
                Try Again
              </LoadingButton>
            )}

            {onDismiss && (
              <button
                onClick={onDismiss}
                className="px-3 py-2 text-sm font-medium text-red-800 bg-red-50 border border-red-300 rounded-md hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                Dismiss
              </button>
            )}

            {showDetails && (
              <button
                onClick={() => setShowFullDetails(!showFullDetails)}
                className="px-3 py-2 text-sm font-medium text-red-800 bg-red-50 border border-red-300 rounded-md hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
              >
                {showFullDetails ? 'Hide Details' : 'Show Details'}
              </button>
            )}
          </div>

          {/* Error details */}
          {showDetails && showFullDetails && (
            <details className="mt-4 p-3 bg-red-100 border border-red-200 rounded-md">
              <summary className="cursor-pointer text-sm font-medium text-red-800">
                Technical Details
              </summary>
              <div className="mt-2 text-xs text-red-700 font-mono">
                <pre className="whitespace-pre-wrap overflow-auto max-h-40">
                  {JSON.stringify(error, null, 2)}
                </pre>
              </div>
            </details>
          )}
        </div>

        {/* Close button */}
        {onDismiss && (
          <div className="ml-auto pl-3">
            <div className="-mx-1.5 -my-1.5">
              <button
                onClick={onDismiss}
                className="inline-flex rounded-md p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-600"
              >
                <span className="sr-only">Dismiss</span>
                <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                </svg>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Success display component for consistency
export interface SuccessDisplayProps {
  message: string;
  title?: string;
  onDismiss?: () => void;
  className?: string;
  variant?: 'inline' | 'modal' | 'banner';
}

export function SuccessDisplay({
  message,
  title = 'Success',
  onDismiss,
  className = '',
  variant = 'inline',
}: SuccessDisplayProps) {
  const getVariantClasses = () => {
    const baseClasses = 'rounded-lg border';
    
    switch (variant) {
      case 'modal':
        return `${baseClasses} bg-white shadow-lg p-6`;
      case 'banner':
        return `${baseClasses} bg-green-50 border-green-200 p-4`;
      case 'inline':
      default:
        return `${baseClasses} bg-green-50 border-green-200 p-4`;
    }
  };

  return (
    <div className={`${getVariantClasses()} ${className}`} role="alert">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg className="h-5 w-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium text-green-800">
            {title}
          </h3>
          <div className="mt-2 text-sm text-green-700">
            <p>{message}</p>
          </div>
        </div>

        {onDismiss && (
          <div className="ml-auto pl-3">
            <div className="-mx-1.5 -my-1.5">
              <button
                onClick={onDismiss}
                className="inline-flex rounded-md p-1.5 text-green-500 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-600"
              >
                <span className="sr-only">Dismiss</span>
                <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                </svg>
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Hook for managing error display state
export function useErrorDisplay() {
  const [error, setError] = useState<unknown>(null);
  const [isVisible, setIsVisible] = useState(false);

  const showError = (err: unknown) => {
    setError(err);
    setIsVisible(true);
  };

  const hideError = () => {
    setIsVisible(false);
    setTimeout(() => setError(null), 300); // Allow for animation
  };

  const clearError = () => {
    setError(null);
    setIsVisible(false);
  };

  return {
    error,
    isVisible,
    showError,
    hideError,
    clearError,
  };
}