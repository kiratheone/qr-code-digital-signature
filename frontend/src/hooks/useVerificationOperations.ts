/**
 * Verification Operations Hook
 * Handles document verification flow with React Query
 */

import { useMutation, useQuery } from '@tanstack/react-query';
import { VerificationService } from '@/lib/services';
import { apiClient } from '@/lib/api';
import type { VerificationInfo, VerificationResult } from '@/lib/types';

// Create service instance
const verificationService = new VerificationService(apiClient);

// Query keys for React Query
export const verificationKeys = {
  all: ['verification'] as const,
  info: (documentId: string) => [...verificationKeys.all, 'info', documentId] as const,
  result: (documentId: string) => [...verificationKeys.all, 'result', documentId] as const,
};

/**
 * Hook for document verification operations
 */
export function useVerificationOperations(documentId: string) {
  // Query for getting verification info
  const verificationInfoQuery = useQuery({
    queryKey: verificationKeys.info(documentId),
    queryFn: () => verificationService.getVerificationInfo(documentId),
    enabled: !!documentId && verificationService.isValidDocumentId(documentId),
    retry: (failureCount, error: any) => {
      // Don't retry on 404 errors (document not found)
      if (error?.status === 404) {
        return false;
      }
      return failureCount < 2;
    },
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  // Mutation for verifying document
  const verifyDocumentMutation = useMutation({
    mutationFn: (file: File) => verificationService.verifyDocument(documentId, file),
    onSuccess: (result: VerificationResult) => {
      console.log('Document verification completed:', result.status);
    },
    onError: (error) => {
      console.error('Document verification failed:', error);
    },
  });

  return {
    // Data
    verificationInfo: verificationInfoQuery.data,
    verificationResult: verifyDocumentMutation.data,
    
    // Loading states
    isLoadingInfo: verificationInfoQuery.isLoading,
    isVerifying: verifyDocumentMutation.isPending,
    
    // Error states
    infoError: verificationInfoQuery.error,
    verificationError: verifyDocumentMutation.error,
    
    // Actions
    verifyDocument: verifyDocumentMutation.mutate,
    
    // Success states
    verificationSuccess: verifyDocumentMutation.isSuccess,
    
    // Utilities
    refetchInfo: verificationInfoQuery.refetch,
    resetVerification: verifyDocumentMutation.reset,
    
    // Document ID validation
    isValidDocumentId: verificationService.isValidDocumentId(documentId),
  };
}

/**
 * Hook for QR code parsing and validation
 */
export function useQRCodeOperations() {
  const parseQRCode = (qrData: string) => {
    return verificationService.parseQRCodeData(qrData);
  };

  const createVerificationUrl = (documentId: string, baseUrl?: string) => {
    return verificationService.createVerificationUrl(documentId, baseUrl);
  };

  return {
    parseQRCode,
    createVerificationUrl,
  };
}

/**
 * Hook for verification result display utilities
 */
export function useVerificationDisplay() {
  const getStatusInfo = (result: VerificationResult) => {
    return verificationService.getVerificationStatusInfo(result);
  };

  const generateSummary = (result: VerificationResult) => {
    return verificationService.generateVerificationSummary(result);
  };

  const formatVerificationDate = (dateString: string) => {
    return verificationService.formatVerificationDate(dateString);
  };

  const validateFileForVerification = (file: File) => {
    return verificationService.validateFileForVerification(file);
  };

  return {
    getStatusInfo,
    generateSummary,
    formatVerificationDate,
    validateFileForVerification,
  };
}

/**
 * Hook for complete verification flow (info + verification)
 */
export function useVerificationFlow(documentId: string) {
  const {
    verificationInfo,
    verificationResult,
    isLoadingInfo,
    isVerifying,
    infoError,
    verificationError,
    verifyDocument,
    verificationSuccess,
    refetchInfo,
    resetVerification,
    isValidDocumentId,
  } = useVerificationOperations(documentId);

  const { getStatusInfo, generateSummary, validateFileForVerification } = useVerificationDisplay();

  // Combined loading state
  const isLoading = isLoadingInfo || isVerifying;
  
  // Combined error state
  const error = infoError || verificationError;
  
  // Verification status info (only available after verification)
  const statusInfo = verificationResult ? getStatusInfo(verificationResult) : null;
  
  // Verification summary (only available after verification)
  const summary = verificationResult ? generateSummary(verificationResult) : null;

  // Enhanced verify function with validation
  const verifyDocumentWithValidation = (file: File) => {
    const validation = validateFileForVerification(file);
    if (!validation.isValid) {
      throw new Error(validation.error);
    }
    verifyDocument(file);
  };

  return {
    // Document info
    documentInfo: verificationInfo,
    
    // Verification result and display
    verificationResult,
    statusInfo,
    summary,
    
    // States
    isLoading,
    isLoadingInfo,
    isVerifying,
    error,
    verificationSuccess,
    
    // Actions
    verifyDocument: verifyDocumentWithValidation,
    refetchInfo,
    resetVerification,
    
    // Validation
    isValidDocumentId,
    validateFile: validateFileForVerification,
    
    // Flow state helpers
    hasDocumentInfo: !!verificationInfo && !infoError,
    hasVerificationResult: !!verificationResult && verificationSuccess,
    canVerify: !!verificationInfo && !infoError && isValidDocumentId,
  };
}