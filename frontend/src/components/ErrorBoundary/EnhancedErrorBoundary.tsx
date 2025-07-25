'use client';

import React, { Component, ReactNode } from 'react';
import { ErrorFallback, ErrorFallbackProps } from './ErrorFallback';
import { ErrorRecovery } from './ErrorRecovery';
import { useNotificationHelpers } from '@/components/UI/Notifications';

interface ErrorInfo {
  componentStack: string;
  errorBoundary?: string;
  errorBoundaryStack?: string;
}

interface EnhancedErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  errorId: string | null;
  retryCount: number;
  lastErrorTime: Date | null;
}

interface EnhancedErrorBoundaryProps {
  children: ReactNode;
  fallback?: React.ComponentType<ErrorFallbackProps>;
  onError?: (error: Error, errorInfo: ErrorInfo, errorId: string) => void;
  resetOnPropsChange?: boolean;
  resetKeys?: Array<string | number>;
  maxRetries?: number;
  retryDelay?: number;
  enableRecovery?: boolean;
  isolateErrors?: boolean;
  level?: 'page' | 'section' | 'component';
  name?: string;
}

export class EnhancedErrorBoundary extends Component<EnhancedErrorBoundaryProps, EnhancedErrorBoundaryState> {
  private resetTimeoutId: number | null = null;
  private retryTimeoutId: number | null = null;

  constructor(props: EnhancedErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
      retryCount: 0,
      lastErrorTime: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<EnhancedErrorBoundaryState> {
    const errorId = `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    return {
      hasError: true,
      error,
      errorId,
      lastErrorTime: new Date(),
    };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    const enhancedErrorInfo: ErrorInfo = {
      componentStack: errorInfo.componentStack,
      errorBoundary: this.props.name || 'EnhancedErrorBoundary',
      errorBoundaryStack: errorInfo.errorBoundaryStack,
    };

    this.setState({
      error,
      errorInfo: enhancedErrorInfo,
    });

    // Generate error ID for tracking
    const errorId = this.state.errorId || `error_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    // Call the onError callback if provided
    if (this.props.onError) {
      this.props.onError(error, enhancedErrorInfo, errorId);
    }

    // Log error details
    this.logError(error, enhancedErrorInfo, errorId);

    // Auto-retry for certain types of errors
    this.handleAutoRetry(error);
  }

  componentDidUpdate(prevProps: EnhancedErrorBoundaryProps) {
    const { resetKeys, resetOnPropsChange } = this.props;
    const { hasError } = this.state;

    // Reset error state if resetKeys have changed
    if (hasError && resetKeys && prevProps.resetKeys) {
      const hasResetKeyChanged = resetKeys.some(
        (key, index) => key !== prevProps.resetKeys?.[index]
      );

      if (hasResetKeyChanged) {
        this.resetError();
      }
    }

    // Reset error state if any props have changed and resetOnPropsChange is true
    if (hasError && resetOnPropsChange && prevProps !== this.props) {
      this.resetError();
    }
  }

  componentWillUnmount() {
    if (this.resetTimeoutId) {
      clearTimeout(this.resetTimeoutId);
    }
    if (this.retryTimeoutId) {
      clearTimeout(this.retryTimeoutId);
    }
  }

  private logError = (error: Error, errorInfo: ErrorInfo, errorId: string) => {
    const errorDetails = {
      errorId,
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
      timestamp: new Date().toISOString(),
      userAgent: typeof window !== 'undefined' ? window.navigator.userAgent : 'unknown',
      url: typeof window !== 'undefined' ? window.location.href : 'unknown',
      level: this.props.level || 'component',
      boundaryName: this.props.name || 'EnhancedErrorBoundary',
    };

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.group(`🚨 Error Boundary: ${errorDetails.boundaryName}`);
      console.error('Error:', error);
      console.error('Error Info:', errorInfo);
      console.error('Error Details:', errorDetails);
      console.groupEnd();
    }

    // In production, send to error reporting service
    if (process.env.NODE_ENV === 'production') {
      // Example: Send to error reporting service
      // errorReportingService.captureException(error, {
      //   extra: errorDetails,
      //   tags: {
      //     errorBoundary: true,
      //     level: this.props.level || 'component',
      //   },
      // });
    }
  };

