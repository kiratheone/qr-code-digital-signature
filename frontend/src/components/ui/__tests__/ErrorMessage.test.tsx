/**
 * ErrorMessage Component Tests
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ErrorMessage } from '../ErrorMessage';
import { ApiClientError } from '@/lib/api';

describe('ErrorMessage', () => {
  it('renders nothing when error is null', () => {
    const { container } = render(<ErrorMessage error={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders string error message', () => {
    render(<ErrorMessage error="Something went wrong" />);
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
  });

  it('renders generic Error object', () => {
    const error = new Error('Test error message');
    render(<ErrorMessage error={error} />);
    expect(screen.getByText('Test error message')).toBeInTheDocument();
    expect(screen.getByText('Unexpected Error')).toBeInTheDocument();
  });

  it('renders ApiClientError with proper styling', () => {
    const error = new ApiClientError(401, 'UNAUTHORIZED', 'Please log in to continue');
    render(<ErrorMessage error={error} />);
    expect(screen.getByText('Authentication Required')).toBeInTheDocument();
    expect(screen.getByText('Please log in to continue.')).toBeInTheDocument();
  });

  it('shows details when showDetails is true', () => {
    const error = new ApiClientError(400, 'VALIDATION_FAILED', 'Invalid input', 'Field is required');
    render(<ErrorMessage error={error} showDetails={true} />);
    
    const detailsButton = screen.getByText('Technical Details');
    expect(detailsButton).toBeInTheDocument();
    
    fireEvent.click(detailsButton);
    expect(screen.getByText('Field is required')).toBeInTheDocument();
  });

  it('calls onRetry when retry button is clicked', () => {
    const onRetry = jest.fn();
    render(<ErrorMessage error="Test error" onRetry={onRetry} />);
    
    const retryButton = screen.getByText('Try Again');
    fireEvent.click(retryButton);
    
    expect(onRetry).toHaveBeenCalledTimes(1);
  });

  it('calls onDismiss when dismiss button is clicked', () => {
    const onDismiss = jest.fn();
    render(<ErrorMessage error="Test error" onDismiss={onDismiss} />);
    
    const dismissButton = screen.getByText('Dismiss');
    fireEvent.click(dismissButton);
    
    expect(onDismiss).toHaveBeenCalledTimes(1);
  });

  it('applies correct severity styling for different error types', () => {
    const { rerender } = render(
      <ErrorMessage error={new ApiClientError(500, 'INTERNAL_ERROR', 'Server error')} />
    );
    expect(screen.getByRole('alert')).toHaveClass('bg-red-50');

    rerender(
      <ErrorMessage error={new ApiClientError(401, 'UNAUTHORIZED', 'Auth required')} />
    );
    expect(screen.getByRole('alert')).toHaveClass('bg-yellow-50');

    rerender(
      <ErrorMessage error={new ApiClientError(400, 'VALIDATION_FAILED', 'Invalid input')} />
    );
    expect(screen.getByRole('alert')).toHaveClass('bg-blue-50');
  });

  it('handles network errors appropriately', () => {
    const error = new ApiClientError(0, 'NETWORK_ERROR', 'Network connection failed');
    render(<ErrorMessage error={error} />);
    
    expect(screen.getByText('Connection Problem')).toBeInTheDocument();
    expect(screen.getByText(/Unable to connect to the server/)).toBeInTheDocument();
  });

  it('handles file-related errors', () => {
    const error = new ApiClientError(413, 'FILE_TOO_LARGE', 'File too large');
    render(<ErrorMessage error={error} />);
    
    expect(screen.getByText('File Too Large')).toBeInTheDocument();
    expect(screen.getByText(/file you're trying to upload is too large/)).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<ErrorMessage error="Test error" className="custom-class" />);
    expect(screen.getByRole('alert')).toHaveClass('custom-class');
  });

  it('shows both retry and dismiss buttons when both handlers are provided', () => {
    const onRetry = jest.fn();
    const onDismiss = jest.fn();
    
    render(<ErrorMessage error="Test error" onRetry={onRetry} onDismiss={onDismiss} />);
    
    expect(screen.getByText('Try Again')).toBeInTheDocument();
    expect(screen.getByText('Dismiss')).toBeInTheDocument();
  });
});

// Add role="alert" to the ErrorMessage component for accessibility
declare global {
  namespace jest {
    interface Matchers<R> {
      toBeInTheDocument(): R;
      toHaveClass(className: string): R;
    }
  }
}