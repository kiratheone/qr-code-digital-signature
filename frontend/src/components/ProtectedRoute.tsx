/**
 * Protected Route Component
 * Handles authentication-based route protection with loading states
 */

'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useRequireAuth } from '@/hooks';
import { PageLoading } from '@/components/ui/LoadingSpinner';

interface ProtectedRouteProps {
  children: React.ReactNode;
  redirectTo?: string;
  fallback?: React.ReactNode;
}

export function ProtectedRoute({ 
  children, 
  redirectTo = '/login',
  fallback 
}: ProtectedRouteProps) {
  const router = useRouter();
  const { isAuthenticated, isLoading, user } = useRequireAuth();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push(redirectTo);
    }
  }, [isAuthenticated, isLoading, router, redirectTo]);

  // Show loading state while checking authentication
  if (isLoading) {
    return fallback || <PageLoading text="Checking authentication..." />;
  }

  // Don't render children if not authenticated
  if (!isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}

/**
 * Higher-order component for protecting pages
 */
export function withProtectedRoute<P extends object>(
  Component: React.ComponentType<P>,
  options?: {
    redirectTo?: string;
    fallback?: React.ReactNode;
  }
) {
  const ProtectedComponent = (props: P) => {
    return (
      <ProtectedRoute 
        redirectTo={options?.redirectTo}
        fallback={options?.fallback}
      >
        <Component {...props} />
      </ProtectedRoute>
    );
  };

  ProtectedComponent.displayName = `withProtectedRoute(${Component.displayName || Component.name})`;
  
  return ProtectedComponent;
}