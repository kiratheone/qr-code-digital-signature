import { QueryClient } from '@tanstack/react-query';
import { 
  createRobustRequest, 
  CacheManager, 
  NetworkManager, 
  RequestQueue,
  createErrorRecoveryHandler 
} from '../apiUtils';

// Mock QueryClient
const mockQueryClient = {
  invalidateQueries: jest.fn(),
  removeQueries: jest.fn(),
  clear: jest.fn(),
  prefetchQuery: jest.fn(),
  getQueryCache: jest.fn(() => ({
    getAll: jest.fn(() => []),
  })),
  getMutationCache: jest.fn(() => ({
    getAll: jest.fn(() => []),
  })),
} as unknown as QueryClient;

describe('apiUtils', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.clearAllTimers();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('createRobustRequest', () => {
    it('succeeds on first attempt', async () => {
      const mockFn = jest.fn().mockResolvedValue('success');
      
      const result = await createRobustRequest(mockFn);
      
      expect(result).toBe('success');
      expect(mockFn).toHaveBeenCalledTimes(1);
    });

    it('does not retry on client errors (4xx)', async () => {
      const mockFn = jest.fn().mockRejectedValue({ status: 400, message: 'Bad request' });
      
      await expect(createRobustRequest(mockFn)).rejects.toEqual({
        status: 400,
        message: 'Bad request',
      });
      
      expect(mockFn).toHaveBeenCalledTimes(1);
    });
  });

  describe('CacheManager', () => {
    let cacheManager: CacheManager;

    beforeEach(() => {
      cacheManager = new CacheManager(mockQueryClient);
    });

    it('invalidates documents', () => {
      cacheManager.invalidateDocuments();
      
      expect(mockQueryClient.invalidateQueries).toHaveBeenCalledWith({
        queryKey: ['documents'],
      });
    });

    it('invalidates specific document', () => {
      cacheManager.invalidateDocument('doc-123');
      
      expect(mockQueryClient.invalidateQueries).toHaveBeenCalledWith({
        queryKey: ['document', 'doc-123'],
      });
    });

    it('invalidates verification', () => {
      cacheManager.invalidateVerification('doc-123');
      
      expect(mockQueryClient.invalidateQueries).toHaveBeenCalledWith({
        queryKey: ['verification', 'doc-123'],
      });
    });

    it('clears all caches', () => {
      cacheManager.clearAll();
      
      expect(mockQueryClient.clear).toHaveBeenCalled();
    });
  });

  describe('NetworkManager', () => {
    let networkManager: NetworkManager;

    beforeEach(() => {
      // Reset singleton
      (NetworkManager as unknown as { instance: NetworkManager | undefined }).instance = undefined;
      
      // Mock navigator.onLine
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true,
      });
      
      networkManager = NetworkManager.getInstance();
    });

    it('returns singleton instance', () => {
      const instance1 = NetworkManager.getInstance();
      const instance2 = NetworkManager.getInstance();
      
      expect(instance1).toBe(instance2);
    });

    it('tracks online status', () => {
      expect(networkManager.getStatus()).toBe(true);
      
      // Simulate going offline
      Object.defineProperty(navigator, 'onLine', { value: false });
      window.dispatchEvent(new Event('offline'));
      
      expect(networkManager.getStatus()).toBe(false);
    });

    it('notifies listeners of status changes', () => {
      const listener = jest.fn();
      const unsubscribe = networkManager.onStatusChange(listener);
      
      // Simulate going offline
      window.dispatchEvent(new Event('offline'));
      
      expect(listener).toHaveBeenCalledWith(false);
      
      // Unsubscribe and test no more notifications
      unsubscribe();
      window.dispatchEvent(new Event('online'));
      
      expect(listener).toHaveBeenCalledTimes(1);
    });
  });

  describe('RequestQueue', () => {
    let requestQueue: RequestQueue;

    beforeEach(() => {
      // Reset singleton
      (RequestQueue as unknown as { instance: RequestQueue | undefined }).instance = undefined;
      requestQueue = RequestQueue.getInstance();
    });

    it('returns singleton instance', () => {
      const instance1 = RequestQueue.getInstance();
      const instance2 = RequestQueue.getInstance();
      
      expect(instance1).toBe(instance2);
    });

    it('processes requests when online', async () => {
      const mockRequest = jest.fn().mockResolvedValue('result');
      
      // Mock network as online
      const networkManager = NetworkManager.getInstance();
      jest.spyOn(networkManager, 'getStatus').mockReturnValue(true);
      
      const result = await requestQueue.enqueue(mockRequest);
      
      expect(result).toBe('result');
      expect(mockRequest).toHaveBeenCalledTimes(1);
    });

    it('clears queue', () => {
      const mockRequest = jest.fn();
      
      // Add request to queue but don't process
      const promise = requestQueue.enqueue(mockRequest);
      
      requestQueue.clear();
      
      expect(promise).rejects.toThrow('Queue cleared');
    });
  });

  describe('createErrorRecoveryHandler', () => {
    it('creates error recovery handler', () => {
      const handler = createErrorRecoveryHandler(mockQueryClient);
      
      expect(handler).toHaveProperty('retryFailedQueries');
      expect(handler).toHaveProperty('retryFailedMutations');
      expect(handler).toHaveProperty('clearErrors');
    });

    it('retries failed queries', () => {
      const mockQuery = {
        state: { status: 'error' },
        fetch: jest.fn(),
      };
      
      const mockQueryCache = {
        getAll: jest.fn().mockReturnValue([mockQuery]),
      };
      
      (mockQueryClient.getQueryCache as jest.Mock).mockReturnValue(mockQueryCache);
      
      const handler = createErrorRecoveryHandler(mockQueryClient);
      handler.retryFailedQueries();
      
      expect(mockQuery.fetch).toHaveBeenCalled();
    });
  });
});