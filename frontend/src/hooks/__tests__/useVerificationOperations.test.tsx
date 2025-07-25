/**
 * @jest-environment jsdom
 */

import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactNode } from 'react';
import { useVerificationOperations } from '../useVerificationOperations';
import * as verificationApi from '@/api/verification';

// Mock the verification API
jest.mock('@/api/verification');
const mockVerificationApi = verificationApi as jest.Mocked<typeof verificationApi>;

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

describe('useVerificationOperations', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('useVerificationInfo', () => {
    it('should fetch verification info successfully', async () => {
      const mockVerificationInfo = {
        documentId: 'doc-123',
        filename: 'test-document.pdf',
        issuer: 'Test Issuer',
        createdAt: '2023-01-01T00:00:00Z',
        fileSize: 1024000,
      };

      mockVerificationApi.getVerificationInfo.mockResolvedValue(mockVerificationInfo);

      const { result } = renderHook(
        () => useVerificationOperations().useVerificationInfo('doc-123'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockVerificationInfo);
      expect(mockVerificationApi.getVerificationInfo).toHaveBeenCalledWith('doc-123');
    });

    it('should handle fetch verification info error', async () => {
      const mockError = new Error('Document not found');
      mockVerificationApi.getVerificationInfo.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useVerificationOperations().useVerificationInfo('doc-123'),
        { wrapper: createWrapper() }
      );

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should not fetch when documentId is not provided', () => {
      const { result } = renderHook(
        () => useVerificationOperations().useVerificationInfo(''),
        { wrapper: createWrapper() }
      );

      expect(result.current.data).toBeUndefined();
      expect(mockVerificationApi.getVerificationInfo).not.toHaveBeenCalled();
    });

    it('should refetch when documentId changes', async () => {
      const mockVerificationInfo1 = {
        documentId: 'doc-123',
        filename: 'test1.pdf',
        issuer: 'Issuer 1',
        createdAt: '2023-01-01T00:00:00Z',
        fileSize: 1024000,
      };

      const mockVerificationInfo2 = {
        documentId: 'doc-456',
        filename: 'test2.pdf',
        issuer: 'Issuer 2',
        createdAt: '2023-01-02T00:00:00Z',
        fileSize: 2048000,
      };

      mockVerificationApi.getVerificationInfo
        .mockResolvedValueOnce(mockVerificationInfo1)
        .mockResolvedValueOnce(mockVerificationInfo2);

      const { result, rerender } = renderHook(
        ({ documentId }) => useVerificationOperations().useVerificationInfo(documentId),
        {
          wrapper: createWrapper(),
          initialProps: { documentId: 'doc-123' },
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockVerificationInfo1);
      expect(mockVerificationApi.getVerificationInfo).toHaveBeenCalledWith('doc-123');

      // Change documentId
      rerender({ documentId: 'doc-456' });

      await waitFor(() => {
        expect(result.current.data).toEqual(mockVerificationInfo2);
      });

      expect(mockVerificationApi.getVerificationInfo).toHaveBeenCalledWith('doc-456');
    });
  });

  describe('useVerifyDocument', () => {
    it('should verify document successfully with valid status', async () => {
      const mockVerificationResult = {
        status: 'valid',
        message: 'Document is valid and authentic',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: {
          hashMatch: true,
          signatureValid: true,
        },
      };

      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test pdf content'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockVerificationResult);
      expect(mockVerificationApi.verifyDocument).toHaveBeenCalledWith('doc-123', mockFile);
    });

    it('should verify document with modified status', async () => {
      const mockVerificationResult = {
        status: 'modified',
        message: 'QR code is valid, but document content has been modified',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: {
          hashMatch: false,
          signatureValid: true,
        },
      };

      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['modified content'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockVerificationResult);
    });

    it('should verify document with invalid status', async () => {
      const mockVerificationResult = {
        status: 'invalid',
        message: 'Document signature is invalid or QR code is forged',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: {
          hashMatch: false,
          signatureValid: false,
        },
      };

      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['invalid content'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data).toEqual(mockVerificationResult);
    });

    it('should handle verification error', async () => {
      const mockError = new Error('Verification service unavailable');
      mockVerificationApi.verifyDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument(),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(result.current.error).toEqual(mockError);
    });

    it('should call onSuccess callback', async () => {
      const mockVerificationResult = {
        status: 'valid',
        message: 'Document is valid',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: { hashMatch: true, signatureValid: true },
      };

      const onSuccessMock = jest.fn();
      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument({ onSuccess: onSuccessMock }),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(onSuccessMock).toHaveBeenCalledWith(mockVerificationResult, verificationData, undefined);
    });

    it('should call onError callback', async () => {
      const mockError = new Error('Verification failed');
      const onErrorMock = jest.fn();

      mockVerificationApi.verifyDocument.mockRejectedValue(mockError);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument({ onError: onErrorMock }),
        { wrapper: createWrapper() }
      );

      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      const verificationData = { documentId: 'doc-123', file: mockFile };

      result.current.mutate(verificationData);

      await waitFor(() => {
        expect(result.current.isError).toBe(true);
      });

      expect(onErrorMock).toHaveBeenCalledWith(mockError, verificationData, undefined);
    });

    it('should handle different file types', async () => {
      const mockVerificationResult = {
        status: 'valid',
        message: 'Document is valid',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: { hashMatch: true, signatureValid: true },
      };

      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations().useVerifyDocument(),
        { wrapper: createWrapper() }
      );

      // Test with different file types
      const pdfFile = new File(['pdf content'], 'test.pdf', { type: 'application/pdf' });
      const docFile = new File(['doc content'], 'test.doc', { type: 'application/msword' });

      // Test PDF file
      result.current.mutate({ documentId: 'doc-123', file: pdfFile });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockVerificationApi.verifyDocument).toHaveBeenCalledWith('doc-123', pdfFile);

      // Reset and test DOC file
      result.current.reset();
      result.current.mutate({ documentId: 'doc-123', file: docFile });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(mockVerificationApi.verifyDocument).toHaveBeenCalledWith('doc-123', docFile);
    });
  });

  describe('integration', () => {
    it('should provide all verification operations', () => {
      const { result } = renderHook(
        () => useVerificationOperations(),
        { wrapper: createWrapper() }
      );

      expect(result.current).toHaveProperty('useVerificationInfo');
      expect(result.current).toHaveProperty('useVerifyDocument');

      expect(typeof result.current.useVerificationInfo).toBe('function');
      expect(typeof result.current.useVerifyDocument).toBe('function');
    });

    it('should work with both operations together', async () => {
      const mockVerificationInfo = {
        documentId: 'doc-123',
        filename: 'test.pdf',
        issuer: 'Test Issuer',
        createdAt: '2023-01-01T00:00:00Z',
        fileSize: 1024000,
      };

      const mockVerificationResult = {
        status: 'valid',
        message: 'Document is valid',
        documentId: 'doc-123',
        verifiedAt: '2023-01-01T12:00:00Z',
        details: { hashMatch: true, signatureValid: true },
      };

      mockVerificationApi.getVerificationInfo.mockResolvedValue(mockVerificationInfo);
      mockVerificationApi.verifyDocument.mockResolvedValue(mockVerificationResult);

      const { result } = renderHook(
        () => useVerificationOperations(),
        { wrapper: createWrapper() }
      );

      // Get verification info
      const infoQuery = result.current.useVerificationInfo('doc-123');
      await waitFor(() => {
        expect(infoQuery.isSuccess).toBe(true);
      });

      expect(infoQuery.data).toEqual(mockVerificationInfo);

      // Verify document
      const verifyMutation = result.current.useVerifyDocument();
      const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
      
      verifyMutation.mutate({ documentId: 'doc-123', file: mockFile });

      await waitFor(() => {
        expect(verifyMutation.isSuccess).toBe(true);
      });

      expect(verifyMutation.data).toEqual(mockVerificationResult);
    });
  });
});