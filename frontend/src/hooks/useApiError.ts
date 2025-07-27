import { useState, useCallback } from 'react';
import { 
  formatApiError, 
  isNetworkError, 
  isServerError, 
  isRetryableError,
  getRetryDelay,
  extractErrorDetails,
  EnhancedApiError 
} from '@/utils/apiErrorUtils';
import { ApiErrorResponse } from '@/api/client';

// Type for error objects with status codes
interface ErrorWithStatus {
  status?: number;
  status_code?: number;
}

export interface ErrorState {
  error: EnhancedApiError | ApiErrorResponse | null;
  userMessage: string | null;
  isRetryable: boolean;
  retryCount: number;
  lastRetryAt: Date | null;
}

export function useApiError() {
  const [errorState, setErrorState] = useState<ErrorState>({
    error: null,
    userMessage: null,
    isRetryable: false,
    retryCount: 0,
    lastRetryAt: null,
  });

  const handleError = useCallback((err: unknown, context?: { operation?: string; retryCount?: number }) => {
    const retryCount = context?.retryCount || 0;
    
    if (typeof err === 'object' && err !== null) {
      // Handle enhanced API errors from backend
      if ('type' in err && 'code' in err && 'message' in err) {
        const enhancedError = err as EnhancedApiError;
        setErrorState({
          error: enhancedError,
          userMessage: formatApiError(enhancedError),
          isRetryable: isRetryableError(enhancedError),
          retryCount,
          lastRetryAt: retryCount > 0 ? new Date() : null,
        });
      }
      // Handle legacy API errors
      else if ('status' in err && 'message' in err) {
        const apiError = err as ApiErrorResponse;
        setErrorState({
          error: apiError,
          userMessage: formatApiError(apiError),
          isRetryable: isRetryableError(apiError),
          retryCount,
          lastRetryAt: retryCount > 0 ? new Date() : null,
        });
      }
      // Handle standard JS errors
      else if (err instanceof Error) {
        const apiError: ApiErrorResponse = {
          status: isNetworkError(err) ? 0 : 500,
          message: err.message,
          details: 'An unexpected error occurred',
        };
        setErrorState({
          error: apiError,
          userMessage: formatApiError(apiError),
          isRetryable: isRetryableError(apiError),
          retryCount,
          lastRetryAt: retryCount > 0 ? new Date() : null,
        });
      }
      // Handle unknown object errors
      else {
        const apiError: ApiErrorResponse = {
          status: 500,
          message: 'An unexpected error occurred',
          details: JSON.stringify(err),
        };
        setErrorState({
          error: apiError,
          userMessage: formatApiError(apiError),
          isRetryable: false,
          retryCount,
          lastRetryAt: retryCount > 0 ? new Date() : null,
        });
      }
    } else {
      // Handle primitive errors
      const apiError: ApiErrorResponse = {
        status: 500,
        message: 'An unexpected error occurred',
        details: String(err),
      };
      setErrorState({
        error: apiError,
        userMessage: formatApiError(apiError),
        isRetryable: false,
        retryCount,
        lastRetryAt: retryCount > 0 ? new Date() : null,
      });
    }
  }, []);

  const clearError = useCallback(() => {
    setErrorState({
      error: null,
      userMessage: null,
      isRetryable: false,
      retryCount: 0,
      lastRetryAt: null,
    });
  }, []);

  // Check if the error is a network error
  const isOffline = useCallback(() => {
    return errorState.error !== null && (
      (errorState.error as ErrorWithStatus).status === 0 || 
      (errorState.error as ErrorWithStatus).status_code === 0 ||
      isNetworkError(errorState.error)
    );
  }, [errorState.error]);

  // Check if the error is a server error (5xx)
  const isServerIssue = useCallback(() => {
    return errorState.error !== null && isServerError(errorState.error);
  }, [errorState.error]);

  // Check if the error is a client error (4xx)
  const isClientError = useCallback(() => {
    const error = errorState.error as ErrorWithStatus;
    return errorState.error !== null && (
      (error.status !== undefined && error.status >= 400 && error.status < 500) ||
      (error.status_code !== undefined && error.status_code >= 400 && error.status_code < 500)
    );
  }, [errorState.error]);

  // Get retry delay for the current error
  const getRetryDelayMs = useCallback(() => {
    if (!errorState.error || !errorState.isRetryable) {
      return 0;
    }
    return getRetryDelay(errorState.error, errorState.retryCount + 1);
  }, [errorState.error, errorState.isRetryable, errorState.retryCount]);

  // Get error details for logging
  const getErrorDetails = useCallback(() => {
    if (!errorState.error) {
      return null;
    }
    return extractErrorDetails(errorState.error);
  }, [errorState.error]);

  return {
    // Legacy compatibility
    error: errorState.error,
    userMessage: errorState.userMessage,
    
    // Enhanced error state
    errorState,
    
    // Actions
    handleError,
    clearError,
    
    // Checks
    isOffline,
    isServerIssue,
    isClientError,
    
    // Retry functionality
    isRetryable: errorState.isRetryable,
    retryCount: errorState.retryCount,
    getRetryDelayMs,
    
    // Utilities
    getErrorDetails,
  };
}