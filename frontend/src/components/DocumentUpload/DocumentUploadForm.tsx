import React, { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useDropzone } from 'react-dropzone';
import { DocumentUploadRequest, UploadProgressState, QRCodePosition } from '@/types/document';
import { ErrorDisplay } from '@/components/UI/ErrorDisplay';
import { LoadingButton } from '@/components/UI/LoadingSpinner';
import { AnimatedProgressBar } from '@/components/UI/ProgressIndicator';
import { useNotificationHelpers } from '@/components/UI/Notifications';

// Maximum file size: 50MB
const MAX_FILE_SIZE = 50 * 1024 * 1024;
const ACCEPTED_FILE_TYPES = ['application/pdf'];

interface DocumentUploadFormProps {
  onSubmit: (data: DocumentUploadRequest) => Promise<void>;
  isLoading?: boolean;
}

export default function DocumentUploadForm({ onSubmit, isLoading = false }: DocumentUploadFormProps) {
  const [uploadState, setUploadState] = useState<UploadProgressState>({
    isUploading: false,
    progress: 0,
  });
  const [showPositionOptions, setShowPositionOptions] = useState(false);
  const [submitError, setSubmitError] = useState<unknown>(null);
  const { showSuccess, showError } = useNotificationHelpers();
  
  const { register, handleSubmit, setValue, watch, formState: { errors }, reset } = useForm<DocumentUploadRequest>({
    defaultValues: {
      position: {
        page: undefined, // Default: last page
        x: undefined,    // Default: center
        y: undefined     // Default: bottom
      }
    }
  });
  
  const selectedFile = watch('file');
  const position = watch('position');

  // Reset form state when component unmounts
  useEffect(() => {
    return () => {
      reset();
      setUploadState({
        isUploading: false,
        progress: 0
      });
    };
  }, [reset]);

  const onDrop = React.useCallback((acceptedFiles: File[]) => {
    if (acceptedFiles.length > 0) {
      const file = acceptedFiles[0];
      
      // Validate file type
      if (!ACCEPTED_FILE_TYPES.includes(file.type)) {
        setUploadState(prev => ({ ...prev, error: 'Only PDF files are accepted' }));
        return;
      }
      
      // Validate file size
      if (file.size > MAX_FILE_SIZE) {
        setUploadState(prev => ({ ...prev, error: 'File size exceeds 50MB limit' }));
        return;
      }
      
      setValue('file', file);
      setUploadState(prev => ({ ...prev, error: undefined }));
    }
  }, [setValue]);

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
      
      setUploadState(prev => ({ ...prev, error: errorMessage }));
    }
  }, [fileRejections]);

  const handleFormSubmit = async (data: DocumentUploadRequest) => {
    try {
      setUploadState({ isUploading: true, progress: 0 });
      setSubmitError(null);
      
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setUploadState(prev => ({
          ...prev,
          progress: Math.min(prev.progress + 10, 90)
        }));
      }, 300);
      
      await onSubmit(data);
      
      clearInterval(progressInterval);
      setUploadState({
        isUploading: false,
        progress: 100,
        success: {
          id: 'temp-id', // This will be replaced by actual response
          filename: data.file.name,
          issuer: data.issuer,
          documentHash: 'hash-placeholder',
          createdAt: new Date().toISOString(),
          status: 'success'
        }
      });
      
      showSuccess('Document Uploaded', 'Your document has been successfully signed and is ready for download.');
      
    } catch (error) {
      clearInterval(progressInterval);
      setUploadState(prev => ({
        ...prev,
        isUploading: false,
        error: error instanceof Error ? error.message : 'Failed to upload document'
      }));
      setSubmitError(error);
      showError('Upload Failed', 'There was an error uploading your document. Please try again.');
    }
  };

  const handleRetry = async () => {
    if (selectedFile) {
      const formData = {
        file: selectedFile,
        issuer: watch('issuer'),
        description: watch('description'),
        position: watch('position'),
      };
      await handleFormSubmit(formData);
    }
  };

  return (
    <div className="w-full max-w-2xl mx-auto bg-white rounded-lg shadow-md p-6">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Upload Document for Signing</h2>
      
      {(uploadState.error || submitError) && (
        <div className="mb-4">
          <ErrorDisplay
            error={submitError || uploadState.error}
            title="Upload Error"
            onRetry={handleRetry}
            onDismiss={() => {
              setUploadState(prev => ({ ...prev, error: undefined }));
              setSubmitError(null);
            }}
            showDetails={true}
          />
        </div>
      )}
      
      {uploadState.success && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-md">
          <div className="flex items-center">
            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
            Document uploaded successfully!
          </div>
        </div>
      )}
      
      <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            PDF Document *
          </label>
          <div
            {...getRootProps()}
            className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors
              ${isDragActive ? 'border-blue-400 bg-blue-50' : 'border-gray-300 hover:border-gray-400'}
              ${isDragReject ? 'border-red-500 bg-red-50' : ''}
              ${errors.file ? 'border-red-500' : ''}
            `}
          >
            <input {...getInputProps()} data-testid="file-input" />
            
            {selectedFile ? (
              <div className="flex flex-col items-center">
                <svg className="w-8 h-8 text-gray-500 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p className="text-sm font-medium text-gray-900">{selectedFile.name}</p>
                <p className="text-xs text-gray-500">
                  {(selectedFile.size / (1024 * 1024)).toFixed(2)} MB
                </p>
                <button 
                  type="button"
                  className="mt-2 text-xs text-red-600 hover:text-red-800"
                  onClick={(e) => {
                    e.stopPropagation();
                    setValue('file', undefined as any);
                  }}
                >
                  Remove file
                </button>
              </div>
            ) : (
              <div className="flex flex-col items-center">
                <svg className="w-10 h-10 text-gray-400 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
                <p className="text-gray-600 mb-1">
                  {isDragActive ? 'Drop the PDF here' : 'Drag & drop your PDF here'}
                </p>
                <p className="text-xs text-gray-500">
                  Only PDF files up to 50MB are accepted
                </p>
                <button
                  type="button"
                  onClick={(e) => e.stopPropagation()}
                  className="mt-3 px-3 py-1 text-sm text-blue-600 border border-blue-600 rounded-md hover:bg-blue-50"
                >
                  Browse files
                </button>
              </div>
            )}
          </div>
          {errors.file && (
            <p className="mt-1 text-sm text-red-600">{errors.file.message}</p>
          )}
        </div>
        
        <div>
          <label htmlFor="issuer" className="block text-sm font-medium text-gray-700 mb-1">
            Issuer Name *
          </label>
          <input
            id="issuer"
            type="text"
            className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500
              ${errors.issuer ? 'border-red-500' : 'border-gray-300'}
            `}
            placeholder="Organization or person issuing this document"
            {...register('issuer', { 
              required: 'Issuer name is required',
              maxLength: {
                value: 255,
                message: 'Issuer name cannot exceed 255 characters'
              }
            })}
            data-testid="issuer-input"
          />
          {errors.issuer && (
            <p className="mt-1 text-sm text-red-600">{errors.issuer.message}</p>
          )}
        </div>
        
        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description (Optional)
          </label>
          <textarea
            id="description"
            rows={3}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Additional information about this document"
            {...register('description')}
            data-testid="description-input"
          />
        </div>
        
        <div className="border-t border-gray-200 pt-4">
          <div className="flex items-center mb-2">
            <input
              id="custom-position"
              type="checkbox"
              className="h-4 w-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
              checked={showPositionOptions}
              onChange={() => setShowPositionOptions(!showPositionOptions)}
              data-testid="custom-position-checkbox"
            />
            <label htmlFor="custom-position" className="ml-2 block text-sm text-gray-700">
              Customize QR Code position (optional)
            </label>
          </div>
          
          {showPositionOptions && (
            <div className="pl-6 space-y-3 mt-2">
              <div>
                <label htmlFor="page" className="block text-sm font-medium text-gray-700 mb-1">
                  Page Number (leave empty for last page)
                </label>
                <input
                  id="page"
                  type="number"
                  min="1"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Last page"
                  {...register('position.page', {
                    valueAsNumber: true,
                    min: {
                      value: 1,
                      message: 'Page number must be at least 1'
                    }
                  })}
                  data-testid="page-input"
                />
                {errors.position?.page && (
                  <p className="mt-1 text-sm text-red-600">{errors.position.page.message}</p>
                )}
              </div>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label htmlFor="x-position" className="block text-sm font-medium text-gray-700 mb-1">
                    X Position (0-100%)
                  </label>
                  <input
                    id="x-position"
                    type="number"
                    min="0"
                    max="100"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Center"
                    {...register('position.x', {
                      valueAsNumber: true,
                      min: {
                        value: 0,
                        message: 'X position must be between 0 and 100'
                      },
                      max: {
                        value: 100,
                        message: 'X position must be between 0 and 100'
                      }
                    })}
                    data-testid="x-position-input"
                  />
                  {errors.position?.x && (
                    <p className="mt-1 text-sm text-red-600">{errors.position.x.message}</p>
                  )}
                </div>
                
                <div>
                  <label htmlFor="y-position" className="block text-sm font-medium text-gray-700 mb-1">
                    Y Position (0-100%)
                  </label>
                  <input
                    id="y-position"
                    type="number"
                    min="0"
                    max="100"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="Bottom"
                    {...register('position.y', {
                      valueAsNumber: true,
                      min: {
                        value: 0,
                        message: 'Y position must be between 0 and 100'
                      },
                      max: {
                        value: 100,
                        message: 'Y position must be between 0 and 100'
                      }
                    })}
                    data-testid="y-position-input"
                  />
                  {errors.position?.y && (
                    <p className="mt-1 text-sm text-red-600">{errors.position.y.message}</p>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
        
        {uploadState.isUploading && (
          <div className="mt-4">
            <AnimatedProgressBar
              progress={uploadState.progress}
              showPercentage={true}
              color="blue"
              className="mb-2"
            />
            <p className="text-sm text-gray-600 text-center">
              Uploading and signing document...
            </p>
          </div>
        )}
        
        <div className="flex justify-end">
          <LoadingButton
            type="submit"
            isLoading={isLoading || uploadState.isUploading}
            loadingText="Uploading..."
            disabled={!selectedFile}
            variant="primary"
            size="md"
            data-testid="submit-button"
          >
            Sign Document
          </LoadingButton>
        </div>
      </form>
    </div>
  );
}
