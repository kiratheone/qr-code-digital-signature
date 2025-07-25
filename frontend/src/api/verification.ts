import { useMutation, useQuery } from '@tanstack/react-query';
import { VerificationResponse } from '@/types/verification';
import { get, post } from './client';

// API functions
export async function getVerificationInfo(docId: string): Promise<VerificationResponse> {
  return get<VerificationResponse>(`/api/verify/${docId}`);
}

export async function verifyDocument(docId: string, file: File): Promise<VerificationResponse> {
  const formData = new FormData();
  formData.append('file', file);
  
  return post<VerificationResponse>(
    `/api/verify/${docId}/upload`,
    formData,
    {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    }
  );
}

// React Query hooks
export function useVerificationInfo(docId: string) {
  return useQuery({
    queryKey: ['verification', docId],
    queryFn: () => getVerificationInfo(docId),
    enabled: !!docId,
    retry: 1, // Only retry once for verification info
    staleTime: 0, // Always fresh for verification
    gcTime: 5 * 60 * 1000, // Keep in cache for 5 minutes
    select: (data) => ({
      ...data,
      createdAt: new Date(data.createdAt).toISOString(),
    }),
  });
}

export function useVerifyDocument() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ docId, file }: { docId: string; file: File }) => 
      verifyDocument(docId, file),
    onSuccess: (data, variables) => {
      // Update verification info cache with the result
      queryClient.setQueryData(['verification', variables.docId], data);
      
      // Log verification attempt for analytics
      if (process.env.NODE_ENV !== 'production') {
        console.log('Document verification completed:', {
          docId: variables.docId,
          status: data.status,
          filename: data.filename,
        });
      }
    },
    onError: (error, variables) => {
      console.error('Document verification failed:', {
        docId: variables.docId,
        error,
      });
    },
  });
}