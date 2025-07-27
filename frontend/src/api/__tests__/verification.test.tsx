import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useVerificationInfo, useVerifyDocument } from '../verification';
import * as client from '../client';
import { ReactNode } from 'react';

// Mock the client module
jest.mock('../client', () => ({
  get: jest.fn(),
  post: jest.fn(),
}));

// Mock useQueryClient
const mockSetQueryData = jest.fn();
jest.mock('@tanstack/react-query', () => ({
  ...jest.requireActual('@tanstack/react-query'),
  useQueryClient: () => ({
    setQueryData: mockSetQueryData,
  }),
}));

const mockGet = client.get as jest.Mock;
const mockPost = client.post as jest.Mock;

// Create a wrapper for React Query
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  const TestWrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );

  TestWrapper.displayName = 'TestWrapper';

  return TestWrapper;
};

describe('Verification API Hooks', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useVerificationInfo', () => {
    it('fetches verification info successfully', async () => {
      const mockVerificationInfo = {
        documentId: '123',
        filename: 'test.pdf',
        issuer: 'Test Issuer',
        createdAt: '2023-01-01T00:00:00.000Z',
        status: 'pending',
        message: 'Please upload the document to verify',
      };

      mockGet.mockResolvedValueOnce(mockVerificationInfo);

      const { result } = renderHook(() => useVerificationInfo('123'), {
        wrapper: createWrapper(),
      });

      expect(result.current.isLoading).toBe(true);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockGet).toHaveBeenCalledWith('/api/verify/123');
      expect(result.current.data).toEqual(mockVerificationInfo);
    });

    it('handles error when fetching verification info', async () => {
      const mockError = { status: 404, message: 'Document not found' };

      // Mock the error for both the initial call and the retry
      mockGet.mockRejectedValue(mockError);

      const { result } = renderHook(() => useVerificationInfo('123'), {
        wrapper: createWrapper(),
      });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      }, { timeout: 3000 });

      expect(mockGet).toHaveBeenCalledWith('/api/verify/123');
      expect(result.current.error).toBeDefined();
    });
  });

  describe('useVerifyDocument', () => {
    it('verifies a document successfully', async () => {
      const mockVerificationResult = {
        documentId: '123',
        filename: 'test.pdf',
        issuer: 'Test Issuer',
        createdAt: '2023-01-01T00:00:00.000Z',
        status: 'valid',
        message: 'Document is valid',
      };

      mockPost.mockResolvedValueOnce(mockVerificationResult);

      const { result } = renderHook(() => useVerifyDocument(), {
        wrapper: createWrapper(),
      });

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });

      result.current.mutate({ docId: '123', file: mockFile });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockPost).toHaveBeenCalled();
      expect(result.current.data).toEqual(mockVerificationResult);
    });

    it('handles error when verifying a document', async () => {
      const mockError = { status: 400, message: 'Invalid file format' };

      mockPost.mockRejectedValueOnce(mockError);

      const { result } = renderHook(() => useVerifyDocument(), {
        wrapper: createWrapper(),
      });

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });

      result.current.mutate({ docId: '123', file: mockFile });

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(mockPost).toHaveBeenCalled();
      expect(result.current.error).toBeDefined();
    });
  });
});