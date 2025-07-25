import React, { useState, useEffect } from 'react';
import { useDropzone } from 'react-dropzone';
import { VerificationState } from '@/types/verification';

// Maximum file size: 50MB
const MAX_FILE_SIZE = 50 * 1024 * 1024;
const ACCEPTED_FILE_TYPES = ['application/pdf'];

interface DocumentVerificationUploadProps {
  onUpload: (file: File) => Promise<void>;
  isLoading?: boolean;
}

export default function DocumentVerificationUpload({ onUpload, isLoading = false }: DocumentVerificationUploadProps) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [verificationState, setVerificationState] = useState<VerificationState>({
    isVerifying: false,
    progress: 0,
  });

  // Reset state when component unmounts
  useEffect(() => {
    return () => {
      setSelectedFile(null);
      setVerificationState({
        isVerifying: false,
        progress: 0
      });
    };
  }, []);

  const onDrop = React.useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles.length > 0) {
      const file = acceptedFiles[0];
      
      // Validate file type
      if (!ACCEPTED_FILE_TYPES.includes(file.type)) {
        setVerificationState(prev => ({ ...prev, error: 'Only PDF files are accepted' }));
        return;
      }
      
      // Validate file size
      if (file.size > MAX_FILE_SIZE) {
        setVerificationState(prev => ({ ...prev, error: 'File size exceeds 50MB limit' }));
        return;
      }
      
      setSelectedFile(file);
      setVerificationState(prev => ({ ...prev, error: undefined }));
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive, isDragReject, fileRejections } = useDropzone({
    onDrop,
    accept: {
      'application/pdf': ['.pdf']
    },
    maxSize: MAX_FILE_SIZE,
    multiple: false
  });

  // Handle file rejection messages
  useEffect(() => {
    if (fileRejections.length > 0) {
      const rejection = fileRejections[0];
      let errorMessage = 'Invalid file';
      
      if (rejection.errors.some(e => e.code === 'file-too-large')) {
        errorMessage = `File size exceeds 50MB limit (${(rejection.file.size / (1024 * 1024)).toFixed(2)} MB)`;
      } else if (rejection.errors.some(e => e.code === 'file-invalid-type')) {
        errorMessage = 'Only PDF files are accepted';
      }
      
      setVerificationState(prev => ({ ...prev, error: errorMessage }));
    }
  }, [fileRejections]);

  const handleVerify = async () => {
    if (!selectedFile) return;
    
    try {
      setVerificationState({ isVerifying: true, progress: 0 });
      
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setVerificationState(prev => ({
          ...prev,
          progress: Math.min(prev.progress + 10, 90)
        }));
      }, 300);
      
      await onUpload(selectedFile);
      
      clearInterval(progressInterval);
      setVerificationState(prev => ({
        ...prev,
        isVerifying: false,
        progress: 100
      }));
    } catch (error) {
      setVerificationState(prev => ({
        ...prev,
        isVerifying: false,
        error: error instanceof Error ? error.message : 'Failed to verify document'
      }));
    }
  };

  return (
    <div className="w-full max-w-md mx-auto bg-white rounded-lg shadow-md p-4 sm:p-6 transition-all duration-300">
      <h2 className="text-lg sm:text-xl font-bold mb-4 text-gray-800 flex items-center">
        <svg className="w-5 h-5 mr-2 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        Upload Document for Verification
      </h2>
      
      {verificationState.error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md animate-fadeIn">
          <div className="flex items-center">
            <svg className="w-5 h-5 mr-2 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <span className="text-sm">{verificationState.error}</span>
          </div>
        </div>
      )}
      
      <div
        {...getRootProps()}
        className={`border-2 border-dashed rounded-lg p-4 sm:p-6 text-center cursor-pointer transition-colors mb-4
          ${isDragActive ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}
          ${isDragReject ? 'border-red-500 bg-red-50' : ''}
          focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
        `}
        data-testid="dropzone"
      >
        <input {...getInputProps()} data-testid="file-input" />
        
        {selectedFile ? (
          <div className="flex flex-col items-center">
            <svg className="w-8 h-8 text-blue-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p className="text-sm font-medium text-gray-900 break-all">{selectedFile.name}</p>
            <p className="text-xs text-gray-500">
              {(selectedFile.size / (1024 * 1024)).toFixed(2)} MB
            </p>
            <button 
              type="button"
              className="mt-2 text-xs text-red-600 hover:text-red-800 transition-colors duration-200"
              onClick={(e) => {
                e.stopPropagation();
                setSelectedFile(null);
              }}
              data-testid="remove-file-button"
            >
              Remove file
            </button>
          </div>
        ) : (
          <div className="flex flex-col items-center">
            <svg className="w-10 h-10 text-blue-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <p className="text-gray-600 mb-1 text-sm sm:text-base">
              {isDragActive ? 'Drop the PDF here' : 'Drag & drop your PDF here'}
            </p>
            <p className="text-xs text-gray-500">
              Only PDF files up to 50MB are accepted
            </p>
            <button
              type="button"
              onClick={(e) => e.stopPropagation()}
              className="mt-3 px-3 py-1 text-sm text-blue-600 border border-blue-600 rounded-md hover:bg-blue-50 transition-colors duration-200"
              data-testid="browse-button"
            >
              Browse files
            </button>
          </div>
        )}
      </div>
      
      {verificationState.isVerifying && (
        <div className="mt-4 mb-4 animate-fadeIn">
          <div className="flex justify-between text-sm text-gray-600 mb-1">
            <span>Verifying document...</span>
            <span>{verificationState.progress}%</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5 overflow-hidden">
            <div 
              className="bg-blue-600 h-2.5 rounded-full transition-all duration-300" 
              style={{ width: `${verificationState.progress}%` }}
            ></div>
          </div>
        </div>
      )}
      
      <div className="flex justify-center">
        <button
          type="button"
          onClick={handleVerify}
          disabled={isLoading || verificationState.isVerifying || !selectedFile}
          className={`px-4 py-2 rounded-md text-white font-medium w-full transition-colors duration-200
            ${isLoading || verificationState.isVerifying || !selectedFile
              ? 'bg-gray-400 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500'}
          `}
          data-testid="verify-button"
        >
          {isLoading || verificationState.isVerifying ? (
            <span className="flex items-center justify-center">
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Verifying...
            </span>
          ) : 'Verify Document'}
        </button>
      </div>
    </div>
  );
}