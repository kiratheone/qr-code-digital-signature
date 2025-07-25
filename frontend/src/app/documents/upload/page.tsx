'use client';

import { useState } from 'react';
import DocumentUploadForm from '@/components/DocumentUpload/DocumentUploadForm';
import { DocumentUploadRequest, DocumentUploadResponse } from '@/types/document';
import { uploadDocument } from '@/api/document';
import Link from 'next/link';

export default function DocumentUploadPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<{success?: DocumentUploadResponse; error?: string}>({}); 

  const handleSubmit = async (data: DocumentUploadRequest) => {
    setIsLoading(true);
    setResult({});
    
    try {
      // Call the API to upload and sign the document
      const response = await uploadDocument(data);
      
      setResult({ success: response });
      return response;
    } catch (error) {
      console.error('Error signing document:', error);
      setResult({ 
        error: error instanceof Error 
          ? error.message 
          : 'Failed to sign document. Please try again.' 
      });
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const handleReset = () => {
    setResult({});
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Sign Document</h1>
          <Link 
            href="/documents" 
            className="text-blue-600 hover:text-blue-800 flex items-center"
          >
            <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            Back to Documents
          </Link>
        </div>
        <p className="text-gray-600 mt-2">
          Upload a PDF document to sign it with a secure digital signature and QR code.
        </p>
      </div>
      
      {result.error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md text-red-700">
          <div className="flex items-start">
            <svg className="w-5 h-5 mr-2 mt-0.5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div>
              <p className="font-medium">Error</p>
              <p>{result.error}</p>
              <button 
                onClick={handleReset}
                className="mt-2 text-sm text-red-700 hover:text-red-900 underline"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      )}
      
      {result.success && (
        <div className="mb-6 p-6 bg-green-50 border border-green-200 rounded-md text-green-700">
          <div className="flex items-center mb-4">
            <svg className="w-8 h-8 mr-3 text-green-500" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            <h2 className="text-xl font-bold">Document Signed Successfully!</h2>
          </div>
          
          <div className="bg-white p-4 rounded-md border border-green-200 mb-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <p className="text-sm text-gray-500">Document ID</p>
                <p className="font-medium">{result.success.id}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Filename</p>
                <p className="font-medium">{result.success.filename}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Issuer</p>
                <p className="font-medium">{result.success.issuer}</p>
              </div>
              <div>
                <p className="text-sm text-gray-500">Created</p>
                <p className="font-medium">{new Date(result.success.createdAt).toLocaleString()}</p>
              </div>
            </div>
          </div>
          
          <div className="flex flex-wrap gap-3">
            <a 
              href={`/api/documents/${result.success.id}/download`} 
              className="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
              target="_blank"
              rel="noopener noreferrer"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download Signed Document
            </a>
            
            <Link
              href={`/documents/${result.success.id}`}
              className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              View Document Details
            </Link>
            
            <button
              onClick={handleReset}
              className="inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-gray-700 rounded-md hover:bg-gray-50"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              Sign Another Document
            </button>
          </div>
        </div>
      )}
      
      {!result.success && (
        <>
          <DocumentUploadForm onSubmit={handleSubmit} isLoading={isLoading} />
          
          <div className="mt-8 bg-blue-50 border border-blue-200 rounded-md p-4">
            <h3 className="text-lg font-medium text-blue-800 mb-2">How it works</h3>
            <ol className="list-decimal list-inside text-blue-700 space-y-1 pl-2">
              <li>Upload your PDF document (max 50MB)</li>
              <li>Enter the issuer information</li>
              <li>Optionally customize QR code position</li>
              <li>Our system will generate a secure hash of your document</li>
              <li>A digital signature will be created using RSA encryption</li>
              <li>A QR code containing verification information will be added to your document</li>
              <li>Download your signed document with embedded QR code</li>
            </ol>
          </div>
          
          <div className="mt-6 bg-yellow-50 border border-yellow-200 rounded-md p-4">
            <div className="flex items-start">
              <svg className="w-5 h-5 text-yellow-600 mr-2 mt-0.5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
              </svg>
              <div>
                <h3 className="text-sm font-medium text-yellow-800">Important Note</h3>
                <p className="text-sm text-yellow-700 mt-1">
                  Once a document is signed, the original content cannot be modified without invalidating the signature. 
                  Make sure your document is finalized before signing.
                </p>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
