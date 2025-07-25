'use client';

import React from 'react';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  color?: 'primary' | 'secondary' | 'white' | 'gray';
  className?: string;
}

const sizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
  xl: 'h-12 w-12',
};

const colorClasses = {
  primary: 'text-indigo-600',
  secondary: 'text-gray-600',
  white: 'text-white',
  gray: 'text-gray-400',
};

export function LoadingSpinner({ 
  size = 'md', 
  color = 'primary', 
  className = '' 
}: LoadingSpinnerProps) {
  return (
    <svg
      className={`animate-spin ${sizeClasses[size]} ${colorClasses[color]} ${className}`}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
    >
      <circle
        className="opacity-25"
        cx="12"
        cy="12"
        r="10"
        stroke="currentColor"
        strokeWidth="4"
      />
      <path
        className="opacity-75"
        fill="currentColor"
        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  );
}

interface LoadingOverlayProps {
  isLoading: boolean;
  message?: string;
  children: React.ReactNode;
  className?: string;
}

export function LoadingOverlay({ 
  isLoading, 
  message = 'Loading...', 
  children, 
  className = '' 
}: LoadingOverlayProps) {
  return (
    <div className={`relative ${className}`}>
      {children}
      {isLoading && (
        <div className="absolute inset-0 bg-white bg-opacity-75 flex items-center justify-center z-50">
          <div className="flex flex-col items-center space-y-2">
            <LoadingSpinner size="lg" />
            <p className="text-sm text-gray-600">{message}</p>
          </div>
        </div>
      )}
    </div>
  );
}

interface LoadingButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  isLoading: boolean;
  loadingText?: string;
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md' | 'lg';
}

export function LoadingButton({ 
  isLoading, 
  loadingText = 'Loading...', 
  children, 
  disabled,
  className = '',
  variant = 'primary',
  size = 'md',
  ...props 
}: LoadingButtonProps) {
  const baseClasses = 'relative inline-flex items-center justify-center font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-200';
  
  const variantClasses = {
    primary: 'bg-indigo-600 text-white hover:bg-indigo-700 focus:ring-indigo-500 disabled:bg-indigo-400',
    secondary: 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50 focus:ring-indigo-500 disabled:bg-gray-100',
    danger: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500 disabled:bg-red-400',
  };
  
  const sizeClasses = {
    sm: 'px-3 py-2 text-sm',
    md: 'px-4 py-2 text-sm',
    lg: 'px-6 py-3 text-base',
  };

  return (
    <button
      {...props}
      disabled={disabled || isLoading}
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className} ${
        isLoading || disabled ? 'cursor-not-allowed opacity-75' : ''
      }`}
    >
      {isLoading && (
        <LoadingSpinner 
          size="sm" 
          color={variant === 'secondary' ? 'gray' : 'white'} 
          className="mr-2" 
        />
      )}
      <span className={isLoading ? 'ml-2' : ''}>
        {isLoading ? loadingText : children}
      </span>
    </button>
  );
}

interface ProgressBarProps {
  progress: number; // 0-100
  size?: 'sm' | 'md' | 'lg';
  color?: 'primary' | 'success' | 'warning' | 'error';
  showPercentage?: boolean;
  className?: string;
}

const progressSizeClasses = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-3',
};

const progressColorClasses = {
  primary: 'bg-indigo-600',
  success: 'bg-green-600',
  warning: 'bg-yellow-600',
  error: 'bg-red-600',
};

export function ProgressBar({ 
  progress, 
  size = 'md', 
  color = 'primary', 
  showPercentage = false,
  className = '' 
}: ProgressBarProps) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div className={`w-full ${className}`}>
      <div className={`w-full bg-gray-200 rounded-full ${progressSizeClasses[size]}`}>
        <div
          className={`${progressColorClasses[color]} ${progressSizeClasses[size]} rounded-full transition-all duration-300 ease-out`}
          style={{ width: `${clampedProgress}%` }}
        />
      </div>
      {showPercentage && (
        <div className="flex justify-between text-xs text-gray-600 mt-1">
          <span>0%</span>
          <span>{Math.round(clampedProgress)}%</span>
          <span>100%</span>
        </div>
      )}
    </div>
  );
}

interface LoadingSkeletonProps {
  lines?: number;
  className?: string;
}

export function LoadingSkeleton({ lines = 3, className = '' }: LoadingSkeletonProps) {
  return (
    <div className={`animate-pulse ${className}`}>
      {Array.from({ length: lines }).map((_, index) => (
        <div
          key={index}
          className={`bg-gray-200 rounded h-4 mb-2 ${
            index === lines - 1 ? 'w-3/4' : 'w-full'
          }`}
        />
      ))}
    </div>
  );
}

interface LoadingCardProps {
  className?: string;
}

export function LoadingCard({ className = '' }: LoadingCardProps) {
  return (
    <div className={`animate-pulse bg-white p-6 rounded-lg shadow ${className}`}>
      <div className="flex items-center space-x-4">
        <div className="rounded-full bg-gray-200 h-10 w-10" />
        <div className="flex-1 space-y-2">
          <div className="h-4 bg-gray-200 rounded w-3/4" />
          <div className="h-4 bg-gray-200 rounded w-1/2" />
        </div>
      </div>
      <div className="mt-4 space-y-2">
        <div className="h-4 bg-gray-200 rounded" />
        <div className="h-4 bg-gray-200 rounded" />
        <div className="h-4 bg-gray-200 rounded w-5/6" />
      </div>
    </div>
  );
}