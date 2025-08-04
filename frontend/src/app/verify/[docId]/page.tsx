/**
 * Document Verification Page
 * Page for verifying document authenticity using QR code
 * Uses hooks and components with clean separation
 * Implements lazy loading for better performance
 */

'use client';

import React, { Suspense, lazy } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { LoadingSpinner } from '@/components/ui/LoadingSpinner';
import { useVerificationFlow } from '@/hooks';

// Lazy load verification components for better performance
const VerificationForm = lazy(() => import('@/components/VerificationForm').then(module => ({ default: module.VerificationForm })));
const VerificationResult = lazy(() => import('@/components/VerificationResult').then(module => ({ default: module.VerificationResult })));

export default function VerifyDocumentPage() {
  const params = useParams();
  const router = useRouter();
  const documentId = params.docId as string;

  // Use custom hook for all verification logic
  const {
    documentInfo,
    verificationResult,
    statusInfo,
    summary,
    isLoading,
    isLoadingInfo,
    isVerifying,
    error,
    verificationSuccess,
    verifyDocument,
    refetchInfo,
    resetVerification,
    isValidDocumentId,
    hasDocumentInfo,
    hasVerificationResult,
    canVerify,
  } = useVerificationFlow(documentId);

  // Handle document verification
  const handleDocumentVerify = (file: File) => {
    verifyDocument(file);
  };

  // Handle error dismissal
  const handleErrorDismiss = () => {
    resetVerification();
  };

  // Handle verify another document
  const handleVerifyAnother = () => {
    resetVerification();
  };

  // Handle go to home
  const handleGoHome = () => {
    router.push('/');
  };

  // Handle retry loading document info
  const handleRetryLoadInfo = () => {
    refetchInfo();
  };

  // Show error if document ID is invalid
  if (!isValidDocumentId) {
    return (
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-medium text-red-800">Invalid Document ID</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>The document ID in the URL is not valid. Please check the QR code or link and try again.</p>
              </div>
              <div className="mt-4">
                <button
                  onClick={handleGoHome}
                  className="text-sm font-medium text-red-800 hover:text-red-600"
                >
                  Go to Home
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show loading state while fetching document info
  if (isLoadingInfo) {
    return (
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white shadow rounded-lg p-6">
          <div className="animate-pulse">
            <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
            <div className="space-y-3">
              <div className="h-4 bg-gray-200 rounded w-full"></div>
              <div className="h-4 bg-gray-200 rounded w-3/4"></div>
              <div className="h-4 bg-gray-200 rounded w-1/2"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Show error if document info couldn't be loaded
  if (error && !hasDocumentInfo) {
    return (
      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-red-50 border border-red-200 rounded-md p-6">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-6 w-6 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 15.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-lg font-medium text-red-800">Document Not Found</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>
                  {error instanceof Error ? error.message : 'The document could not be found. It may have been deleted or the link may be incorrect.'}
                </p>
              </div>
              <div className="mt-4 space-x-4">
                <button
                  onClick={handleRetryLoadInfo}
                  className="text-sm font-medium text-red-800 hover:text-red-600"
                >
                  Try Again
                </button>
                <button
                  onClick={handleGoHome}
                  className="text-sm font-medium text-red-800 hover:text-red-600"
                >
                  Go to Home
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8 text-center">
        <h1 className="text-3xl font-bold text-gray-900">Document Verification</h1>
        <p className="mt-2 text-gray-600">
          Verify the authenticity and integrity of a digitally signed document.
        </p>
      </div>

      {/* Show verification result if verification is complete */}
      {hasVerificationResult && verificationResult ? (
        <Suspense fallback={<LoadingSpinner />}>
          <VerificationResult
            result={verificationResult}
            onVerifyAnother={handleVerifyAnother}
            onGoHome={handleGoHome}
          />
        </Suspense>
      ) : (
        /* Show verification form if document info is available */
        hasDocumentInfo && documentInfo && (
          <Suspense fallback={<LoadingSpinner />}>
            <VerificationForm
              documentInfo={documentInfo}
              onVerify={handleDocumentVerify}
              isLoading={isVerifying}
              error={error instanceof Error ? error.message : null}
              onErrorDismiss={handleErrorDismiss}
            />
          </Suspense>
        )
      )}

      {/* Navigation */}
      <div className="mt-8 text-center">
        <button
          onClick={handleGoHome}
          className="text-sm text-gray-500 hover:text-gray-700"
        >
          ← Back to Home
        </button>
      </div>
    </div>
  );
}