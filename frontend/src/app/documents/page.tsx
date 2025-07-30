'use client';

import { useState, useEffect, lazy, Suspense } from 'react';
import { getDocuments, deleteDocument } from '@/api/document';
import { Document } from '@/types/document';
import Link from 'next/link';
import DocumentList from '@/components/DocumentManagement/DocumentList';
import SearchBar from '@/components/DocumentManagement/SearchBar';
import Pagination from '@/components/DocumentManagement/Pagination';
import { LoadingSpinner } from '@/components/UI/LoadingSpinner';

// Lazy load heavy components
const DocumentDetailsModal = lazy(() => import('@/components/DocumentManagement/DocumentDetailsModal'));
const DeleteConfirmationDialog = lazy(() => import('@/components/DocumentManagement/DeleteConfirmationDialog'));

export default function DocumentManagementPage() {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedDocument, setSelectedDocument] = useState<Document | null>(null);
  const [documentToDelete, setDocumentToDelete] = useState<Document | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const [isDetailsModalOpen, setIsDetailsModalOpen] = useState(false);
  const limit = 10;

  const fetchDocuments = async (page = 1, search = '') => {
    setIsLoading(true);
    try {
      const response = await getDocuments(page, limit, search);
      setDocuments(response.documents);
      setTotalPages(Math.ceil(response.total / limit));
      setCurrentPage(page);
    } catch (err) {
      console.error('Error fetching documents:', err);
      setError('Failed to load documents. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchDocuments(currentPage, searchQuery);
  }, [currentPage, searchQuery]);

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setCurrentPage(1); // Reset to first page on new search
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleViewDetails = (document: Document) => {
    setSelectedDocument(document);
    setIsDetailsModalOpen(true);
  };

  const handleDeleteClick = (document: Document) => {
    setDocumentToDelete(document);
    setIsDeleteModalOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!documentToDelete) return;
    
    try {
      await deleteDocument(documentToDelete.id);
      fetchDocuments(currentPage, searchQuery);
      setIsDeleteModalOpen(false);
      setDocumentToDelete(null);
    } catch (err) {
      console.error('Error deleting document:', err);
      setError('Failed to delete document. Please try again.');
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-gray-900">Document Management</h1>
          <Link 
            href="/documents/upload" 
            className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            Sign New Document
          </Link>
        </div>
        <p className="text-gray-600 mt-2">
          Manage your digitally signed documents
        </p>
      </div>

      <div className="mb-6">
        <SearchBar onSearch={handleSearch} />
      </div>

      {error && (
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-md text-red-700">
          <div className="flex items-start">
            <svg className="w-5 h-5 mr-2 mt-0.5" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div>
              <p className="font-medium">Error</p>
              <p>{error}</p>
            </div>
          </div>
        </div>
      )}

      <DocumentList 
        documents={documents}
        isLoading={isLoading}
        onViewDetails={handleViewDetails}
        onDelete={handleDeleteClick}
      />

      <div className="mt-6">
        <Pagination 
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
        />
      </div>

      {isDetailsModalOpen && selectedDocument && (
        <Suspense fallback={<LoadingSpinner />}>
          <DocumentDetailsModal
            document={selectedDocument}
            onClose={() => setIsDetailsModalOpen(false)}
          />
        </Suspense>
      )}

      {isDeleteModalOpen && documentToDelete && (
        <Suspense fallback={<LoadingSpinner />}>
          <DeleteConfirmationDialog
            document={documentToDelete}
            onConfirm={handleDeleteConfirm}
            onCancel={() => setIsDeleteModalOpen(false)}
          />
        </Suspense>
      )}
    </div>
  );
}