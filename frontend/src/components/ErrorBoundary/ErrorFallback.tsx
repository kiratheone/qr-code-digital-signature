'use client';

import React from 'react';
import { ErrorFallbackProps } from './ErrorBoundary';

export function ErrorFallback({ error, resetError, hasError }: ErrorFallbackProps) {
  if (!hasError || !error) {
    return null;
  }

  const isDevelopment = process.env.NODE_ENV === 'development';

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <div className="mx-auto h-12 w-12 text-red-500">
            <svg
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
          </div>
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900">
            Something went wrong
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            We apologize for the inconvenience. An unexpected error has occurred.
          </p>
        </div>

        <div className="mt-8 space-y-4">
          <button
            onClick={resetError}
            className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
          >
            Try Again
          </button>

          <button
            onClick={() => window.location.reload()}
            className="group relative w-full flex justify-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
          >
            Reload Page
          </button>

          <button
            onClick={() => window.history.back()}
            className="group relative w-full flex justify-center py-2 px-4 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-colors duration-200"
          >
            Go Back
          </button>
        </div>

        {isDevelopment && (
          <details className="mt-8 p-4 bg-red-50 border border-red-200 rounded-md">
            <summary className="cursor-pointer text-sm font-medium text-red-800 hover:text-red-900">
              Error Details (Development Only)
            </summary>
            <div className="mt-4 space-y-2">
              <div>
                <h4 className="text-sm font-medium text-red-800">Error Message:</h4>
                <p className="text-sm text-red-700 font-mono bg-red-100 p-2 rounded mt-1">
                  {error.message}
                </p>
              </div>
              {error.stack && (
                <div>
                  <h4 className="text-sm font-medium text-red-800">Stack Trace:</h4>
                  <pre className="text-xs text-red-700 font-mono bg-red-100 p-2 rounded mt-1 overflow-auto max-h-40">
                    {error.stack}
                  </pre>
                </div>
              )}
            </div>
          </details>
        )}

        <div className="text-center">
          <p className="text-xs text-gray-500">
            If this problem persists, please contact support.
          </p>
        </div>
      </div>
    </div>
  );
}

// Specialized error fallbacks for different contexts

export function ApiErrorFallback({ error, resetError }: ErrorFallbackProps) {
  return (
    <div className="rounded-md bg-red-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg
            className="h-5 w-5 text-red-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-red-800">
            Failed to load data
          </h3>
          <div className="mt-2 text-sm text-red-700">
            <p>
              {error?.message || 'An error occurred while fetching data from the server.'}
            </p>
          </div>
          <div className="mt-4">
            <button
              type="button"
              onClick={resetError}
              className="bg-red-50 text-red-800 rounded-md text-sm font-medium hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600 px-3 py-2"
            >
              Try again
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export function FormErrorFallback({ error, resetError }: ErrorFallbackProps) {
  return (
    <div className="rounded-md bg-yellow-50 p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg
            className="h-5 w-5 text-yellow-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-yellow-800">
            Form Error
          </h3>
          <div className="mt-2 text-sm text-yellow-700">
            <p>
              {error?.message || 'An error occurred while processing the form.'}
            </p>
          </div>
          <div className="mt-4">
            <button
              type="button"
              onClick={resetError}
              className="bg-yellow-50 text-yellow-800 rounded-md text-sm font-medium hover:bg-yellow-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-yellow-50 focus:ring-yellow-600 px-3 py-2"
            >
              Reset form
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}