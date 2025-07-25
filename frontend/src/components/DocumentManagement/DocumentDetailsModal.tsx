import React from 'react';

interface Document {
  id: string;
  filename: string;
  issuer: string;
  documentHash: string;
  createdAt: string;
  status: string;
  qrCodeData?: string;
  signatureData?: string;
  fileSize?: number;
}

interface DocumentDetailsModalProps {
  document: Document;
  onClose: () => void;
}

export default function DocumentDetailsModal({ document, onClose }: DocumentDetailsModalProps) {
  return (
    <div className="fixed inset-0 z-10 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        <div className="fixed inset-0 transition-opacity" aria-hidden="true">
          <div className="absolute inset-0 bg-gray-500 opacity-75"></div>
        </div>

        <span className="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

        <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
          <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
            <div className="sm:flex sm:items-start">
              <div className="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Document Details
                </h3>
                <div className="mt-2 space-y-3">
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Document ID</p>
                    <p className="text-sm text-gray-900 break-all">{document.id}</p>
                  </div>
                  
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Filename</p>
                    <p className="text-sm text-gray-900">{document.filename}</p>
                  </div>
                  
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Issuer</p>
                    <p className="text-sm text-gray-900">{document.issuer}</p>
                  </div>
                  
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Created At</p>
                    <p className="text-sm text-gray-900">{new Date(document.createdAt).toLocaleString()}</p>
                  </div>
                  
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Document Hash</p>
                    <p className="text-sm text-gray-900 break-all">{document.documentHash}</p>
                  </div>
                  
                  {document.fileSize && (
                    <div className="bg-gray-50 p-3 rounded-md">
                      <p className="text-sm font-medium text-gray-500">File Size</p>
                      <p className="text-sm text-gray-900">{(document.fileSize / (1024 * 1024)).toFixed(2)} MB</p>
                    </div>
                  )}
                  
                  <div className="bg-gray-50 p-3 rounded-md">
                    <p className="text-sm font-medium text-gray-500">Status</p>
                    <p className="text-sm">
                      <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                        {document.status}
                      </span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
            <a
              href={`/api/documents/${document.id}/download`}
              className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
              target="_blank"
              rel="noopener noreferrer"
            >
              Download Document
            </a>
            <button
              type="button"
              className="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
              onClick={onClose}
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}