/**
 * VerificationResult Component
 * Presentation component for displaying document verification results
 */

import React from 'react';
import { Button } from './ui/Button';
import type { VerificationResult } from '@/lib/types';

interface VerificationResultProps {
  result: VerificationResult;
  onVerifyAnother?: () => void;
  onGoHome?: () => void;
}

export function VerificationResult({
  result,
  onVerifyAnother,
  onGoHome,
}: VerificationResultProps) {
  const getStatusInfo = () => {
    switch (result.status) {
      case 'valid':
        return {
          icon: '✅',
          color: 'text-green-600',
          bgColor: 'bg-green-50',
          borderColor: 'border-green-200',
          title: 'Document is Valid',
          description: 'The document is authentic and has not been modified since signing.',
        };
      
      case 'modified':
        return {
          icon: '⚠️',
          color: 'text-yellow-600',
          bgColor: 'bg-yellow-50',
          borderColor: 'border-yellow-200',
          title: 'Document Modified',
          description: 'The QR code is valid, but the document content has been changed since signing.',
        };
      
      case 'invalid':
        return {
          icon: '❌',
          color: 'text-red-600',
          bgColor: 'bg-red-50',
          borderColor: 'border-red-200',
          title: 'Document Invalid',
          description: 'The document signature is invalid or the QR code is corrupted.',
        };
      
      default:
        return {
          icon: '❓',
          color: 'text-gray-600',
          bgColor: 'bg-gray-50',
          borderColor: 'border-gray-200',
          title: 'Unknown Status',
          description: 'Unable to determine document verification status.',
        };
    }
  };

  const statusInfo = getStatusInfo();

  const formatDate = (dateString: string): string => {
    try {
      return new Date(dateString).toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch {
      return 'Invalid date';
    }
  };

  const getDetailStatus = (isValid: boolean) => {
    return isValid ? 'success' : 'error';
  };

  const getDetailIcon = (status: 'success' | 'error') => {
    return status === 'success' ? '✓' : '✗';
  };

  const getDetailColor = (status: 'success' | 'error') => {
    return status === 'success' ? 'text-green-600' : 'text-red-600';
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* Main Result */}
      <div className={`bg-white shadow rounded-lg p-6 mb-6 border-l-4 ${statusInfo.borderColor}`}>
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <span className="text-3xl">{statusInfo.icon}</span>
          </div>
          <div className="ml-4">
            <h2 className={`text-xl font-semibold ${statusInfo.color}`}>
              {statusInfo.title}
            </h2>
            <p className="mt-1 text-sm text-gray-600">
              {statusInfo.description}
            </p>
          </div>
        </div>

        <div className="mt-4 text-sm text-gray-500">
          <p>Verified on {formatDate(result.verified_at)}</p>
        </div>
      </div>

      {/* Detailed Results */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Verification Details</h3>
        
        <div className="space-y-4">
          {/* QR Code Validity */}
          <div className="flex items-center justify-between py-2 border-b border-gray-100">
            <div className="flex items-center">
              <span className={`text-lg mr-3 ${getDetailColor(getDetailStatus(result.details.qr_valid))}`}>
                {getDetailIcon(getDetailStatus(result.details.qr_valid))}
              </span>
              <span className="text-sm font-medium text-gray-900">QR Code Validity</span>
            </div>
            <span className={`text-sm ${getDetailColor(getDetailStatus(result.details.qr_valid))}`}>
              {result.details.qr_valid ? 'Valid' : 'Invalid'}
            </span>
          </div>

          {/* Document Hash Match */}
          <div className="flex items-center justify-between py-2 border-b border-gray-100">
            <div className="flex items-center">
              <span className={`text-lg mr-3 ${getDetailColor(getDetailStatus(result.details.hash_matches))}`}>
                {getDetailIcon(getDetailStatus(result.details.hash_matches))}
              </span>
              <span className="text-sm font-medium text-gray-900">Document Hash Match</span>
            </div>
            <span className={`text-sm ${getDetailColor(getDetailStatus(result.details.hash_matches))}`}>
              {result.details.hash_matches ? 'Matches' : 'Does not match'}
            </span>
          </div>

          {/* Digital Signature */}
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center">
              <span className={`text-lg mr-3 ${getDetailColor(getDetailStatus(result.details.signature_valid))}`}>
                {getDetailIcon(getDetailStatus(result.details.signature_valid))}
              </span>
              <span className="text-sm font-medium text-gray-900">Digital Signature</span>
            </div>
            <span className={`text-sm ${getDetailColor(getDetailStatus(result.details.signature_valid))}`}>
              {result.details.signature_valid ? 'Valid' : 'Invalid'}
            </span>
          </div>
        </div>
      </div>

      {/* Hash Comparison */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Hash Comparison</h3>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Original Document Hash</label>
            <div className="mt-1 p-2 bg-gray-50 rounded-md">
              <code className="text-xs text-gray-900 break-all">{result.details.original_hash}</code>
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700">Uploaded Document Hash</label>
            <div className="mt-1 p-2 bg-gray-50 rounded-md">
              <code className="text-xs text-gray-900 break-all">{result.details.uploaded_hash}</code>
            </div>
          </div>

          {result.details.hash_matches ? (
            <div className="flex items-center text-sm text-green-600">
              <svg className="h-4 w-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
              </svg>
              Hashes match - document content is unchanged
            </div>
          ) : (
            <div className="flex items-center text-sm text-red-600">
              <svg className="h-4 w-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
              Hashes do not match - document content has been modified
            </div>
          )}
        </div>
      </div>

      {/* Additional Information */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Additional Information</h3>
        
        <div className="text-sm text-gray-600">
          <p className="mb-2">
            <strong>Message:</strong> {result.message}
          </p>
          
          <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-md">
            <h4 className="text-sm font-medium text-blue-800 mb-2">What this means:</h4>
            <div className="text-sm text-blue-700">
              {result.status === 'valid' && (
                <p>
                  This document is authentic and has not been tampered with since it was digitally signed. 
                  You can trust that the content is exactly as it was when originally signed.
                </p>
              )}
              {result.status === 'modified' && (
                <p>
                  While the QR code and signature are valid, the document content has been changed since signing. 
                  This could indicate tampering or unauthorized modifications.
                </p>
              )}
              {result.status === 'invalid' && (
                <p>
                  The document verification failed. This could mean the document is forged, the QR code is corrupted, 
                  or the digital signature is invalid. Do not trust this document&apos;s authenticity.
                </p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex justify-center space-x-4">
        {onVerifyAnother && (
          <Button
            variant="secondary"
            onClick={onVerifyAnother}
          >
            Verify Another Document
          </Button>
        )}
        
        {onGoHome && (
          <Button
            variant="primary"
            onClick={onGoHome}
          >
            Go to Home
          </Button>
        )}
      </div>
    </div>
  );
}