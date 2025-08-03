/**
 * Login Page
 * Authentication page for user login
 * Uses hooks and components with clean separation
 */

'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useAuthOperations, useAuthValidation } from '@/hooks';

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [usernameError, setUsernameError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  // Use custom hooks for authentication logic
  const {
    isAuthenticated,
    isLoggingIn,
    loginError,
    loginSuccess,
    login,
    resetLogin,
  } = useAuthOperations();

  const { validateUsername, validatePassword } = useAuthValidation();

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      router.push('/documents');
    }
  }, [isAuthenticated, router]);

  // Redirect on successful login
  useEffect(() => {
    if (loginSuccess) {
      router.push('/documents');
    }
  }, [loginSuccess, router]);

  // Form validation
  const validateForm = (): boolean => {
    let isValid = true;
    
    // Reset errors
    setUsernameError('');
    setPasswordError('');

    // Validate username
    const usernameValidation = validateUsername(username);
    if (!usernameValidation.isValid) {
      setUsernameError(usernameValidation.error || 'Invalid username');
      isValid = false;
    }

    // Validate password
    if (!password.trim()) {
      setPasswordError('Password is required');
      isValid = false;
    }

    return isValid;
  };

  // Handle form submission
  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    login({ username: username.trim(), password });
  };

  // Handle input changes with error clearing
  const handleUsernameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setUsername(value);
    if (usernameError && value.trim()) {
      setUsernameError('');
    }
  };

  const handlePasswordChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setPassword(value);
    if (passwordError && value.trim()) {
      setPasswordError('');
    }
  };

  // Handle error dismissal
  const handleErrorDismiss = () => {
    resetLogin();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Access the digital signature system
          </p>
        </div>

        {/* Login Error */}
        {loginError && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Login failed
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{loginError instanceof Error ? loginError.message : 'Invalid username or password'}</p>
                </div>
                <div className="mt-4">
                  <button
                    type="button"
                    onClick={handleErrorDismiss}
                    className="text-sm font-medium text-red-800 hover:text-red-600"
                  >
                    Dismiss
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <Input
              label="Username"
              type="text"
              value={username}
              onChange={handleUsernameChange}
              placeholder="Enter your username"
              error={usernameError}
              disabled={isLoggingIn}
              required
              autoComplete="username"
            />

            <Input
              label="Password"
              type="password"
              value={password}
              onChange={handlePasswordChange}
              placeholder="Enter your password"
              error={passwordError}
              disabled={isLoggingIn}
              required
              autoComplete="current-password"
            />
          </div>

          <div>
            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isLoggingIn}
              disabled={!username.trim() || !password.trim() || isLoggingIn}
              className="w-full"
            >
              {isLoggingIn ? 'Signing in...' : 'Sign in'}
            </Button>
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Don&apos;t have an account?{' '}
              <button
                type="button"
                onClick={() => router.push('/register')}
                className="font-medium text-blue-600 hover:text-blue-500"
                disabled={isLoggingIn}
              >
                Register here
              </button>
            </p>
          </div>
        </form>

        <div className="mt-6">
          <div className="text-center">
            <button
              onClick={() => router.push('/')}
              className="text-sm text-gray-500 hover:text-gray-700"
              disabled={isLoggingIn}
            >
              ← Back to Home
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}