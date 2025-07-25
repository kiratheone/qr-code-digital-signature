import { renderHook, act } from '@testing-library/react';
import { useApiError } from '../useApiError';

describe('useApiError', () => {
  it('handles structured API errors', () => {
    const { result } = renderHook(() => useApiError());
    
    const apiError = {
      status: 404,
      message: 'Resource not found',
      details: 'The requested document does not exist',
    };
    
    act(() => {
      result.current.handleError(apiError);
    });
    
    expect(result.current.error).toEqual(apiError);
    expect(result.current.userMessage).toBe('Resource not found');
  });
  
  it('handles JavaScript Error objects', () => {
    const { result } = renderHook(() => useApiError());
    
    const jsError = new Error('Something went wrong');
    
    act(() => {
      result.current.handleError(jsError);
    });
    
    expect(result.current.error).toEqual({
      status: 500,
      message: 'Something went wrong',
      details: 'An unexpected error occurred',
    });
    expect(result.current.userMessage).toBe('Something went wrong');
  });
  
  it('handles network errors', () => {
    const { result } = renderHook(() => useApiError());
    
    const networkError = new Error('Network connection lost');
    
    act(() => {
      result.current.handleError(networkError);
    });
    
    expect(result.current.isOffline()).toBe(true);
    expect(result.current.isServerIssue()).toBe(false);
  });
  
  it('handles server errors', () => {
    const { result } = renderHook(() => useApiError());
    
    const serverError = {
      status: 500,
      message: 'Internal server error',
    };
    
    act(() => {
      result.current.handleError(serverError);
    });
    
    expect(result.current.isOffline()).toBe(false);
    expect(result.current.isServerIssue()).toBe(true);
  });
  
  it('handles unknown error types', () => {
    const { result } = renderHook(() => useApiError());
    
    act(() => {
      result.current.handleError('Just a string error');
    });
    
    expect(result.current.error).toEqual({
      status: 500,
      message: 'An unexpected error occurred',
      details: 'Just a string error',
    });
  });
  
  it('clears errors', () => {
    const { result } = renderHook(() => useApiError());
    
    act(() => {
      result.current.handleError({ status: 400, message: 'Error message' });
    });
    
    expect(result.current.error).not.toBeNull();
    expect(result.current.userMessage).not.toBeNull();
    
    act(() => {
      result.current.clearError();
    });
    
    expect(result.current.error).toBeNull();
    expect(result.current.userMessage).toBeNull();
  });
});