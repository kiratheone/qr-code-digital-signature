import { ApiErrorResponse } from '@/api/client';

// Enhanced API error interface to match backend error format
export interface EnhancedApiError {
  type?: string;
  code?: string;
  message: string;
  details?: string;
  timestamp?: string;
  request_id?: string;
}

// Generic API error interface for type safety
interface ApiError {
  type?: string;
  code?: string;
  message?: string;
  status_code?: number;
  status?: number;
  details?: string;
  validation_errors?: Array<{ field: string; message: string }>;
  retry_after?: number;
  path?: string;
  method?: string;
  status_code: number;
}

/**
 * Formats an API error into a user-friendly message
 * @param error The API error response
 * @returns A user-friendly error message
 */
export function formatApiError(error: unknown): string {
  if (!error) {
    return 'An unknown error occurred';
  }
  
  // Handle enhanced API errors from backend
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    
    // Check for new backend error format
    if (apiError.type && apiError.code && apiError.message) {
      return getErrorMessage(apiError.type, apiError.code, apiError.message, apiError.status_code);
    }
    
    // Handle legacy format
    if ('status' in error) {
      const statusCode = (error as ApiErrorResponse).status;
      const message = (error as ApiErrorResponse).message;
      
      // If there's a custom message, use it
      if (message && typeof message === 'string') {
        return message;
      }
      
      // Handle common HTTP status codes with default messages
      return getDefaultErrorMessage(statusCode);
    }
    
    if ('message' in error && typeof error.message === 'string') {
      return error.message;
    }
  }
  
  // Handle Error objects
  if (error instanceof Error) {
    return error.message;
  }
  
  // Handle primitive errors
  return String(error);
}

/**
 * Gets user-friendly error message based on error type and code
 */
function getErrorMessage(type: string, code: string, message: string, statusCode: number): string {
  // Use backend message for user-friendly errors
  if (type === 'VALIDATION_ERROR') {
    return message || 'Please check your input and try again.';
  }
  
  if (type === 'AUTHENTICATION_ERROR') {
    return 'Please log in to continue.';
  }
  
  if (type === 'AUTHORIZATION_ERROR') {
    return 'You do not have permission to perform this action.';
  }
  
  if (type === 'NOT_FOUND') {
    return message || 'The requested resource was not found.';
  }
  
  if (type === 'CONFLICT') {
    return message || 'This operation could not be completed due to a conflict.';
  }
  
  if (type === 'RATE_LIMIT_EXCEEDED') {
    return 'Too many requests. Please try again later.';
  }
  
  if (type === 'INTERNAL_ERROR') {
    return 'A server error occurred. Please try again later.';
  }
  
  // Fallback to status code handling
  return getDefaultErrorMessage(statusCode);
}

/**
 * Gets default error message based on HTTP status code
 */
function getDefaultErrorMessage(statusCode: number): string {
  switch (statusCode) {
    case 400:
      return 'Invalid request. Please check your input and try again.';
    case 401:
      return 'Authentication required. Please log in again.';
    case 403:
      return 'You do not have permission to perform this action.';
    case 404:
      return 'The requested resource was not found.';
    case 409:
      return 'This operation could not be completed due to a conflict.';
    case 429:
      return 'Too many requests. Please try again later.';
    case 500:
      return 'Server error. Please try again later.';
    case 502:
      return 'Service temporarily unavailable. Please try again later.';
    case 503:
      return 'Service temporarily unavailable. Please try again later.';
    default:
      return `Error ${statusCode}: An unexpected error occurred`;
  }
}

/**
 * Determines if an error is a network error
 * @param error The error to check
 * @returns True if the error is a network error
 */
export function isNetworkError(error: unknown): boolean {
  if (error instanceof Error) {
    return error.message.toLowerCase().includes('network') || 
           error.message.toLowerCase().includes('connection') ||
           error.message.toLowerCase().includes('offline') ||
           error.message.toLowerCase().includes('fetch');
  }
  
  // Check for network-related status codes
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    const status = apiError.status || apiError.status_code;
    return status === 0 || status === undefined;
  }
  
  return false;
}

/**
 * Determines if an error is a server error (5xx)
 * @param error The error to check
 * @returns True if the error is a server error
 */
export function isServerError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    const status = apiError.status || apiError.status_code;
    return status >= 500 && status < 600;
  }
  
  return false;
}

/**
 * Determines if an error is a client error (4xx)
 * @param error The error to check
 * @returns True if the error is a client error
 */
export function isClientError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    const status = apiError.status || apiError.status_code;
    return status >= 400 && status < 500;
  }
  
  return false;
}

/**
 * Determines if an error is retryable
 * @param error The error to check
 * @returns True if the error is retryable
 */
export function isRetryableError(error: unknown): boolean {
  if (isNetworkError(error)) {
    return true;
  }
  
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    const status = apiError.status || apiError.status_code;
    
    // Retry on server errors and rate limiting
    return status === 429 || status >= 500;
  }
  
  return false;
}

/**
 * Gets retry delay based on error type
 * @param error The error
 * @param attempt The current attempt number
 * @returns Delay in milliseconds
 */
export function getRetryDelay(error: unknown, attempt: number): number {
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    
    // Check for Retry-After header value in error details
    if (apiError.details && typeof apiError.details === 'string') {
      const retryMatch = apiError.details.match(/retry after (\d+) seconds/i);
      if (retryMatch) {
        return parseInt(retryMatch[1]) * 1000;
      }
    }
  }
  
  // Exponential backoff: 1s, 2s, 4s, 8s, etc.
  return Math.min(1000 * Math.pow(2, attempt - 1), 30000);
}

/**
 * Extracts error details for logging
 * @param error The error
 * @returns Error details object
 */
export function extractErrorDetails(error: unknown): Record<string, unknown> {
  if (typeof error === 'object' && error !== null) {
    const apiError = error as ApiError;
    return {
      type: apiError.type,
      code: apiError.code,
      message: apiError.message,
      details: apiError.details,
      status: apiError.status || apiError.status_code,
      request_id: apiError.request_id,
      path: apiError.path,
      method: apiError.method,
      timestamp: apiError.timestamp,
    };
  }
  
  if (error instanceof Error) {
    return {
      name: error.name,
      message: error.message,
      stack: error.stack,
    };
  }
  
  return {
    error: String(error),
  };
}