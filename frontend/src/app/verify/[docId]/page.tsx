'use client';

import React, { useState, useEffect, lazy, Suspense } from 'react';
import { useParams } from 'next/navigation';
import { getVerificationInfo, verifyDocument } from '@/api/verification';
import { VerificationResponse, VerificationState } from '@/types/verification';
import { formatDate } from '@/utils/dateUtils';

// Lazy load verification components
const DocumentVerificationUpload = lazy(() => import('@/components/Verification/DocumentVerificationUpload'));
const VerificationResult = lazy(() => import('@/components/Verification/VerificationResult'));

export default function VerifyPage() {
  const params = useParams();
  const docId = params.docId as string;
  
  const [documentInfo, setDocumentInfo] = useState<VerificationResponse | null>(null);
  const [verificationState, setVerificationState] = useState<VerificationState>({
    isVerifying: false,
    progress: 0,
  });
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  useEffect(() => {
    async function fetchDocumentInfo() {
      setIsLoading(true);
      try {
        const info = await getVerificationInfo(docId);
        setDocumentInfo(info);
        setError(null);
      } catch (err) {
        setError('Failed to load document information. The document may not exist or the QR code is invalid.');
        setDocumentInfo(null);
      } finally {
        setIsLoading(false);
      }
    }

    if (docId) {
      fetchDocumentInfo();
    }
  }, [docId]);

  const handleVerify = async (file: File) => {
    try {
      setVerificationState({
        isVerifying: true,
        progress: 0,
      });
      
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setVerificationState(prev => ({
          ...prev,
          progress: Math.min(prev.progress + 10, 90)
        }));
      }, 300);
      
      const result = await verifyDocument(docId, file);
      
      clearInterval(progressInterval);
      setVerificationState({
        isVerifying: false,
        progress: 100,
        result,
      });
    } catch (err) {
      setVerificationState(prev => ({
        ...prev,
        isVerifying: false,
        error: err instanceof Error ? err.message : 'Failed to verify document',
      }));
    }
  };

  const handleVerifyAnother = () => {
    setVerificationState({ isVerifying: false, progress: 0 });
  };

  return (
    <div className="container mx-auto px-4 py-6 sm:py-8">
      <div className="max-w-3xl mx-auto">
        <header className="mb-6">
          <h1 className="text-2xl sm:text-3xl font-bold text-center text-gray-800">Document Verification</h1>
          <p className="text-center text-gray-600 mt-2">Verify the authenticity of your document</p>
        </header>
        
        {isLoading && (
          <div className="flex justify-center items-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          </div>
        )}
        
        {error && !isLoading && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-md shadow-sm">
            <div className="flex items-center">
              <svg className="w-5 h-5 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
              <span className="text-sm sm:text-base">{error}</span>
            </div>
          </div>
        )}
        
        {documentInfo && !verificationState.result && !isLoading && (
          <div className="mb-6">
            <div className="bg-white p-5 sm:p-6 rounded-lg shadow-md border border-gray-200">
              <h2 className="text-lg sm:text-xl font-semibold mb-4 flex items-center">
                <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Document Information
              </h2>
              <div className="space-y-3">
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <div className="col-span-1 sm:col-span-3">
                    <p className="text-sm font-medium text-gray-500">Document Name</p>
                    <p className="text-sm sm:text-base font-medium text-gray-800 break-words">{documentInfo.filename}</p>
                  </div>
                  <div className="col-span-1 sm:col-span-2">
                    <p className="text-sm font-medium text-gray-500">Issuer</p>
                    <p className="text-sm sm:text-base text-gray-800">{documentInfo.issuer}</p>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-gray-500">Creation Date</p>
                    <p className="text-sm sm:text-base text-gray-800">{formatDate(documentInfo.createdAt)}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
        
        {!verificationState.result && !isLoading ? (
          <Suspense fallback={
            <div className="flex justify-center items-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
            </div>
          }>
            <DocumentVerificationUpload 
              onUpload={handleVerify} 
              isLoading={verificationState.isVerifying} 
            />
          </Suspense>
        ) : verificationState.result && (
          <div className="mt-6 animate-fadeIn">
            <Suspense fallback={
              <div className="flex justify-center items-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
              </div>
            }>
              <VerificationResult result={verificationState.result} />
            </Suspense>
            <div className="mt-6 text-center">
              <button
                onClick={handleVerifyAnother}
                className="px-4 py-2 bg-blue-100 hover:bg-blue-200 rounded-md text-blue-800 font-medium transition-colors duration-200"
                data-testid="verify-another-button"
              >
                Verify Another Document
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}