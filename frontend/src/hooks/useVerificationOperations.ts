import { useState, useCallback, useEffect, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useApiError } from './useApiError';
import { useVerifyDocument } from '@/api/verification';
import { CacheManager, NetworkManager } from '@/api/apiUtils';
import { VerificationResponse } from '@/types/verification';

export function useVerificationOperations() {
  const [isLoading, setIsLoading] = useState(false);
  const [isOffline, setIsOffline] = useState(false);
  const { error, handleError, clearError } = useApiError();
  const queryClient = useQueryClient();
  const cacheManager = useMemo(() => new CacheManager(queryClient), [queryClient]);
  
  // Get verification mutation
  const verifyMutation = useVerifyDocument();
  
  // Monitor network status
  useEffect(() => {
    const networkManager = NetworkManager.getInstance();
    setIsOffline(!networkManager.getStatus());
    
    const unsubscribe = networkManager.onStatusChange((online) => {
      setIsOffline(!online);
    });
    
    return unsubscribe;
  }, []);
  
  // Get verification info with error handling
  const getVerificationInfo = useCallback((docId: string) => {
    // Note: This should be called at the component level, not inside a callback
    // Return a function that can be used to invalidate and refetch
    return () => {
      cacheManager.invalidateVerification(docId);
    };
  }, [cacheManager]);
  
  // Verify document with error handling
  const verifyDocument = useCallback(async (docId: string, file: File): Promise<VerificationResponse> => {
    clearError();
    setIsLoading(true);
    
    try {
      const result = await verifyMutation.mutateAsync({ docId, file });
      setIsLoading(false);
      return result;
    } catch (err) {
      handleError(err);
      setIsLoading(false);
      throw err;
    }
  }, [verifyMutation, handleError, clearError]);
  
  // Refresh verification data
  const refreshVerification = useCallback((docId: string) => {
    cacheManager.invalidateVerification(docId);
  }, [cacheManager]);

  return {
    getVerificationInfo,
    verifyDocument,
    refreshVerification,
    isLoading,
    isOffline,
    error,
    clearError,
  };
}