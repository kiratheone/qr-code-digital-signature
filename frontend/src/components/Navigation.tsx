/**
 * Navigation Component
 * Simple navigation bar with authentication-aware links
 */

'use client';

import React from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuthOperations } from '@/hooks';
import { Button } from '@/components/ui/Button';

export function Navigation() {
  const router = useRouter();
  const { isAuthenticated, user, logout, isLoggingOut } = useAuthOperations();

  const handleLogout = () => {
    logout();
  };

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center py-4">
          {/* Logo/Brand */}
          <div className="flex items-center">
            <Link href="/" className="text-xl font-semibold text-gray-900 hover:text-gray-700">
              Digital Signature System
            </Link>
          </div>

          {/* Navigation Links */}
          <div className="flex items-center space-x-4">
            {isAuthenticated ? (
              <>
                {/* Authenticated Navigation */}
                <Link
                  href="/documents"
                  className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Documents
                </Link>
                
                {/* User Info */}
                <div className="flex items-center space-x-3">
                  <span className="text-sm text-gray-600">
                    Welcome, {user?.full_name || user?.username}
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={handleLogout}
                    isLoading={isLoggingOut}
                    disabled={isLoggingOut}
                  >
                    {isLoggingOut ? 'Signing out...' : 'Sign Out'}
                  </Button>
                </div>
              </>
            ) : (
              <>
                {/* Unauthenticated Navigation */}
                <Link
                  href="/login"
                  className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
                >
                  Sign In
                </Link>
                <Link
                  href="/register"
                  className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                >
                  Register
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}