'use client';

import React, { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useNotificationHelpers } from '@/components/UI/Notifications';
import { LoadingSpinner } from '@/components/UI/LoadingSpinner';
import { NetworkManager, createErrorRecoveryHandler } from '@/utils/apiUtils';

interface ErrorRecoveryProps {
  error?: Error | null;
  onRetry?: () => void;
  showNetworkStatus?: boolean;
  className?: string;
}

export function ErrorRecovery({ 
  error, 
  onRetry, 
  showNetworkStatus = true,
  className = '' 
}: ErrorRecoveryProps) {
  const [isRetrying, setIsRetrying] = useState(false);
  const [isOnline, setIsOnline] = useState(true);
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useNotificationHelpers();

  useEffect(() => {
    if (!showNetworkStatus) return;

    const networkManager = NetworkManager.getInstance();
    setIsOnline(networkManager.getStatus());

    const unsubscribe = networkManager.onStatusChange((online) => {
      setIsOnline(online);
      
      if (online) {
        showSuccess('Connection Restored', 'You are back online!');
        // Retry failed queries when coming back online
        const errorRecovery = createErrorRecoveryHandler(queryClient);
        errorRecovery.retryFailedQueries();
      } else {
        showError('Connection Lost', 'You are currently offline. Some features may not work.', {
          persistent: true,
        });
      }
    });

    return unsubscribe;
  }, [showNetworkStatus, queryClient, showSuccess, showError]);

  const handleRetry = async () => {
    if (isRetrying) return;

    setIsRetrying(true);
    
    try {
      if (onRetry) {
        await onRetry();
      } else {
        // Default retry behavior
        const errorRecovery = createErrorRecoveryHandler(queryClient);
        errorRecovery.retryFailedQueries();
      }
      
      showSuccess('Retry Successful', 'The operation completed successfully.');
    } catch (retryError) {
      showError('Retry Failed', 'The retry attempt was unsuccessful. Please try again.');
      console.error('Retry failed:', retryError);
    } finally {
      setIsRetrying(false);
    }
  };

  const handleRefresh = () => {
    window.location.reload();
  };

  const handleGoBack = () => {
    if (window.history.length > 1) {
      window.history.back();
    } else {
      window.location.href = '/';
    }
  };

  const handleClearCache = () => {
    queryClient.clear();
    localStorage.clear();
    sessionStorage.clear();
    showSuccess('Cache Cleared', 'Application cache has been cleared.');
  };

  return (
    <div className={`bg-white rounded-lg shadow-md p-6 ${className}`}>
      <div className="text-center">
        <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100 mb-4">
          <svg
            className="h-6 w-6 text-red-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
            />
          </svg>
        </div>
        
        <h3 className="text-lg font-medium text-gray-900 mb-2">
          Something went wrong
        </h3>
        
        {error && (
          <p className="text-sm text-gray-600 mb-4">
            {error.message || 'An unexpected error occurred'}
          </p>
        )}

        {showNetworkStatus && (
          <div className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium mb-4 ${
            isOnline 
              ? 'bg-green-100 text-green-800' 
              : 'bg-red-100 text-red-800'
          }`}>
            <div className={`w-2 h-2 rounded-full mr-2 ${
              isOnline ? 'bg-green-400' : 'bg-red-400'
            }`} />
            {isOnline ? 'Online' : 'Offline'}
          </div>
        )}

        <div className="space-y-3">
          <button
            onClick={handleRetry}
            disabled={isRetrying || (!isOnline && showNetworkStatus)}
            className="w-full flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors duration-200"
          >
            {isRetrying ? (
              <>
                <LoadingSpinner size="sm" color="white" className="mr-2" />
                Retrying...
              </>
            ) : (
              'Try Again'
            )}
          </button>

          <div className="grid grid-cols-2 gap-3">
            <button
              onClick={handleRefresh}
              className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
            >
              Refresh Page
            </button>

            <button
              onClick={handleGoBack}
              className="px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
            >
              Go Back
            </button>
          </div>

          <details className="text-left">
            <summary className="cursor-pointer text-sm text-gray-500 hover:text-gray-700 transition-colors duration-200">
              Advanced Options
            </summary>
            <div className="mt-3 space-y-2">
              <button
                onClick={handleClearCache}
                className="w-full px-4 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-50 rounded-md transition-colors duration-200"
              >
                Clear Cache & Reload
              </button>
              
              <a
                href={`mailto:support@example.com?subject=Error Report&body=Error: ${encodeURIComponent(error?.message || 'Unknown error')}`}
                className="block w-full px-4 py-2 text-sm text-gray-600 hover:text-gray-800 hover:bg-gray-50 rounded-md transition-colors duration-200 text-center"
              >
                Report Issue
              </a>
            </div>
          </details>
        </div>
      </div>
    </div>
  );
}

// Hook for using error recovery in components
export function useErrorRecovery() {
  const queryClient = useQueryClient();
  const { showSuccess, showError } = useNotificationHelpers();

  const retryFailedOperations = async () => {
    try {
      const errorRecovery = createErrorRecoveryHandler(queryClient);
      errorRecovery.retryFailedQueries();
      showSuccess('Retry Initiated', 'Attempting to retry failed operations...');
    } catch (error) {
      showError('Retry Failed', 'Unable to retry operations. Please try again manually.');
    }
  };

  const clearErrorState = () => {
    const errorRecovery = createErrorRecoveryHandler(queryClient);
    errorRecovery.clearErrors();
    showSuccess('Errors Cleared', 'Error states have been cleared.');
  };

  return {
    retryFailedOperations,
    clearErrorState,
  };
}