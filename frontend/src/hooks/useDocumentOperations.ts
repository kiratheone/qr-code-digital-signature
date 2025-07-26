import { useState, useCallback, useEffect, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useApiError } from './useApiError';
import {
  useDocuments,
  useDocument,
  useUploadDocument,
  useDeleteDocument
} from '@/api/document';
import { CacheManager, NetworkManager } from '@/api/apiUtils';
import { DocumentUploadRequest } from '@/types/document';

export function useDocumentOperations() {
  const [isLoading, setIsLoading] = useState(false);
  const [isOffline, setIsOffline] = useState(false);
  const { error, handleError, clearError } = useApiError();
  const queryClient = useQueryClient();
  const cacheManager = useMemo(() => new CacheManager(queryClient), [queryClient]);

  // Get document upload mutation
  const uploadMutation = useUploadDocument();

  // Get document delete mutation
  const deleteMutation = useDeleteDocument();

  // Monitor network status
  useEffect(() => {
    const networkManager = NetworkManager.getInstance();
    setIsOffline(!networkManager.getStatus());

    const unsubscribe = networkManager.onStatusChange((online) => {
      setIsOffline(!online);
      if (online) {
        // Retry failed queries when back online
        queryClient.getQueryCache().getAll().forEach(query => {
          if (query.state.status === 'error') {
            query.fetch();
          }
        });
      }
    });

    return unsubscribe;
  }, [queryClient]);

  // Upload document with error handling
  const uploadDocument = useCallback(async (data: DocumentUploadRequest) => {
    clearError();
    setIsLoading(true);

    try {
      const result = await uploadMutation.mutateAsync(data);
      setIsLoading(false);
      return result;
    } catch (err) {
      handleError(err);
      setIsLoading(false);
      throw err;
    }
  }, [uploadMutation, handleError, clearError]);

  // Delete document with error handling
  const deleteDocument = useCallback(async (id: string) => {
    clearError();
    setIsLoading(true);

    try {
      const result = await deleteMutation.mutateAsync(id);
      setIsLoading(false);
      return result;
    } catch (err) {
      handleError(err);
      setIsLoading(false);
      throw err;
    }
  }, [deleteMutation, handleError, clearError]);

  // Expose the hooks directly for use in components
  const useDocumentsHook = useDocuments;
  const useDocumentHook = useDocument;

  // Prefetch documents for better UX
  const prefetchDocuments = useCallback(async (page = 1, limit = 10, search?: string) => {
    try {
      await cacheManager.prefetchDocuments(page, limit, search);
    } catch (error) {
      // Prefetch errors are not critical
      console.warn('Failed to prefetch documents:', error);
    }
  }, [cacheManager]);

  // Prefetch document details
  const prefetchDocument = useCallback(async (docId: string) => {
    try {
      await cacheManager.prefetchDocument(docId);
    } catch (error) {
      // Prefetch errors are not critical
      console.warn('Failed to prefetch document:', error);
    }
  }, [cacheManager]);

  // Refresh all document data
  const refreshDocuments = useCallback(() => {
    cacheManager.invalidateDocuments();
  }, [cacheManager]);

  return {
    uploadDocument,
    deleteDocument,
    useDocuments: useDocumentsHook,
    useDocument: useDocumentHook,
    prefetchDocuments,
    prefetchDocument,
    refreshDocuments,
    isLoading,
    isOffline,
    error,
    clearError,
  };
}