  private handleAutoRetry = (error: Error) => {
    const { maxRetries = 3, retryDelay = 1000 } = this.props;
    const { retryCount } = this.state;

    // Check if error is retryable (e.g., network errors, chunk load errors)
    const isRetryableError = this.isRetryableError(error);

    if (isRetryableError && retryCount < maxRetries) {
      const delay = retryDelay * Math.pow(2, retryCount); // Exponential backoff
      
      this.retryTimeoutId = window.setTimeout(() => {
        this.setState(prevState => ({
          retryCount: prevState.retryCount + 1,
        }));
        this.resetError();
      }, delay);
    }
  };

  private isRetryableError = (error: Error): boolean => {
    const retryablePatterns = [
      /loading chunk \d+ failed/i,
      /loading css chunk \d+ failed/i,
      /network error/i,
      /fetch error/i,
      /timeout/i,
    ];

    return retryablePatterns.some(pattern => pattern.test(error.message));
  };

  resetError = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
    });
  };

  resetWithRetryCount = () => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
      errorId: null,
      retryCount: 0,
      lastErrorTime: null,
    });
  };

  render() {
    const { hasError, error, errorInfo, errorId, retryCount } = this.state;
    const { 
      children, 
      fallback: FallbackComponent = ErrorFallback,
      enableRecovery = true,
      level = 'component',
    } = this.props;

    if (hasError && error) {
      const errorProps: ErrorFallbackProps = {
        error,
        resetError: this.resetError,
        hasError,
      };

      // For page-level errors, show full error recovery
      if (level === 'page' && enableRecovery) {
        return (
          <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
            <div className="max-w-lg w-full">
              <ErrorRecovery
                error={error}
                onRetry={this.resetError}
                className="mb-6"
              />
              <details className="mt-4">
                <summary className="cursor-pointer text-sm text-gray-600 hover:text-gray-800">
                  Technical Details
                </summary>
                <div className="mt-2 p-3 bg-gray-100 rounded text-xs font-mono">
                  <div><strong>Error ID:</strong> {errorId}</div>
                  <div><strong>Retry Count:</strong> {retryCount}</div>
                  <div><strong>Message:</strong> {error.message}</div>
                  {process.env.NODE_ENV === 'development' && (
                    <>
                      <div className="mt-2"><strong>Stack:</strong></div>
                      <pre className="whitespace-pre-wrap text-xs">{error.stack}</pre>
                      {errorInfo && (
                        <>
                          <div className="mt-2"><strong>Component Stack:</strong></div>
                          <pre className="whitespace-pre-wrap text-xs">{errorInfo.componentStack}</pre>
                        </>
                      )}
                    </>
                  )}
                </div>
              </details>
            </div>
          </div>
        );
      }

      // For section-level errors, show inline error with recovery
      if (level === 'section' && enableRecovery) {
        return (
          <div className="p-4 border border-red-200 rounded-lg bg-red-50">
            <div className="flex items-start">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3 flex-1">
                <h3 className="text-sm font-medium text-red-800">
                  Section Error
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error.message}</p>
                </div>
                <div className="mt-4">
                  <button
                    type="button"
                    onClick={this.resetError}
                    className="bg-red-50 text-red-800 rounded-md text-sm font-medium hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600 px-3 py-2"
                  >
                    Try again
                  </button>
                </div>
              </div>
            </div>
          </div>
        );
      }

      // Default fallback component
      return <FallbackComponent {...errorProps} />;
    }

    return children;
  }
}

// Higher-order component for easier usage
export function withEnhancedErrorBoundary<P extends object>(
  Component: React.ComponentType<P>,
  errorBoundaryProps?: Omit<EnhancedErrorBoundaryProps, 'children'>
) {
  const WrappedComponent = (props: P) => (
    <EnhancedErrorBoundary {...errorBoundaryProps}>
      <Component {...props} />
    </EnhancedErrorBoundary>
  );

  WrappedComponent.displayName = `withEnhancedErrorBoundary(${Component.displayName || Component.name})`;

  return WrappedComponent;
}

// Hook for triggering error boundary from within components
export function useErrorHandler() {
  return (error: Error, context?: Record<string, any>) => {
    // Add context to error if provided
    if (context) {
      (error as any).context = context;
    }
    throw error;
  };
}

// Hook for error boundary context
export function useErrorBoundaryContext() {
  const [errorCount, setErrorCount] = React.useState(0);
  const [lastError, setLastError] = React.useState<Error | null>(null);

  const reportError = React.useCallback((error: Error) => {
    setErrorCount(prev => prev + 1);
    setLastError(error);
  }, []);

  const clearErrors = React.useCallback(() => {
    setErrorCount(0);
    setLastError(null);
  }, []);

  return {
    errorCount,
    lastError,
    reportError,
    clearErrors,
  };
}