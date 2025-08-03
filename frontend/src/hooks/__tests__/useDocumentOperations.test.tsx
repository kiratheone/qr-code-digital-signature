/**
 * Tests for useDocumentOperations hook
 * Focuses on critical hook behaviors and React Query integration
 */

import { renderHook } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { DocumentList, Document } from '@/lib/types';

// Mock the service with factory function
jest.mock('@/lib/services', () => ({
  DocumentService: jest.fn().mockImplementation(() => ({
    getDocuments: jest.fn().mockResolvedValue({
      documents: [],
      total: 0,
      page: 1,
      per_page: 10,
    }),
    signDocument: jest.fn(),
    deleteDocument: jest.fn(),
    getDocumentById: jest.fn(),
  })),
}));

// Mock API client
jest.mock('@/lib/api', () => ({
  apiClient: {},
}));

import { useDocumentOperations, useDocument } from '../useDocumentOperations';

describe('useDocumentOperations', () => {
  let queryClient: QueryClient;

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  afterEach(() => {
    queryClient.clear();
  });

  it('should initialize with correct default values', () => {
    const { result } = renderHook(() => useDocumentOperations(), { wrapper });

    expect(result.current.documents).toEqual([]);
    expect(result.current.totalDocuments).toBe(0);
    expect(result.current.currentPage).toBe(1);
    expect(result.current.perPage).toBe(10);
    expect(typeof result.current.signDocument).toBe('function');
    expect(typeof result.current.deleteDocument).toBe('function');
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should handle pagination parameters', () => {
    const { result } = renderHook(() => useDocumentOperations(2, 20), { wrapper });

    expect(result.current.currentPage).toBe(2);
    expect(result.current.perPage).toBe(20);
  });
});

describe('useDocument', () => {
  let queryClient: QueryClient;

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  it('should initialize correctly with document ID', () => {
    const { result } = renderHook(() => useDocument('123'), { wrapper });

    expect(result.current.document).toBeUndefined();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should not fetch when documentId is empty', () => {
    const { result } = renderHook(() => useDocument(''), { wrapper });

    expect(result.current.isLoading).toBe(false);
  });
});