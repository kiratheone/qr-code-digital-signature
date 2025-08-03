/**
 * Register Page
 * User registration page with form validation
 * Uses hooks and components with clean separation
 */

'use client';

import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { useAuthOperations, useAuthValidation } from '@/hooks';

export default function RegisterPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    confirmPassword: '',
    fullName: '',
    email: '',
  });
  const [errors, setErrors] = useState({
    username: '',
    password: '',
    confirmPassword: '',
    fullName: '',
    email: '',
  });

  // Use custom hooks for authentication logic
  const {
    isAuthenticated,
    isRegistering,
    registerError,
    registerSuccess,
    register,
    resetRegister,
  } = useAuthOperations();

  const { validateUsername, validatePassword, isValidEmail } = useAuthValidation();

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      router.push('/documents');
    }
  }, [isAuthenticated, router]);

  // Redirect on successful registration
  useEffect(() => {
    if (registerSuccess) {
      router.push('/documents');
    }
  }, [registerSuccess, router]);

  // Form validation
  const validateForm = (): boolean => {
    let isValid = true;
    const newErrors = {
      username: '',
      password: '',
      confirmPassword: '',
      fullName: '',
      email: '',
    };

    // Validate username
    const usernameValidation = validateUsername(formData.username);
    if (!usernameValidation.isValid) {
      newErrors.username = usernameValidation.error || 'Invalid username';
      isValid = false;
    }

    // Validate password
    const passwordValidation = validatePassword(formData.password);
    if (!passwordValidation.isValid) {
      newErrors.password = passwordValidation.errors[0] || 'Invalid password';
      isValid = false;
    }

    // Validate confirm password
    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
      isValid = false;
    }

    // Validate full name
    if (!formData.fullName.trim()) {
      newErrors.fullName = 'Full name is required';
      isValid = false;
    } else if (formData.fullName.trim().length < 2) {
      newErrors.fullName = 'Full name must be at least 2 characters';
      isValid = false;
    }

    // Validate email
    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
      isValid = false;
    } else if (!isValidEmail(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
      isValid = false;
    }

    setErrors(newErrors);
    return isValid;
  };

  // Handle form submission
  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    register({
      username: formData.username.trim(),
      password: formData.password,
      fullName: formData.fullName.trim(),
      email: formData.email.trim(),
    });
  };

  // Handle input changes with error clearing
  const handleInputChange = (field: keyof typeof formData) => (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const value = event.target.value;
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error when user starts typing
    if (errors[field] && value.trim()) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  // Handle error dismissal
  const handleErrorDismiss = () => {
    resetRegister();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Create your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Join the digital signature system
          </p>
        </div>

        {/* Registration Error */}
        {registerError && (
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Registration failed
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{registerError instanceof Error ? registerError.message : 'Registration failed. Please try again.'}</p>
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
              label="Full Name"
              type="text"
              value={formData.fullName}
              onChange={handleInputChange('fullName')}
              placeholder="Enter your full name"
              error={errors.fullName}
              disabled={isRegistering}
              required
              autoComplete="name"
            />

            <Input
              label="Username"
              type="text"
              value={formData.username}
              onChange={handleInputChange('username')}
              placeholder="Choose a username"
              error={errors.username}
              disabled={isRegistering}
              required
              autoComplete="username"
            />

            <Input
              label="Email"
              type="email"
              value={formData.email}
              onChange={handleInputChange('email')}
              placeholder="Enter your email address"
              error={errors.email}
              disabled={isRegistering}
              required
              autoComplete="email"
            />

            <Input
              label="Password"
              type="password"
              value={formData.password}
              onChange={handleInputChange('password')}
              placeholder="Create a password"
              error={errors.password}
              disabled={isRegistering}
              required
              autoComplete="new-password"
            />

            <Input
              label="Confirm Password"
              type="password"
              value={formData.confirmPassword}
              onChange={handleInputChange('confirmPassword')}
              placeholder="Confirm your password"
              error={errors.confirmPassword}
              disabled={isRegistering}
              required
              autoComplete="new-password"
            />
          </div>

          <div>
            <Button
              type="submit"
              variant="primary"
              size="lg"
              isLoading={isRegistering}
              disabled={isRegistering}
              className="w-full"
            >
              {isRegistering ? 'Creating account...' : 'Create Account'}
            </Button>
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600">
              Already have an account?{' '}
              <button
                type="button"
                onClick={() => router.push('/login')}
                className="font-medium text-blue-600 hover:text-blue-500"
                disabled={isRegistering}
              >
                Sign in here
              </button>
            </p>
          </div>
        </form>

        <div className="mt-6">
          <div className="text-center">
            <button
              onClick={() => router.push('/')}
              className="text-sm text-gray-500 hover:text-gray-700"
              disabled={isRegistering}
            >
              ← Back to Home
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}