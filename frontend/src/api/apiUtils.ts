import { QueryClient } from '@tanstack/react-query';
import { ApiErrorResponse } from './client';

// Enhanced API utilities for better error handling and caching

export interface ApiRequestOptions {
  retries?: number;
  retryDelay?: number;
  timeout?: number;
  signal?: AbortSignal;
}

// Create a request with automatic retry and timeout
export async function createRobustRequest<T>(
  requestFn: () => Promise<T>,
  options: ApiRequestOptions = {}
): Promise<T> {
  const {
    retries = 3,
    retryDelay = 1000,
    timeout = 30000,
    signal,
  } = options;

  let lastError: unknown;
  
  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      // Create timeout promise
      const timeoutPromise = new Promise<never>((_, reject) => {
        const timeoutId = setTimeout(() => {
          reject(new Error('Request timeout'));
        }, timeout);
        
        // Clear timeout if signal is aborted
        signal?.addEventListener('abort', () => {
          clearTimeout(timeoutId);
          reject(new Error('Request aborted'));
        });
      });

      // Race between request and timeout
      const result = await Promise.race([
        requestFn(),
        timeoutPromise,
      ]);

      return result;
    } catch (error) {
      lastError = error;
      
      // Don't retry on client errors (4xx) or if it's the last attempt
      if (
        attempt === retries ||
        (error as ApiErrorResponse)?.status >= 400 && (error as ApiErrorResponse)?.status < 500
      ) {
        break;
      }
      
      // Wait before retrying with exponential backoff
      const delay = retryDelay * Math.pow(2, attempt);
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  throw lastError;
}

// Cache invalidation utilities
export class CacheManager {
  constructor(private queryClient: QueryClient) {}

  // Invalidate all document-related queries
  invalidateDocuments() {
    this.queryClient.invalidateQueries({ queryKey: ['documents'] });
  }

  // Invalidate specific document
  invalidateDocument(docId: string) {
    this.queryClient.invalidateQueries({ queryKey: ['document', docId] });
  }

  // Invalidate verification queries
  invalidateVerification(docId: string) {
    this.queryClient.invalidateQueries({ queryKey: ['verification', docId] });
  }

  // Clear all caches
  clearAll() {
    this.queryClient.clear();
  }

  // Prefetch document list
  async prefetchDocuments(page = 1, limit = 10, search?: string) {
    const { getDocuments } = await import('./document');
    
    await this.queryClient.prefetchQuery({
      queryKey: ['documents', page, limit, search],
      queryFn: () => getDocuments(page, limit, search),
      staleTime: 2 * 60 * 1000, // 2 minutes
    });
  }

  // Prefetch document details
  async prefetchDocument(docId: string) {
    const { getDocumentById } = await import('./document');
    
    await this.queryClient.prefetchQuery({
      queryKey: ['document', docId],
      queryFn: () => getDocumentById(docId),
      staleTime: 5 * 60 * 1000, // 5 minutes
    });
  }
}

// Network status utilities
export class NetworkManager {
  private static instance: NetworkManager;
  private isOnline = true;
  private listeners: ((online: boolean) => void)[] = [];

  static getInstance(): NetworkManager {
    if (!NetworkManager.instance) {
      NetworkManager.instance = new NetworkManager();
    }
    return NetworkManager.instance;
  }

  constructor() {
    if (typeof window !== 'undefined') {
      this.isOnline = navigator.onLine;
      
      window.addEventListener('online', () => {
        this.isOnline = true;
        this.notifyListeners(true);
      });
      
      window.addEventListener('offline', () => {
        this.isOnline = false;
        this.notifyListeners(false);
      });
    }
  }

  getStatus(): boolean {
    return this.isOnline;
  }

  onStatusChange(callback: (online: boolean) => void): () => void {
    this.listeners.push(callback);
    
    // Return unsubscribe function
    return () => {
      this.listeners = this.listeners.filter(listener => listener !== callback);
    };
  }

  private notifyListeners(online: boolean) {
    this.listeners.forEach(listener => listener(online));
  }
}

// Request queue for offline support
export class RequestQueue {
  private static instance: RequestQueue;
  private queue: Array<{
    id: string;
    execute: () => Promise<void>;
  }> = [];
  private processing = false;

  static getInstance(): RequestQueue {
    if (!RequestQueue.instance) {
      RequestQueue.instance = new RequestQueue();
    }
    return RequestQueue.instance;
  }

  async enqueue<T>(request: () => Promise<T>): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const id = Math.random().toString(36).substr(2, 9);
      
      const execute = async () => {
        try {
          const result = await request();
          resolve(result);
        } catch (error) {
          reject(error);
        }
      };
      
      this.queue.push({ id, execute });
      this.processQueue();
    });
  }

  private async processQueue() {
    if (this.processing || this.queue.length === 0) {
      return;
    }

    this.processing = true;
    const networkManager = NetworkManager.getInstance();

    while (this.queue.length > 0 && networkManager.getStatus()) {
      const item = this.queue.shift();
      if (!item) continue;

      try {
        await item.execute();
      } catch (error) {
        // Error handling is done within the execute function
        console.warn('Queue item failed:', error);
      }
    }

    this.processing = false;
  }

  clear() {
    // Clear the queue - individual promises will handle their own rejection
    this.queue = [];
  }
}

// Error recovery utilities
export function createErrorRecoveryHandler(queryClient: QueryClient) {
  return {
    // Retry failed queries
    retryFailedQueries: () => {
      queryClient.getQueryCache().getAll().forEach(query => {
        if (query.state.status === 'error') {
          query.fetch();
        }
      });
    },

    // Retry failed mutations
    retryFailedMutations: () => {
      queryClient.getMutationCache().getAll().forEach(mutation => {
        if (mutation.state.status === 'error') {
          // Note: Mutations can't be automatically retried like queries
          // This would need to be handled at the component level
          console.log('Failed mutation found:', mutation.options.mutationKey);
        }
      });
    },

    // Clear error states
    clearErrors: () => {
      queryClient.getQueryCache().getAll().forEach(query => {
        if (query.state.status === 'error') {
          query.reset();
        }
      });
    },
  };
}