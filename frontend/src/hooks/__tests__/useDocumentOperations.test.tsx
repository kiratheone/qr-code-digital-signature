/**
 * @jest-environment jsdom
 */

import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import { useDocumentOperations } from '../useDocumentOperations';
import * as documentApi from '@/api/document';

// Mock the document API
jest.mock('@/api/document');
const mockDocumentApi = documentApi as jest.Mocked<typeof documentApi>;

// Create a wrapper for React Query
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
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

describe('useDocumentOperations', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useDocuments', () => {
    it('should fetch documents successfully', async () => {
      const mockDocuments = {
        documents: [
          {
            id: '1',
            filename: 'test1.pdf',
            issuer: 'Test Issuer 1',
            created_at: '2023-01-01T00:00:00Z',
            document_hash: 'hash1',
            signature_data: 'sig1',
          },
          {
            id: '2',
            filename: 'test2.pdf',
            issuer: 'Test Issuer 2',
            created_at: '2023-01-02T00:00:00Z',
            document_hash: 'hash2',
            signature_data: 'sig2',
          },
        ],
        total: 2,
        page: 1,
        limit: 10,
      };

      mockDocumentApi.getDocuments.mockResolvedValue(mockDocuments);

      const { result } = renderHook(
        () => useDocumentOperations().useDocuments(1, 10, ''),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockDocuments);
      expect(mockDocumentApi.getDocuments).toHaveBeenCalledWith(1, 10, '');
    });

    it('should handle fetch documents error', async () => {
      const mockError = new Error('Failed to fetch documents');
      mockDocumentApi.getDocuments.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useDocuments(1, 10, ''),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should refetch when parameters change', async () => {
      const mockDocuments = {
        documents: [],
        total: 0,
        page: 1,
        limit: 10,
      };

      mockDocumentApi.getDocuments.mockResolvedValue(mockDocuments);

      const { result, rerender } = renderHook(
        ({ page, limit, search }) => useDocumentOperations().useDocuments(page, limit, search),
        {
          wrapper: createWrapper(),
          initialProps: { page: 1, limit: 10, search: '' },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockDocumentApi.getDocuments).toHaveBeenCalledWith(1, 10, '');

      // Change parameters
      rerender({ page: 2, limit: 5, search: 'test' });

      await waitFor(() => {
        expect(mockDocumentApi.getDocuments).toHaveBeenCalledWith(2, 5, 'test');
      });
    });
  });

  describe('useDocument', () => {
    it('should fetch single document successfully', async () => {
      const mockDocument = {
        id: '1',
        filename: 'test.pdf',
        issuer: 'Test Issuer',
        created_at: '2023-01-01T00:00:00Z',
        document_hash: 'hash1',
        signature_data: 'sig1',
      };

      mockDocumentApi.getDocument.mockResolvedValue(mockDocument);

      const { result } = renderHook(
        () => useDocumentOperations().useDocument('1'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockDocument);
      expect(mockDocumentApi.getDocument).toHaveBeenCalledWith('1');
    });

    it('should handle fetch single document error', async () => {
      const mockError = new Error('Document not found');
      mockDocumentApi.getDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useDocument('1'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should not fetch when id is not provided', () => {
      const { result } = renderHook(
        () => useDocumentOperations().useDocument(''),
        { wrapper: createWrapper() }
      );

      expect(result.current.data).toBeUndefined();
      expect(mockDocumentApi.getDocument).not.toHaveBeenCalled();
    });
  });

  describe('useUploadDocument', () => {
    it('should upload document successfully', async () => {
      const mockResponse = {
        id: '1',
        filename: 'test.pdf',
        qr_code: 'base64-qr-code',
        download_url: '/download/1',
      };

      mockDocumentApi.uploadDocument.mockResolvedValue(mockResponse);

      const { result } = renderHook(
        () => useDocumentOperations().useUploadDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const uploadData = { file: mockFile, issuer: 'Test Issuer' };

      result.current.mutate(uploadData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockResponse);
      expect(mockDocumentApi.uploadDocument).toHaveBeenCalledWith(uploadData);
    });

    it('should handle upload document error', async () => {
      const mockError = new Error('Upload failed');
      mockDocumentApi.uploadDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useUploadDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const uploadData = { file: mockFile, issuer: 'Test Issuer' };

      result.current.mutate(uploadData);

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should call onSuccess callback', async () => {
      const mockResponse = { id: '1', filename: 'test.pdf' };
      const onSuccessMock = jest.fn();

      mockDocumentApi.uploadDocument.mockResolvedValue(mockResponse);

      const { result } = renderHook(
        () => useDocumentOperations().useUploadDocument({ onSuccess: onSuccessMock }),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const uploadData = { file: mockFile, issuer: 'Test Issuer' };

      result.current.mutate(uploadData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(onSuccessMock).toHaveBeenCalledWith(mockResponse, uploadData, undefined);
    });

    it('should call onError callback', async () => {
      const mockError = new Error('Upload failed');
      const onErrorMock = jest.fn();

      mockDocumentApi.uploadDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useUploadDocument({ onError: onErrorMock }),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const uploadData = { file: mockFile, issuer: 'Test Issuer' };

      result.current.mutate(uploadData);

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(onErrorMock).toHaveBeenCalledWith(mockError, uploadData, undefined);
    });
  });

  describe('useDeleteDocument', () => {
    it('should delete document successfully', async () => {
      mockDocumentApi.deleteDocument.mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDocumentOperations().useDeleteDocument(),
        { wrapper: createWrapper() }
      );

      result.current.mutate('1');

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockDocumentApi.deleteDocument).toHaveBeenCalledWith('1');
    });

    it('should handle delete document error', async () => {
      const mockError = new Error('Delete failed');
      mockDocumentApi.deleteDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useDeleteDocument(),
        { wrapper: createWrapper() }
      );

      result.current.mutate('1');

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should call onSuccess callback', async () => {
      const onSuccessMock = jest.fn();
      mockDocumentApi.deleteDocument.mockResolvedValue(undefined);

      const { result } = renderHook(
        () => useDocumentOperations().useDeleteDocument({ onSuccess: onSuccessMock }),
        { wrapper: createWrapper() }
      );

      result.current.mutate('1');

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(onSuccessMock).toHaveBeenCalledWith(undefined, '1', undefined);
    });

    it('should call onError callback', async () => {
      const mockError = new Error('Delete failed');
      const onErrorMock = jest.fn();

      mockDocumentApi.deleteDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useDocumentOperations().useDeleteDocument({ onError: onErrorMock }),
        { wrapper: createWrapper() }
      );

      result.current.mutate('1');

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(onErrorMock).toHaveBeenCalledWith(mockError, '1', undefined);
    });
  });

  describe('integration', () => {
    it('should provide all document operations', () => {
      const { result } = renderHook(
        () => useDocumentOperations(),
        { wrapper: createWrapper() }
      );

      expect(result.current).toHaveProperty('useDocuments');
      expect(result.current).toHaveProperty('useDocument');
      expect(result.current).toHaveProperty('useUploadDocument');
      expect(result.current).toHaveProperty('useDeleteDocument');

      expect(typeof result.current.useDocuments).toBe('function');
      expect(typeof result.current.useDocument).toBe('function');
      expect(typeof result.current.useUploadDocument).toBe('function');
      expect(typeof result.current.useDeleteDocument).toBe('function');
    });
  });
});