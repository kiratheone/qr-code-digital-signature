/**
 * Tests for DocumentService
 * Focuses on business logic validation and API interaction
 */

import { DocumentService } from '../DocumentService';
import { ApiClient, ApiClientError } from '@/lib/api';
import type { SignDocumentResponse, DocumentList, Document } from '@/lib/types';

// Mock ApiClient
jest.mock('@/lib/api');
const MockedApiClient = ApiClient as jest.MockedClass<typeof ApiClient>;

describe('DocumentService', () => {
  let documentService: DocumentService;
  let mockApiClient: jest.Mocked<ApiClient>;

  beforeEach(() => {
    mockApiClient = new MockedApiClient() as jest.Mocked<ApiClient>;
    documentService = new DocumentService(mockApiClient);
  });

  describe('signDocument', () => {
    const mockFile = new File(['test content'], 'test.pdf', { type: 'application/pdf' });
    const mockResponse: SignDocumentResponse = {
      document: {
        id: '123',
        user_id: 'user1',
        filename: 'test.pdf',
        issuer: 'John Doe',
        document_hash: 'hash123',
        signature_data: 'sig123',
        qr_code_data: 'qr123',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        file_size: 1024,
        status: 'active',
      },
      download_url: 'http://example.com/download/123',
    };

    it('should sign document successfully', async () => {
      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await documentService.signDocument(mockFile, 'John Doe', 'Test Document', '001/2025');

      expect(mockApiClient.post).toHaveBeenCalledWith(
        '/documents/sign',
        expect.any(FormData)
      );
      expect(result).toEqual(mockResponse);
    });

    it('should validate file is required', async () => {
      await expect(
        documentService.signDocument(null as any, 'John Doe', 'Test Document', '001/2025')
      ).rejects.toThrow('File is required');
    });

    it('should validate issuer is required', async () => {
      await expect(
        documentService.signDocument(mockFile, '', 'Test Document', '001/2025')
      ).rejects.toThrow('Issuer name is required');
    });

    it('should validate title is required', async () => {
      await expect(
        documentService.signDocument(mockFile, 'John Doe', '', '001/2025')
      ).rejects.toThrow('Document title is required');
    });

    it('should validate file type is PDF', async () => {
      const txtFile = new File(['test'], 'test.txt', { type: 'text/plain' });
      
      await expect(
        documentService.signDocument(txtFile, 'John Doe', 'Test Document', '001/2025')
      ).rejects.toThrow('Only PDF files are supported');
    });

    it('should validate file size limit', async () => {
      const largeFile = new File(['x'.repeat(51 * 1024 * 1024)], 'large.pdf', { 
        type: 'application/pdf' 
      });
      
      await expect(
        documentService.signDocument(largeFile, 'John Doe', 'Test Document', '001/2025')
      ).rejects.toThrow('File size must be less than 50MB');
    });

    it('should trim issuer name', async () => {
      mockApiClient.post.mockResolvedValue(mockResponse);

  await documentService.signDocument(mockFile, '  John Doe  ', '  Test Document  ', '001/2025');

  const formData = mockApiClient.post.mock.calls[0][1] as FormData;
  expect(formData.get('issuer')).toBe('John Doe');
  expect(formData.get('letter_number')).toBe('001/2025');
    });
  });

  describe('getDocuments', () => {
    const mockDocumentList: DocumentList = {
      documents: [],
      total: 0,
      page: 1,
      per_page: 10,
    };

    it('should get documents with default pagination', async () => {
      mockApiClient.get.mockResolvedValue(mockDocumentList);

      const result = await documentService.getDocuments();

      expect(mockApiClient.get).toHaveBeenCalledWith('/documents?page=1&per_page=10');
      expect(result).toEqual(mockDocumentList);
    });

    it('should get documents with custom pagination', async () => {
      mockApiClient.get.mockResolvedValue(mockDocumentList);

      await documentService.getDocuments(2, 20);

      expect(mockApiClient.get).toHaveBeenCalledWith('/documents?page=2&per_page=20');
    });

    it('should validate page number', async () => {
      await expect(
        documentService.getDocuments(0)
      ).rejects.toThrow('Page number must be greater than 0');
    });

    it('should validate per page limits', async () => {
      await expect(
        documentService.getDocuments(1, 0)
      ).rejects.toThrow('Per page must be between 1 and 100');

      await expect(
        documentService.getDocuments(1, 101)
      ).rejects.toThrow('Per page must be between 1 and 100');
    });
  });

  describe('getDocumentById', () => {
    const mockDocument: Document = {
      id: '123',
      user_id: 'user1',
      filename: 'test.pdf',
      issuer: 'John Doe',
      document_hash: 'hash123',
      signature_data: 'sig123',
      qr_code_data: 'qr123',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
      file_size: 1024,
      status: 'active',
    };

    it('should get document by ID', async () => {
      mockApiClient.get.mockResolvedValue(mockDocument);

      const result = await documentService.getDocumentById('123');

      expect(mockApiClient.get).toHaveBeenCalledWith('/documents/123');
      expect(result).toEqual(mockDocument);
    });

    it('should validate document ID is required', async () => {
      await expect(
        documentService.getDocumentById('')
      ).rejects.toThrow('Document ID is required');
    });
  });

  describe('deleteDocument', () => {
    it('should delete document by ID', async () => {
      mockApiClient.delete.mockResolvedValue(undefined);

      await documentService.deleteDocument('123');

      expect(mockApiClient.delete).toHaveBeenCalledWith('/documents/123');
    });

    it('should validate document ID is required', async () => {
      await expect(
        documentService.deleteDocument('')
      ).rejects.toThrow('Document ID is required');
    });
  });

  describe('validateFile', () => {
    it('should validate valid PDF file', () => {
      const validFile = new File(['content'], 'test.pdf', { type: 'application/pdf' });
      
      const result = documentService.validateFile(validFile);
      
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should reject missing file', () => {
      const result = documentService.validateFile(null as any);
      
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('File is required');
    });

    it('should reject non-PDF files', () => {
      const txtFile = new File(['content'], 'test.txt', { type: 'text/plain' });
      
      const result = documentService.validateFile(txtFile);
      
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('Only PDF files are supported');
    });

    it('should reject files over size limit', () => {
      const largeFile = new File(['x'.repeat(51 * 1024 * 1024)], 'large.pdf', { 
        type: 'application/pdf' 
      });
      
      const result = documentService.validateFile(largeFile);
      
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('File size must be less than 50MB');
    });

    it('should reject empty files', () => {
      const emptyFile = new File([], 'empty.pdf', { type: 'application/pdf' });
      
      const result = documentService.validateFile(emptyFile);
      
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('File cannot be empty');
    });
  });

  describe('formatFileSize', () => {
    it('should format bytes correctly', () => {
      expect(documentService.formatFileSize(0)).toBe('0 Bytes');
      expect(documentService.formatFileSize(1024)).toBe('1 KB');
      expect(documentService.formatFileSize(1024 * 1024)).toBe('1 MB');
      expect(documentService.formatFileSize(1536)).toBe('1.5 KB');
    });
  });

  describe('formatDate', () => {
    it('should format valid date', () => {
      const result = documentService.formatDate('2024-01-01T12:00:00Z');
      expect(result).toMatch(/Jan 1, 2024/);
    });

    it('should handle invalid date', () => {
      const result = documentService.formatDate('invalid-date');
      expect(result).toBe('Invalid date');
    });
  });
});