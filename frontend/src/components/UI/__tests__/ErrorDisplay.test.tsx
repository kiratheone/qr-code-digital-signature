import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ErrorDisplay, SuccessDisplay, useErrorDisplay } from '../ErrorDisplay';
import { NotificationProvider } from '../Notifications';

// Test wrapper component
const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <NotificationProvider>
    {children}
  </NotificationProvider>
);

describe('ErrorDisplay', () => {
  const mockError = new Error('Test error message');
  const mockApiError = {
    status: 500,
    message: 'Server error',
    details: 'Internal server error occurred',
  };

  it('renders error message', () => {
    render(
      <TestWrapper>
        <ErrorDisplay error={mockError} />
      </TestWrapper>
    );

    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });

  it('renders custom title', () => {
    render(
      <TestWrapper>
        <ErrorDisplay error={mockError} title="Custom Error Title" />
      </TestWrapper>
    );

    expect(screen.getByText('Custom Error Title')).toBeInTheDocument();
  });

  it('shows retry button for retryable errors', () => {
    const retryableError = { status: 500, message: 'Server error' };
    
    render(
      <TestWrapper>
        <ErrorDisplay error={retryableError} onRetry={jest.fn()} />
      </TestWrapper>
    );

    expect(screen.getByText('Try Again')).toBeInTheDocument();
  });

  it('hides retry button when showRetry is false', () => {
    const retryableError = { status: 500, message: 'Server error' };
    
    render(
      <TestWrapper>
        <ErrorDisplay error={retryableError} onRetry={jest.fn()} showRetry={false} />
      </TestWrapper>
    );

    expect(screen.queryByText('Try Again')).not.toBeInTheDocument();
  });

  it('calls onRetry when retry button is clicked', async () => {
    const mockRetry = jest.fn().mockResolvedValue(undefined);
    const retryableError = { status: 500, message: 'Server error' };
    
    render(
      <TestWrapper>
        <ErrorDisplay error={retryableError} onRetry={mockRetry} />
      </TestWrapper>
    );

    const retryButton = screen.getByText('Try Again');
    fireEvent.click(retryButton);

    await waitFor(() => {
      expect(mockRetry).toHaveBeenCalled();
    });
  });

  it('shows loading state during retry', async () => {
    const mockRetry = jest.fn(() => new Promise(resolve => setTimeout(resolve, 100)));
    const retryableError = { status: 500, message: 'Server error' };
    
    render(
      <TestWrapper>
        <ErrorDisplay error={retryableError} onRetry={mockRetry} />
      </TestWrapper>
    );

    const retryButton = screen.getByText('Try Again');
    fireEvent.click(retryButton);

    expect(screen.getByText('Retrying...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Try Again')).toBeInTheDocument();
    });
  });

  it('shows dismiss button when onDismiss is provided', () => {
    const mockDismiss = jest.fn();
    
    render(
      <TestWrapper>
        <ErrorDisplay error={mockError} onDismiss={mockDismiss} />
      </TestWrapper>
    );

    const dismissButtons = screen.getAllByText('Dismiss');
    expect(dismissButtons.length).toBeGreaterThan(0);
  });

  it('calls onDismiss when dismiss button is clicked', () => {
    const mockDismiss = jest.fn();
    
    render(
      <TestWrapper>
        <ErrorDisplay error={mockError} onDismiss={mockDismiss} />
      </TestWrapper>
    );

    const dismissButtons = screen.getAllByText('Dismiss');
    fireEvent.click(dismissButtons[0]); // Click the first dismiss button

    expect(mockDismiss).toHaveBeenCalled();
  });

  it('shows error details when showDetails is true', () => {
    render(
      <TestWrapper>
        <ErrorDisplay error={mockApiError} showDetails={true} />
      </TestWrapper>
    );

    expect(screen.getByText('Show Details')).toBeInTheDocument();
  });

  it('toggles error details visibility', () => {
    render(
      <TestWrapper>
        <ErrorDisplay error={mockApiError} showDetails={true} />
      </TestWrapper>
    );

    const showDetailsButton = screen.getByText('Show Details');
    fireEvent.click(showDetailsButton);

    expect(screen.getByText('Technical Details')).toBeInTheDocument();
    expect(screen.getByText('Hide Details')).toBeInTheDocument();

    const hideDetailsButton = screen.getByText('Hide Details');
    fireEvent.click(hideDetailsButton);

    expect(screen.queryByText('Technical Details')).not.toBeInTheDocument();
    expect(screen.getByText('Show Details')).toBeInTheDocument();
  });

  it('shows connection problem for network errors', () => {
    const networkError = { status: 0, message: 'Network error' };
    
    render(
      <TestWrapper>
        <ErrorDisplay error={networkError} />
      </TestWrapper>
    );

    expect(screen.getByText('Connection Problem')).toBeInTheDocument();
    expect(screen.getByText('Please check your internet connection and try again.')).toBeInTheDocument();
  });

  it('applies different variant classes', () => {
    const { rerender } = render(
      <TestWrapper>
        <ErrorDisplay error={mockError} variant="inline" />
      </TestWrapper>
    );

    let errorDisplay = screen.getByRole('alert');
    expect(errorDisplay).toHaveClass('bg-red-50');

    rerender(
      <TestWrapper>
        <ErrorDisplay error={mockError} variant="modal" />
      </TestWrapper>
    );

    errorDisplay = screen.getByRole('alert');
    expect(errorDisplay).toHaveClass('bg-white', 'shadow-lg');
  });
});

