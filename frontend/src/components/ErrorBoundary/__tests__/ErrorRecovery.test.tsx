import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ErrorRecovery, useErrorRecovery } from '../ErrorRecovery';
import { NotificationProvider } from '@/components/UI/Notifications';

// Mock the NetworkManager
jest.mock('@/utils/apiUtils', () => ({
  NetworkManager: {
    getInstance: () => ({
      getStatus: jest.fn(() => true),
      onStatusChange: jest.fn(() => () => {}),
    }),
  },
  createErrorRecoveryHandler: () => ({
    retryFailedQueries: jest.fn(),
    clearErrors: jest.fn(),
  }),
}));

// Test wrapper component
const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <NotificationProvider>
        {children}
      </NotificationProvider>
    </QueryClientProvider>
  );
};

describe('ErrorRecovery', () => {
  beforeEach(() => {
    // Mock window methods
    Object.defineProperty(window, 'location', {
      value: { reload: jest.fn(), href: '' },
      writable: true,
    });
    
    Object.defineProperty(window, 'history', {
      value: { back: jest.fn(), length: 2 },
      writable: true,
    });

    // Mock localStorage and sessionStorage
    Object.defineProperty(window, 'localStorage', {
      value: { clear: jest.fn() },
      writable: true,
    });
    
    Object.defineProperty(window, 'sessionStorage', {
      value: { clear: jest.fn() },
      writable: true,
    });
  });

  it('renders error recovery component', () => {
    const testError = new Error('Test error message');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} />
      </TestWrapper>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
    expect(screen.getByText('Try Again')).toBeInTheDocument();
  });

  it('calls onRetry when retry button is clicked', async () => {
    const mockRetry = jest.fn().mockResolvedValue(undefined);
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} onRetry={mockRetry} />
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
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} onRetry={mockRetry} />
      </TestWrapper>
    );

    const retryButton = screen.getByText('Try Again');
    fireEvent.click(retryButton);

    expect(screen.getByText('Retrying...')).toBeInTheDocument();
    expect(retryButton).toBeDisabled();

    await waitFor(() => {
      expect(screen.getByText('Try Again')).toBeInTheDocument();
    });
  });

  it('refreshes page when refresh button is clicked', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} />
      </TestWrapper>
    );

    const refreshButton = screen.getByText('Refresh Page');
    fireEvent.click(refreshButton);

    expect(window.location.reload).toHaveBeenCalled();
  });

  it('goes back when go back button is clicked', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} />
      </TestWrapper>
    );

    const goBackButton = screen.getByText('Go Back');
    fireEvent.click(goBackButton);

    expect(window.history.back).toHaveBeenCalled();
  });

  it('shows network status when enabled', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} showNetworkStatus={true} />
      </TestWrapper>
    );

    expect(screen.getByText('Online')).toBeInTheDocument();
  });

  it('hides network status when disabled', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} showNetworkStatus={false} />
      </TestWrapper>
    );

    expect(screen.queryByText('Online')).not.toBeInTheDocument();
  });

  it('shows advanced options', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} />
      </TestWrapper>
    );

    const advancedOptions = screen.getByText('Advanced Options');
    fireEvent.click(advancedOptions);

    expect(screen.getByText('Clear Cache & Reload')).toBeInTheDocument();
    expect(screen.getByText('Report Issue')).toBeInTheDocument();
  });

  it('clears cache when clear cache button is clicked', () => {
    const testError = new Error('Test error');
    
    render(
      <TestWrapper>
        <ErrorRecovery error={testError} />
      </TestWrapper>
    );

    // Open advanced options
    const advancedOptions = screen.getByText('Advanced Options');
    fireEvent.click(advancedOptions);

    const clearCacheButton = screen.getByText('Clear Cache & Reload');
    fireEvent.click(clearCacheButton);

    expect(window.localStorage.clear).toHaveBeenCalled();
    expect(window.sessionStorage.clear).toHaveBeenCalled();
  });
});

// Test component for useErrorRecovery hook
const ErrorRecoveryHookTest = () => {
  const { retryFailedOperations, clearErrorState } = useErrorRecovery();

  return (
    <div>
      <button onClick={retryFailedOperations}>Retry Operations</button>
      <button onClick={clearErrorState}>Clear Errors</button>
    </div>
  );
};

describe('useErrorRecovery hook', () => {
  it('provides retry and clear functions', () => {
    render(
      <TestWrapper>
        <ErrorRecoveryHookTest />
      </TestWrapper>
    );

    expect(screen.getByText('Retry Operations')).toBeInTheDocument();
    expect(screen.getByText('Clear Errors')).toBeInTheDocument();
  });

  it('calls retry operations', () => {
    render(
      <TestWrapper>
        <ErrorRecoveryHookTest />
      </TestWrapper>
    );

    const retryButton = screen.getByText('Retry Operations');
    fireEvent.click(retryButton);

    // Should not throw error
    expect(retryButton).toBeInTheDocument();
  });

  it('calls clear error state', () => {
    render(
      <TestWrapper>
        <ErrorRecoveryHookTest />
      </TestWrapper>
    );

    const clearButton = screen.getByText('Clear Errors');
    fireEvent.click(clearButton);

    // Should not throw error
    expect(clearButton).toBeInTheDocument();
  });
});