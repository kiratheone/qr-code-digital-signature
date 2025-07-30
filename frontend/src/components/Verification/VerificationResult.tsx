import React from 'react';
import { VerificationResponse, VerificationStatus } from '@/types/verification';
import { formatDate } from '@/utils/dateUtils';
import { ErrorDisplay } from '@/components/UI/ErrorDisplay';

interface VerificationResultProps {
  result: VerificationResponse;
  onRetry?: () => Promise<void>;
  error?: unknown;
}

export default function VerificationResult({ result, onRetry, error }: VerificationResultProps) {
  // Show error display if there's an error
  if (error) {
    return (
      <div className="w-full max-w-md mx-auto">
        <ErrorDisplay
          error={error}
          title="Verification Failed"
          onRetry={onRetry}
          showDetails={true}
          variant="modal"
        />
      </div>
    );
  }
  const getStatusIcon = (status: VerificationStatus) => {
    switch (status) {
      case 'valid':
        return (
          <div className="flex items-center justify-center w-16 h-16 bg-green-100 rounded-full mb-4 animate-fadeIn">
            <svg className="w-10 h-10 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
        );
      case 'modified':
        return (
          <div className="flex items-center justify-center w-16 h-16 bg-yellow-100 rounded-full mb-4 animate-fadeIn">
            <svg className="w-10 h-10 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
        );
      case 'invalid':
        return (
          <div className="flex items-center justify-center w-16 h-16 bg-red-100 rounded-full mb-4 animate-fadeIn">
            <svg className="w-10 h-10 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
        );
      default:
        return (
          <div className="flex items-center justify-center w-16 h-16 bg-gray-100 rounded-full mb-4 animate-fadeIn">
            <svg className="w-10 h-10 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        );
    }
  };

  const getStatusBadge = (status: VerificationStatus) => {
    switch (status) {
      case 'valid':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
            ✅ Valid
          </span>
        );
      case 'modified':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
            ⚠️ Modified
          </span>
        );
      case 'invalid':
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
            ❌ Invalid
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
            Pending
          </span>
        );
    }
  };

  const getStatusColor = (status: VerificationStatus) => {
    switch (status) {
      case 'valid':
        return 'text-green-700 bg-green-50 border-green-200';
      case 'modified':
        return 'text-yellow-700 bg-yellow-50 border-yellow-200';
      case 'invalid':
        return 'text-red-700 bg-red-50 border-red-200';
      default:
        return 'text-gray-700 bg-gray-50 border-gray-200';
    }
  };

  const getStatusTitle = (status: VerificationStatus) => {
    switch (status) {
      case 'valid':
        return '✅ Document is valid';
      case 'modified':
        return '⚠️ QR valid, but file content has changed';
      case 'invalid':
        return '❌ QR invalid / signature incorrect';
      default:
        return 'Verification pending';
    }
  };

  const getStatusExplanation = (status: VerificationStatus) => {
    switch (status) {
      case 'valid':
        return 'This document is authentic and has not been modified since it was signed. The digital signature is valid.';
      case 'modified':
        return 'The QR code is authentic, but the document content has been modified since it was signed. The document may have been tampered with.';
      case 'invalid':
        return 'The digital signature is invalid. This document may be forged or the QR code may have been tampered with.';
      default:
        return 'Verification is in progress.';
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className={`p-4 sm:p-6 rounded-lg shadow-md border ${getStatusColor(result.status)}`} data-testid="verification-result">
        <div className="flex flex-col items-center text-center mb-4">
          {getStatusIcon(result.status)}
          <h2 className="text-xl font-bold mb-1">{getStatusTitle(result.status)}</h2>
          <p className="text-sm mb-2">{result.message}</p>
          <p className="text-xs text-gray-600 px-2">{getStatusExplanation(result.status)}</p>
        </div>

        <div className="border-t border-gray-200 pt-4 mt-2">
          <h3 className="text-lg font-semibold mb-3 flex items-center">
            <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Document Information
          </h3>
          
          <div className="space-y-3">
            <div className="bg-white bg-opacity-50 p-2 rounded-md">
              <p className="text-sm font-medium text-gray-500">Document Name</p>
              <p className="text-sm sm:text-base font-medium break-words">{result.filename}</p>
            </div>
            
            <div className="bg-white bg-opacity-50 p-2 rounded-md">
              <p className="text-sm font-medium text-gray-500">Issuer</p>
              <p className="text-sm sm:text-base">{result.issuer}</p>
            </div>
            
            <div className="bg-white bg-opacity-50 p-2 rounded-md">
              <p className="text-sm font-medium text-gray-500">Creation Date</p>
              <p className="text-sm sm:text-base">{formatDate(result.createdAt)}</p>
            </div>
            
            <div className="bg-white bg-opacity-50 p-2 rounded-md">
              <p className="text-sm font-medium text-gray-500">Status</p>
              <div className="mt-1">{getStatusBadge(result.status)}</div>
            </div>
            
            {result.details && (
              <div className="bg-white bg-opacity-50 p-2 rounded-md">
                <p className="text-sm font-medium text-gray-500 flex items-center">
                  <svg className="w-4 h-4 mr-1 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  Additional Details
                </p>
                <div className="text-sm whitespace-pre-wrap bg-gray-50 p-2 rounded-md mt-1 border border-gray-100">
                  {result.details}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}