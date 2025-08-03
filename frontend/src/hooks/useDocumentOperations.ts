/**
 * Document Operations Hook
 * Wraps React Query for document-related operations with loading states and error handling
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { DocumentService } from '@/lib/services';
import { apiClient } from '@/lib/api';
import type { Document, DocumentList, SignDocumentResponse } from '@/lib/types';

// Create service instance
const documentService = new DocumentService(apiClient);

// Query keys for React Query
export const documentKeys = {
  all: ['documents'] as const,
  lists: () => [...documentKeys.all, 'list'] as const,
  list: (page: number, perPage: number) => [...documentKeys.lists(), { page, perPage }] as const,
  details: () => [...documentKeys.all, 'detail'] as const,
  detail: (id: string) => [...documentKeys.details(), id] as const,
};

/**
 * Hook for document operations including signing, listing, and management
 */
export function useDocumentOperations(page: number = 1, perPage: number = 10) {
  const queryClient = useQueryClient();

  // Query for getting documents list
  const documentsQuery = useQuery({
    queryKey: documentKeys.list(page, perPage),
    queryFn: () => documentService.getDocuments(page, perPage),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Mutation for signing documents
  const signDocumentMutation = useMutation({
    mutationFn: ({ file, issuer }: { file: File; issuer: string }) =>
      documentService.signDocument(file, issuer),
    onSuccess: (data: SignDocumentResponse) => {
      // Invalidate and refetch documents list
      queryClient.invalidateQueries({ queryKey: documentKeys.lists() });
      
      // Add the new document to the cache
      queryClient.setQueryData(
        documentKeys.detail(data.document.id),
        data.document
      );
    },
    onError: (error) => {
      console.error('Document signing failed:', error);
    },
  });

  // Mutation for deleting documents
  const deleteDocumentMutation = useMutation({
    mutationFn: (documentId: string) => documentService.deleteDocument(documentId),
    onSuccess: (_, documentId) => {
      // Remove from cache
      queryClient.removeQueries({ queryKey: documentKeys.detail(documentId) });
      
      // Invalidate documents list
      queryClient.invalidateQueries({ queryKey: documentKeys.lists() });
    },
    onError: (error) => {
      console.error('Document deletion failed:', error);
    },
  });

  return {
    // Data
    documents: documentsQuery.data?.documents || [],
    totalDocuments: documentsQuery.data?.total || 0,
    currentPage: documentsQuery.data?.page || page,
    perPage: documentsQuery.data?.per_page || perPage,
    
    // Loading states
    isLoading: documentsQuery.isLoading,
    isSigningDocument: signDocumentMutation.isPending,
    isDeletingDocument: deleteDocumentMutation.isPending,
    
    // Error states
    error: documentsQuery.error,
    signError: signDocumentMutation.error,
    deleteError: deleteDocumentMutation.error,
    
    // Actions
    signDocument: signDocumentMutation.mutate,
    deleteDocument: deleteDocumentMutation.mutate,
    
    // Utilities
    refetch: documentsQuery.refetch,
    
    // Success states
    signSuccess: signDocumentMutation.isSuccess,
    deleteSuccess: deleteDocumentMutation.isSuccess,
    
    // Reset mutations
    resetSignMutation: signDocumentMutation.reset,
    resetDeleteMutation: deleteDocumentMutation.reset,
  };
}

/**
 * Hook for getting a specific document by ID
 */
export function useDocument(documentId: string) {
  const documentQuery = useQuery({
    queryKey: documentKeys.detail(documentId),
    queryFn: () => documentService.getDocumentById(documentId),
    enabled: !!documentId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  return {
    document: documentQuery.data,
    isLoading: documentQuery.isLoading,
    error: documentQuery.error,
    refetch: documentQuery.refetch,
  };
}

/**
 * Hook for document download functionality
 */
export function useDocumentDownload() {
  const downloadMutation = useMutation({
    mutationFn: (documentId: string) => documentService.downloadDocument(documentId),
    onSuccess: (blob, documentId) => {
      // Create download link
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `signed-document-${documentId}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    },
    onError: (error) => {
      console.error('Document download failed:', error);
    },
  });

  return {
    downloadDocument: downloadMutation.mutate,
    isDownloading: downloadMutation.isPending,
    downloadError: downloadMutation.error,
    downloadSuccess: downloadMutation.isSuccess,
    resetDownload: downloadMutation.reset,
  };
}

/**
 * Hook for file validation utilities
 */
export function useFileValidation() {
  const validateFile = (file: File) => {
    return documentService.validateFile(file);
  };

  const formatFileSize = (bytes: number) => {
    return documentService.formatFileSize(bytes);
  };

  const formatDate = (dateString: string) => {
    return documentService.formatDate(dateString);
  };

  return {
    validateFile,
    formatFileSize,
    formatDate,
  };
}