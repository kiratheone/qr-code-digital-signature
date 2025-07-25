import { formatApiError, isNetworkError, isServerError, isClientError } from '../apiErrorUtils';
import { ApiErrorResponse } from '@/api/client';

describe('apiErrorUtils', () => {
  describe('formatApiError', () => {
    it('handles null or undefined errors', () => {
      expect(formatApiError(null)).toBe('An unknown error occurred');
      expect(formatApiError(undefined)).toBe('An unknown error occurred');
    });
    
    it('formats API errors with status codes', () => {
      const error400: ApiErrorResponse = { status: 400, message: 'Bad request' };
      const error401: ApiErrorResponse = { status: 401, message: 'Unauthorized' };
      const error403: ApiErrorResponse = { status: 403, message: 'Forbidden' };
      const error404: ApiErrorResponse = { status: 404, message: 'Not found' };
      const error500: ApiErrorResponse = { status: 500, message: 'Server error' };
      
      // Should use custom message when provided
      expect(formatApiError(error400)).toBe('Bad request');
      expect(formatApiError(error401)).toBe('Unauthorized');
      expect(formatApiError(error403)).toBe('Forbidden');
      expect(formatApiError(error404)).toBe('Not found');
      expect(formatApiError(error500)).toBe('Server error');
      
      // Should use default messages when no custom message
      const error400NoMsg: ApiErrorResponse = { status: 400, message: '' };
      const error404NoMsg: ApiErrorResponse = { status: 404, message: '' };
      
      expect(formatApiError(error400NoMsg)).toBe('Invalid request. Please check your input and try again.');
      expect(formatApiError(error404NoMsg)).toBe('The requested resource was not found.');
    });
    
    it('formats Error objects', () => {
      const error = new Error('Something went wrong');
      expect(formatApiError(error)).toBe('Something went wrong');
    });
    
    it('formats primitive errors', () => {
      expect(formatApiError('Just a string error')).toBe('Just a string error');
      expect(formatApiError(123)).toBe('123');
    });
  });
  
  describe('isNetworkError', () => {
    it('identifies network errors', () => {
      expect(isNetworkError(new Error('Network error'))).toBe(true);
      expect(isNetworkError(new Error('Failed to fetch'))).toBe(false);
      expect(isNetworkError(new Error('Connection refused'))).toBe(true);
      expect(isNetworkError(new Error('You are offline'))).toBe(true);
    });
    
    it('returns false for non-Error objects', () => {
      expect(isNetworkError({ message: 'Network error' })).toBe(false);
      expect(isNetworkError('Network error')).toBe(false);
      expect(isNetworkError(null)).toBe(false);
    });
  });
  
  describe('isServerError', () => {
    it('identifies server errors (5xx)', () => {
      expect(isServerError({ status: 500, message: 'Server error' })).toBe(true);
      expect(isServerError({ status: 502, message: 'Bad gateway' })).toBe(true);
      expect(isServerError({ status: 503, message: 'Service unavailable' })).toBe(true);
      expect(isServerError({ status: 400, message: 'Bad request' })).toBe(false);
      expect(isServerError({ status: 404, message: 'Not found' })).toBe(false);
    });
    
    it('returns false for non-API errors', () => {
      expect(isServerError(new Error('Server error'))).toBe(false);
      expect(isServerError('Server error')).toBe(false);
      expect(isServerError(null)).toBe(false);
    });
  });
  
  describe('isClientError', () => {
    it('identifies client errors (4xx)', () => {
      expect(isClientError({ status: 400, message: 'Bad request' })).toBe(true);
      expect(isClientError({ status: 401, message: 'Unauthorized' })).toBe(true);
      expect(isClientError({ status: 404, message: 'Not found' })).toBe(true);
      expect(isClientError({ status: 500, message: 'Server error' })).toBe(false);
      expect(isClientError({ status: 200, message: 'OK' })).toBe(false);
    });
    
    it('returns false for non-API errors', () => {
      expect(isClientError(new Error('Client error'))).toBe(false);
      expect(isClientError('Client error')).toBe(false);
      expect(isClientError(null)).toBe(false);
    });
  });
});