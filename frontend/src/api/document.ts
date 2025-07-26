import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { DocumentUploadRequest, DocumentUploadResponse, DocumentListResponse, Document } from '@/types/document';
import { get, post, del } from './client';

// API functions
export async function uploadDocument(data: DocumentUploadRequest): Promise<DocumentUploadResponse> {
  const formData = new FormData();
  formData.append('file', data.file);
  formData.append('issuer', data.issuer);
  
  if (data.description) {
    formData.append('description', data.description);
  }
  
  if (data.position) {
    formData.append('position', JSON.stringify(data.position));
  }
  
  return post<DocumentUploadResponse>(
    '/api/documents/sign',
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  );
}

export async function getDocuments(page = 1, limit = 10, search?: string): Promise<DocumentListResponse> {
  const params: Record<string, string | number> = {
    page,
    limit,
  };
  
  if (search) {
    params.search = search;
  }
  
  return get<DocumentListResponse>('/api/documents', params);
}

export async function getDocumentById(id: string): Promise<Document> {
  return get<Document>(`/api/documents/${id}`);
}

export async function deleteDocument(id: string): Promise<{ success: boolean }> {
  return del<{ success: boolean }>(`/api/documents/${id}`);
}

// React Query hooks
export function useDocuments(page = 1, limit = 10, search?: string) {
  return useQuery({
    queryKey: ['documents', page, limit, search],
    queryFn: () => getDocuments(page, limit, search),
    placeholderData: (previousData) => previousData, // Keep previous data while loading new data
    staleTime: 2 * 60 * 1000, // 2 minutes for document list
    select: (data) => ({
      ...data,
      documents: data.documents.map(doc => ({
        ...doc,
        createdAt: new Date(doc.createdAt).toISOString(),
        updatedAt: new Date(doc.updatedAt).toISOString(),
      })),
    }),
  });
}

export function useDocument(id: string) {
  return useQuery({
    queryKey: ['document', id],
    queryFn: () => getDocumentById(id),
    enabled: !!id,
    staleTime: 5 * 60 * 1000, // 5 minutes for individual documents
    select: (data) => ({
      ...data,
      createdAt: new Date(data.createdAt).toISOString(),
      updatedAt: new Date(data.updatedAt).toISOString(),
    }),
  });
}

export function useUploadDocument() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: uploadDocument,
    onSuccess: (data) => {
      // Invalidate documents query to refetch the list
      queryClient.invalidateQueries({ queryKey: ['documents'] });
      
      // Optimistically update the cache with the new document
      queryClient.setQueryData(['document', data.id], data);
      
      // Update documents list cache if possible
      queryClient.setQueriesData(
        { queryKey: ['documents'] },
        (oldData: DocumentListResponse | undefined) => {
          if (!oldData) return oldData;
          
          return {
            ...oldData,
            documents: [data, ...oldData.documents],
            total: oldData.total + 1,
          };
        }
      );
    },
    onError: (error) => {
      // Log upload errors for debugging
      console.error('Document upload failed:', error);
    },
  });
}

export function useDeleteDocument() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: deleteDocument,
    onMutate: async (documentId) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['documents'] });
      await queryClient.cancelQueries({ queryKey: ['document', documentId] });
      
      // Snapshot the previous value
      const previousDocuments = queryClient.getQueriesData({ queryKey: ['documents'] });
      
      // Optimistically remove the document from cache
      queryClient.setQueriesData(
        { queryKey: ['documents'] },
        (oldData: DocumentListResponse | undefined) => {
          if (!oldData) return oldData;
          
          return {
            ...oldData,
            documents: oldData.documents.filter((doc: Document) => doc.id !== documentId),
            total: Math.max(0, oldData.total - 1),
          };
        }
      );
      
      // Remove individual document from cache
      queryClient.removeQueries({ queryKey: ['document', documentId] });
      
      return { previousDocuments };
    },
    onError: (error, documentId, context) => {
      // Rollback on error
      if (context?.previousDocuments) {
        context.previousDocuments.forEach(([queryKey, data]) => {
          queryClient.setQueryData(queryKey, data);
        });
      }
      console.error('Document deletion failed:', error);
    },
    onSettled: () => {
      // Always refetch after error or success
      queryClient.invalidateQueries({ queryKey: ['documents'] });
    },
  });
}
