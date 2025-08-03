/**
 * React Query configuration and client setup
 * Provides centralized query client with sensible defaults for caching and error handling
 */

import { QueryClient } from '@tanstack/react-query';
import { ApiClientError } from './apiClient';

/**
 * Create and configure React Query client with simple defaults
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // Cache data for 5 minutes by default
        staleTime: 5 * 60 * 1000,
        // Keep data in cache for 10 minutes
        gcTime: 10 * 60 * 1000,
        // Retry failed requests up to 2 times
        retry: (failureCount, error) => {
          // Don't retry on authentication errors
          if (error instanceof ApiClientError && error.status === 401) {
            return false;
          }
          // Don't retry on client errors (4xx)
          if (error instanceof ApiClientError && error.status >= 400 && error.status < 500) {
            return false;
          }
          // Retry up to 2 times for other errors
          return failureCount < 2;
        },
        // Retry delay with exponential backoff
        retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      },
      mutations: {
        // Retry mutations once on network errors
        retry: (failureCount, error) => {
          if (error instanceof ApiClientError && error.code === 'NETWORK_ERROR') {
            return failureCount < 1;
          }
          return false;
        },
      },
    },
  });
}

// Default query client instance
export const queryClient = createQueryClient();