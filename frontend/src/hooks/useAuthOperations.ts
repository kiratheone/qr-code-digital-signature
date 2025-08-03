/**
 * Authentication Operations Hook
 * Manages authentication state and operations with React Query
 */

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AuthService } from '@/lib/services';
import { apiClient } from '@/lib/api';
import type { User, LoginResponse, RegisterResponse } from '@/lib/types';

// Create service instance
const authService = new AuthService(apiClient);

// Query keys for React Query
export const authKeys = {
  all: ['auth'] as const,
  user: () => [...authKeys.all, 'user'] as const,
  session: () => [...authKeys.all, 'session'] as const,
};

/**
 * Hook for authentication operations including login, logout, and session management
 */
export function useAuthOperations() {
  const queryClient = useQueryClient();

  // Query for current user session
  const sessionQuery = useQuery({
    queryKey: authKeys.session(),
    queryFn: () => {
      const authState = authService.getAuthState();
      if (!authState.isAuthenticated) {
        throw new Error('Not authenticated');
      }
      return authState;
    },
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // Query for validating session with server
  const validateSessionQuery = useQuery({
    queryKey: authKeys.user(),
    queryFn: () => authService.validateSession(),
    enabled: sessionQuery.data?.isAuthenticated || false,
    retry: (failureCount, error: any) => {
      // Don't retry on 401 errors
      if (error?.status === 401) {
        return false;
      }
      return failureCount < 2;
    },
    staleTime: 10 * 60 * 1000, // 10 minutes
  });

  // Mutation for login
  const loginMutation = useMutation({
    mutationFn: ({ username, password }: { username: string; password: string }) =>
      authService.login(username, password),
    onSuccess: (data: LoginResponse) => {
      // Update session cache
      queryClient.setQueryData(authKeys.session(), {
        user: data.user,
        token: data.token,
        isAuthenticated: true,
        isLoading: false,
      });
      
      // Update user cache
      queryClient.setQueryData(authKeys.user(), data.user);
      
      // Invalidate all queries to refetch with new auth
      queryClient.invalidateQueries();
    },
    onError: (error) => {
      console.error('Login failed:', error);
    },
  });

  // Mutation for registration
  const registerMutation = useMutation({
    mutationFn: ({
      username,
      password,
      fullName,
      email,
    }: {
      username: string;
      password: string;
      fullName: string;
      email: string;
    }) => authService.register(username, password, fullName, email),
    onSuccess: (data: RegisterResponse) => {
      // Update session cache
      queryClient.setQueryData(authKeys.session(), {
        user: data.user,
        token: data.token,
        isAuthenticated: true,
        isLoading: false,
      });
      
      // Update user cache
      queryClient.setQueryData(authKeys.user(), data.user);
      
      // Invalidate all queries to refetch with new auth
      queryClient.invalidateQueries();
    },
    onError: (error) => {
      console.error('Registration failed:', error);
    },
  });

  // Mutation for logout
  const logoutMutation = useMutation({
    mutationFn: () => authService.logout(),
    onSuccess: () => {
      // Clear all auth-related cache
      queryClient.removeQueries({ queryKey: authKeys.all });
      
      // Clear all other cached data
      queryClient.clear();
    },
    onError: (error) => {
      console.error('Logout failed:', error);
      // Still clear cache even if server logout fails
      queryClient.removeQueries({ queryKey: authKeys.all });
      queryClient.clear();
    },
  });

  // Get current authentication state
  const authState = authService.getAuthState();
  const currentUser = validateSessionQuery.data || authState.user;

  return {
    // Authentication state
    user: currentUser,
    isAuthenticated: authState.isAuthenticated && !validateSessionQuery.isError,
    token: authState.token,
    
    // Loading states
    isLoading: sessionQuery.isLoading || validateSessionQuery.isLoading,
    isLoggingIn: loginMutation.isPending,
    isRegistering: registerMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    
    // Error states
    sessionError: sessionQuery.error || validateSessionQuery.error,
    loginError: loginMutation.error,
    registerError: registerMutation.error,
    logoutError: logoutMutation.error,
    
    // Actions
    login: loginMutation.mutate,
    register: registerMutation.mutate,
    logout: logoutMutation.mutate,
    
    // Success states
    loginSuccess: loginMutation.isSuccess,
    registerSuccess: registerMutation.isSuccess,
    logoutSuccess: logoutMutation.isSuccess,
    
    // Reset mutations
    resetLogin: loginMutation.reset,
    resetRegister: registerMutation.reset,
    resetLogout: logoutMutation.reset,
    
    // Utilities
    refetchSession: sessionQuery.refetch,
    validateSession: validateSessionQuery.refetch,
  };
}

/**
 * Hook for authentication validation utilities
 */
export function useAuthValidation() {
  const validatePassword = (password: string) => {
    return authService.validatePassword(password);
  };

  const validateUsername = (username: string) => {
    return authService.validateUsername(username);
  };

  const isValidEmail = (email: string) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  return {
    validatePassword,
    validateUsername,
    isValidEmail,
  };
}

/**
 * Hook for checking if user is authenticated (for route protection)
 */
export function useRequireAuth() {
  const { isAuthenticated, isLoading, user } = useAuthOperations();

  return {
    isAuthenticated,
    isLoading,
    user,
    // Helper to redirect to login if not authenticated
    requireAuth: () => {
      if (!isLoading && !isAuthenticated) {
        // This would typically trigger a redirect to login page
        // Implementation depends on your routing solution
        return false;
      }
      return true;
    },
  };
}