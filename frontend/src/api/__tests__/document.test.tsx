import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useDocuments, useDocument, useUploadDocument, useDeleteDocument } from '../document';
import * as client from '../client';
import { ReactNode } from 'react';

// Mock the client module
jest.mock('../client', () => ({
  get: jest.fn(),
  post: jest.fn(),
  del: jest.fn(),
}));

const mockGet = client.get as jest.Mock;
const mockPost = client.post as jest.Mock;
const mockDel = client.del as jest.Mock;

// Create a wrapper for React Query
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  
  return ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('Document API Hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useDocuments', () => {
    it('fetches documents successfully', async () => {
      const mockDocuments = {
        documents: [{ 
          id: '1', 
          filename: 'test.pdf',
          createdAt: '2023-01-01T00:00:00.000Z',
          updatedAt: '2023-01-01T00:00:00.000Z',
        }],
        total: 1,
        page: 1,
        limit: 10,
        totalPages: 1,
      };
      
      mockGet.mockResolvedValueOnce(mockDocuments);
      
      const { result } = renderHook(() => useDocuments(1, 10), {
        wrapper: createWrapper(),
      });
      
      expect(result.current.isLoading).toBe(true);
      
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
      
      expect(mockGet).toHaveBeenCalledWith('/api/documents', { page: 1, limit: 10 });
      expect(result.current.data).toEqual(mockDocuments);
    });
  });

  describe('useDocument', () => {
    it('fetches a single document successfully', async () => {
      const mockDocument = { 
        id: '1', 
        filename: 'test.pdf',
        createdAt: '2023-01-01T00:00:00.000Z',
        updatedAt: '2023-01-01T00:00:00.000Z',
      };
      
      mockGet.mockResolvedValueOnce(mockDocument);
      
      const { result } = renderHook(() => useDocument('1'), {
        wrapper: createWrapper(),
      });
      
      expect(result.current.isLoading).toBe(true);
      
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
      
      expect(mockGet).toHaveBeenCalledWith('/api/documents/1');
      expect(result.current.data).toEqual(mockDocument);
    });
  });

  describe('useUploadDocument', () => {
    it('uploads a document successfully', async () => {
      const mockResponse = { id: '1', filename: 'test.pdf' };
      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      
      mockPost.mockResolvedValueOnce(mockResponse);
      
      const { result } = renderHook(() => useUploadDocument(), {
        wrapper: createWrapper(),
      });
      
      result.current.mutate({
        file: mockFile,
        issuer: 'Test Issuer',
      });
      
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
      
      expect(mockPost).toHaveBeenCalled();
      expect(result.current.data).toEqual(mockResponse);
    });
  });

  describe('useDeleteDocument', () => {
    it('deletes a document successfully', async () => {
      const mockResponse = { success: true };
      
      mockDel.mockResolvedValueOnce(mockResponse);
      
      const { result } = renderHook(() => useDeleteDocument(), {
        wrapper: createWrapper(),
      });
      
      result.current.mutate('1');
      
      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });
      
      expect(mockDel).toHaveBeenCalledWith('/api/documents/1');
      expect(result.current.data).toEqual(mockResponse);
    });
  });
});