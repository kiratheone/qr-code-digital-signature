'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode, useState } from 'react';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import { useNotificationHelpers } from '@/components/UI/Notifications';
import { formatApiError, isNetworkError, isServerError } from '@/utils/apiErrorUtils';

interface ReactQueryProviderProps {
  children: ReactNode;
}

// Global error handler for React Query
function useGlobalErrorHandler() {
  const { showError, showWarning } = useNotificationHelpers();

  return (error: any) => {
    // Format the error message
    const userMessage = formatApiError(error);
    
    // Determine error type and show appropriate notification
    if (isNetworkError(error)) {
      showWarning(
        'Connection Issue',
        'Please check your internet connection and try again.',
        {
          persistent: true,
          action: {
            label: 'Retry',
            onClick: () => window.location.reload(),
          },
        }
      );
    } else if (isServerError(error)) {
      showError(
        'Server Error',
        userMessage,
        {
          persistent: true,
          action: {
            label: 'Report Issue',
            onClick: () => {
              // TODO: Open support form or email
              console.log('Report issue clicked');
            },
          },
        }
      );
    } else {
      showError('Error', userMessage);
    }

    // Log error for debugging
    if (process.env.NODE_ENV !== 'production') {
      console.error('React Query error:', error);
    }
  };
}

function QueryClientProviderWithErrorHandling({ children }: { children: ReactNode }) {
  const handleGlobalError = useGlobalErrorHandler();

  const [queryClient] = useState(() => new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 5 * 60 * 1000, // 5 minutes
        gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
        retry: (failureCount, error: any) => {
          // Don't retry on 404s or other client errors (4xx)
          if (error?.status >= 400 && error?.status < 500) {
            return false;
          }
          // Don't retry on network errors (handled by axios interceptor)
          if (error?.status === 0) {
            return false;
          }
          // Retry up to 3 times for server errors
          return failureCount < 3;
        },
        retryDelay: attemptIndex => Math.min(1000 * Math.pow(2, attemptIndex), 30000), // Exponential backoff
        refetchOnWindowFocus: false,
        refetchOnMount: true,
        refetchOnReconnect: true,
        refetchInterval: false, // Disable automatic refetching
        networkMode: 'online', // Only run queries when online
        onError: handleGlobalError,
      },
      mutations: {
        retry: (failureCount, error: any) => {
          // Only retry network errors or 5xx errors
          if (!error?.status || error?.status >= 500) {
            return failureCount < 2;
          }
          return false;
        },
        retryDelay: attemptIndex => Math.min(1000 * Math.pow(2, attemptIndex), 10000),
        networkMode: 'online', // Only run mutations when online
        onError: handleGlobalError,
      },
    },
  }));

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {process.env.NODE_ENV === 'development' && <ReactQueryDevtools initialIsOpen={false} />}
    </QueryClientProvider>
  );
}

export default function ReactQueryProvider({ children }: ReactQueryProviderProps) {
  return (
    <QueryClientProviderWithErrorHandling>
      {children}
    </QueryClientProviderWithErrorHandling>
  );
}