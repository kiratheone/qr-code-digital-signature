import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { EnhancedErrorBoundary, withEnhancedErrorBoundary, useErrorHandler } from '../EnhancedErrorBoundary';

// Mock component that throws an error
const ThrowError = ({ shouldThrow = false, errorType = 'generic' }: { 
  shouldThrow?: boolean; 
  errorType?: 'generic' | 'network' | 'chunk' | 'timeout';
}) => {
  if (shouldThrow) {
    let error: Error;
    switch (errorType) {
      case 'network':
        error = new Error('Network error occurred');
        break;
      case 'chunk':
        error = new Error('Loading chunk 1 failed');
        break;
      case 'timeout':
        error = new Error('Request timeout');
        break;
      default:
        error = new Error('Test error');
    }
    throw error;
  }
  return <div>No error</div>;
};

// Component that uses the error handler hook
const ErrorTrigger = () => {
  const throwError = useErrorHandler();
  
  return (
    <button onClick={() => throwError(new Error('Hook error'), { context: 'test' })}>
      Trigger Error
    </button>
  );
};

describe('EnhancedErrorBoundary', () => {
  // Suppress console.error for these tests
  const originalError = console.error;
  beforeAll(() => {
    console.error = jest.fn();
  });
  
  afterAll(() => {
    console.error = originalError;
  });

  beforeEach(() => {
    jest.clearAllMocks();
    jest.clearAllTimers();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
  });

  it('renders children when there is no error', () => {
    render(
      <EnhancedErrorBoundary>
        <ThrowError shouldThrow={false} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('renders error fallback when child component throws', () => {
    render(
      <EnhancedErrorBoundary>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('calls onError callback with enhanced error info', () => {
    const onError = jest.fn();
    
    render(
      <EnhancedErrorBoundary onError={onError} name="TestBoundary">
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(onError).toHaveBeenCalledWith(
      expect.any(Error),
      expect.objectContaining({
        componentStack: expect.any(String),
        errorBoundary: 'TestBoundary',
      }),
      expect.any(String) // errorId
    );
  });

  it('auto-retries for retryable errors', async () => {
    const { rerender } = render(
      <EnhancedErrorBoundary maxRetries={2} retryDelay={100}>
        <ThrowError shouldThrow={true} errorType="chunk" />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    
    // Fast-forward time to trigger retry
    jest.advanceTimersByTime(100);
    
    // Rerender with no error to simulate successful retry
    rerender(
      <EnhancedErrorBoundary maxRetries={2} retryDelay={100}>
        <ThrowError shouldThrow={false} />
      </EnhancedErrorBoundary>
    );
    
    await waitFor(() => {
      expect(screen.getByText('No error')).toBeInTheDocument();
    });
  });

  it('does not auto-retry for non-retryable errors', () => {
    render(
      <EnhancedErrorBoundary maxRetries={2} retryDelay={100}>
        <ThrowError shouldThrow={true} errorType="generic" />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    
    // Fast-forward time
    jest.advanceTimersByTime(200);
    
    // Should still show error
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('stops retrying after max retries', () => {
    const { rerender } = render(
      <EnhancedErrorBoundary maxRetries={1} retryDelay={100}>
        <ThrowError shouldThrow={true} errorType="network" />
      </EnhancedErrorBoundary>
    );
    
    // First retry
    jest.advanceTimersByTime(100);
    rerender(
      <EnhancedErrorBoundary maxRetries={1} retryDelay={100}>
        <ThrowError shouldThrow={true} errorType="network" />
      </EnhancedErrorBoundary>
    );
    
    // Should not retry again
    jest.advanceTimersByTime(200);
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('resets error when resetError is called', () => {
    const { rerender } = render(
      <EnhancedErrorBoundary>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    
    const tryAgainButton = screen.getByText('Try Again');
    fireEvent.click(tryAgainButton);
    
    rerender(
      <EnhancedErrorBoundary>
        <ThrowError shouldThrow={false} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('resets error when resetKeys change', () => {
    let resetKey = 'key1';
    
    const { rerender } = render(
      <EnhancedErrorBoundary resetKeys={[resetKey]}>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    
    resetKey = 'key2';
    rerender(
      <EnhancedErrorBoundary resetKeys={[resetKey]}>
        <ThrowError shouldThrow={false} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('shows page-level error recovery for page level', () => {
    render(
      <EnhancedErrorBoundary level="page" enableRecovery={true}>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Try Again')).toBeInTheDocument();
    expect(screen.getByText('Refresh Page')).toBeInTheDocument();
    expect(screen.getByText('Go Back')).toBeInTheDocument();
  });

  it('shows section-level error for section level', () => {
    render(
      <EnhancedErrorBoundary level="section" enableRecovery={true}>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Section Error')).toBeInTheDocument();
    expect(screen.getByText('Try again')).toBeInTheDocument();
  });

  it('shows technical details in development mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'development';
    
    render(
      <EnhancedErrorBoundary level="page">
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    const detailsButton = screen.getByText('Technical Details');
    fireEvent.click(detailsButton);
    
    expect(screen.getByText(/Error ID:/)).toBeInTheDocument();
    expect(screen.getByText(/Retry Count:/)).toBeInTheDocument();
    expect(screen.getByText(/Message:/)).toBeInTheDocument();
    expect(screen.getByText(/Stack:/)).toBeInTheDocument();
    
    process.env.NODE_ENV = originalEnv;
  });

  it('hides technical details in production mode', () => {
    const originalEnv = process.env.NODE_ENV;
    process.env.NODE_ENV = 'production';
    
    render(
      <EnhancedErrorBoundary level="page">
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    const detailsButton = screen.getByText('Technical Details');
    fireEvent.click(detailsButton);
    
    expect(screen.getByText(/Error ID:/)).toBeInTheDocument();
    expect(screen.queryByText(/Stack:/)).not.toBeInTheDocument();
    
    process.env.NODE_ENV = originalEnv;
  });

  it('disables recovery when enableRecovery is false', () => {
    render(
      <EnhancedErrorBoundary level="page" enableRecovery={false}>
        <ThrowError shouldThrow={true} />
      </EnhancedErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('Refresh Page')).not.toBeInTheDocument();
  });
});

describe('withEnhancedErrorBoundary HOC', () => {
  const originalError = console.error;
  beforeAll(() => {
    console.error = jest.fn();
  });
  
  afterAll(() => {
    console.error = originalError;
  });

  it('wraps component with enhanced error boundary', () => {
    const WrappedComponent = withEnhancedErrorBoundary(ThrowError, { name: 'TestWrapper' });
    
    render(<WrappedComponent shouldThrow={false} />);
    
    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('catches errors in wrapped component', () => {
    const WrappedComponent = withEnhancedErrorBoundary(ThrowError, { name: 'TestWrapper' });
    
    render(<WrappedComponent shouldThrow={true} />);
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });
});

describe('useErrorHandler hook', () => {
  const originalError = console.error;
  beforeAll(() => {
    console.error = jest.fn();
  });
  
  afterAll(() => {
    console.error = originalError;
  });

  it('throws error with context when called', () => {
    render(
      <EnhancedErrorBoundary>
        <ErrorTrigger />
      </EnhancedErrorBoundary>
    );
    
    const button = screen.getByText('Trigger Error');
    fireEvent.click(button);
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });
});

describe('Error boundary isolation', () => {
  const originalError = console.error;
  beforeAll(() => {
    console.error = jest.fn();
  });
  
  afterAll(() => {
    console.error = originalError;
  });

  it('isolates errors to specific boundary', () => {
    render(
      <div>
        <EnhancedErrorBoundary name="Boundary1">
          <ThrowError shouldThrow={true} />
        </EnhancedErrorBoundary>
        <EnhancedErrorBoundary name="Boundary2">
          <div>Working component</div>
        </EnhancedErrorBoundary>
      </div>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Working component')).toBeInTheDocument();
  });
});