describe('SuccessDisplay', () => {
  it('renders success message', () => {
    render(
      <SuccessDisplay message="Operation completed successfully" />
    );

    expect(screen.getByText('Success')).toBeInTheDocument();
    expect(screen.getByText('Operation completed successfully')).toBeInTheDocument();
  });

  it('renders custom title', () => {
    render(
      <SuccessDisplay message="Test message" title="Custom Success" />
    );

    expect(screen.getByText('Custom Success')).toBeInTheDocument();
  });

  it('shows dismiss button when onDismiss is provided', () => {
    const mockDismiss = jest.fn();
    
    render(
      <SuccessDisplay message="Test message" onDismiss={mockDismiss} />
    );

    const dismissButton = screen.getByRole('button', { name: /dismiss/i });
    expect(dismissButton).toBeInTheDocument();
  });

  it('calls onDismiss when dismiss button is clicked', () => {
    const mockDismiss = jest.fn();
    
    render(
      <SuccessDisplay message="Test message" onDismiss={mockDismiss} />
    );

    const dismissButton = screen.getByRole('button', { name: /dismiss/i });
    fireEvent.click(dismissButton);

    expect(mockDismiss).toHaveBeenCalled();
  });
});

// Test component for useErrorDisplay hook
const ErrorDisplayHookTest = () => {
  const { error, isVisible, showError, hideError, clearError } = useErrorDisplay();

  return (
    <div>
      <button onClick={() => showError(new Error('Test error'))}>
        Show Error
      </button>
      <button onClick={hideError}>Hide Error</button>
      <button onClick={clearError}>Clear Error</button>
      
      {isVisible && error && (
        <div data-testid="error-display">
          Error: {(error as Error).message}
        </div>
      )}
      
      <div data-testid="error-state">
        Visible: {isVisible.toString()}
      </div>
    </div>
  );
};

describe('useErrorDisplay hook', () => {
  it('manages error display state', () => {
    render(<ErrorDisplayHookTest />);

    expect(screen.getByTestId('error-state')).toHaveTextContent('Visible: false');
    expect(screen.queryByTestId('error-display')).not.toBeInTheDocument();

    const showButton = screen.getByText('Show Error');
    fireEvent.click(showButton);

    expect(screen.getByTestId('error-state')).toHaveTextContent('Visible: true');
    expect(screen.getByTestId('error-display')).toHaveTextContent('Error: Test error');
  });

  it('hides error display', () => {
    render(<ErrorDisplayHookTest />);

    const showButton = screen.getByText('Show Error');
    fireEvent.click(showButton);

    expect(screen.getByTestId('error-display')).toBeInTheDocument();

    const hideButton = screen.getByText('Hide Error');
    fireEvent.click(hideButton);

    expect(screen.getByTestId('error-state')).toHaveTextContent('Visible: false');
  });

  it('clears error state', () => {
    render(<ErrorDisplayHookTest />);

    const showButton = screen.getByText('Show Error');
    fireEvent.click(showButton);

    expect(screen.getByTestId('error-display')).toBeInTheDocument();

    const clearButton = screen.getByText('Clear Error');
    fireEvent.click(clearButton);

    expect(screen.getByTestId('error-state')).toHaveTextContent('Visible: false');
    expect(screen.queryByTestId('error-display')).not.toBeInTheDocument();
  });
});