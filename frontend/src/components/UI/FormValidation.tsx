'use client';

import React, { useState, useCallback } from 'react';

// Validation rule types
export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: unknown) => string | undefined;
  email?: boolean;
  url?: boolean;
  number?: boolean;
  min?: number;
  max?: number;
}

export interface FieldValidation {
  value: unknown;
  rules: ValidationRule;
  error?: string;
  touched?: boolean;
}

export interface FormValidation {
  [fieldName: string]: FieldValidation;
}

// Validation functions
export function validateField(value: unknown, rules: ValidationRule): string | undefined {
  // Required validation
  if (rules.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
    return 'This field is required';
  }

  // Skip other validations if field is empty and not required
  if (!value || (typeof value === 'string' && value.trim() === '')) {
    return undefined;
  }

  // String validations
  if (typeof value === 'string') {
    if (rules.minLength && value.length < rules.minLength) {
      return `Must be at least ${rules.minLength} characters`;
    }

    if (rules.maxLength && value.length > rules.maxLength) {
      return `Must be no more than ${rules.maxLength} characters`;
    }

    if (rules.email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(value)) {
        return 'Please enter a valid email address';
      }
    }

    if (rules.url) {
      try {
        new URL(value);
      } catch {
        return 'Please enter a valid URL';
      }
    }

    if (rules.pattern && !rules.pattern.test(value)) {
      return 'Please enter a valid format';
    }
  }

  // Number validations
  if (rules.number) {
    const numValue = Number(value);
    if (isNaN(numValue)) {
      return 'Please enter a valid number';
    }

    if (rules.min !== undefined && numValue < rules.min) {
      return `Must be at least ${rules.min}`;
    }

    if (rules.max !== undefined && numValue > rules.max) {
      return `Must be no more than ${rules.max}`;
    }
  }

  // Custom validation
  if (rules.custom) {
    return rules.custom(value);
  }

  return undefined;
}

// Hook for form validation with enhanced error handling
export function useFormValidation(initialFields: FormValidation) {
  const [fields, setFields] = useState<FormValidation>(initialFields);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitAttempts, setSubmitAttempts] = useState(0);
  const [lastSubmitTime, setLastSubmitTime] = useState<Date | null>(null);
  const [validationMode, setValidationMode] = useState<'onBlur' | 'onChange' | 'onSubmit'>('onBlur');

  const validateSingleField = useCallback((fieldName: string, value: unknown) => {
    const field = fields[fieldName];
    if (!field) return undefined;

    const error = validateField(value, field.rules);
    
    setFields(prev => ({
      ...prev,
      [fieldName]: {
        ...prev[fieldName],
        value,
        error,
        touched: true,
      },
    }));

    // Clear submit error when user starts fixing validation errors
    if (submitError && !error) {
      setSubmitError(null);
    }

    return error;
  }, [fields, submitError]);

  const validateAllFields = useCallback(() => {
    const errors: { [key: string]: string } = {};
    let isValid = true;

    const updatedFields = { ...fields };

    Object.keys(fields).forEach(fieldName => {
      const field = fields[fieldName];
      const error = validateField(field.value, field.rules);
      
      updatedFields[fieldName] = {
        ...field,
        error,
        touched: true,
      };

      if (error) {
        errors[fieldName] = error;
        isValid = false;
      }
    });

    setFields(updatedFields);
    return { isValid, errors };
  }, [fields]);

  const updateField = useCallback((fieldName: string, value: unknown) => {
    setFields(prev => ({
      ...prev,
      [fieldName]: {
        ...prev[fieldName],
        value,
      },
    }));

    // Validate on change if in onChange mode or field was previously invalid
    if (validationMode === 'onChange' || fields[fieldName]?.error) {
      setTimeout(() => {
        validateSingleField(fieldName, value);
      }, 300); // Debounce validation
    }
  }, [fields, validateSingleField, validationMode]);

  const touchField = useCallback((fieldName: string) => {
    setFields(prev => ({
      ...prev,
      [fieldName]: {
        ...prev[fieldName],
        touched: true,
      },
    }));
  }, []);

  const resetForm = useCallback(() => {
    const resetFields = { ...initialFields };
    Object.keys(resetFields).forEach(key => {
      resetFields[key] = {
        ...resetFields[key],
        error: undefined,
        touched: false,
      };
    });
    setFields(resetFields);
    setSubmitError(null);
    setSubmitAttempts(0);
    setIsSubmitting(false);
    setLastSubmitTime(null);
    setValidationMode('onBlur');
  }, [initialFields]);

  const handleSubmit = useCallback(async (
    onSubmit: (data: Record<string, unknown>) => Promise<void>,
    options?: { 
      showSuccessMessage?: boolean;
      successMessage?: string;
      onSuccess?: () => void;
      onError?: (error: Error) => void;
      preventDuplicateSubmission?: boolean;
    }
  ) => {
    if (isSubmitting) return;

    // Prevent duplicate submissions within 2 seconds
    if (options?.preventDuplicateSubmission !== false && lastSubmitTime) {
      const timeSinceLastSubmit = Date.now() - lastSubmitTime.getTime();
      if (timeSinceLastSubmit < 2000) {
        setSubmitError('Please wait before submitting again.');
        return;
      }
    }

    setIsSubmitting(true);
    setSubmitAttempts(prev => prev + 1);
    setSubmitError(null);
    setLastSubmitTime(new Date());

    try {
      // Validate all fields first
      const { isValid, errors } = validateAllFields();
      
      if (!isValid) {
        const errorCount = Object.keys(errors).length;
        setSubmitError(`Please fix ${errorCount} error${errorCount > 1 ? 's' : ''} before submitting.`);
        
        // Switch to onChange validation after first failed submit
        if (submitAttempts === 0) {
          setValidationMode('onChange');
        }
        return;
      }

      // Prepare form data
      const formData = Object.keys(fields).reduce((acc, key) => {
        acc[key] = fields[key].value;
        return acc;
      }, {} as Record<string, unknown>);

      // Submit form
      await onSubmit(formData);
      
      // Reset form on success
      if (options?.showSuccessMessage !== false) {
        setSubmitAttempts(0);
        setValidationMode('onBlur'); // Reset validation mode
      }
      
      options?.onSuccess?.();
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'An unexpected error occurred';
      setSubmitError(errorMessage);
      options?.onError?.(error instanceof Error ? error : new Error(errorMessage));
    } finally {
      setIsSubmitting(false);
    }
  }, [fields, isSubmitting, validateAllFields, submitAttempts, lastSubmitTime]);

  const isFormValid = Object.values(fields).every(field => !field.error);
  const hasErrors = Object.values(fields).some(field => field.error && field.touched);
  const canSubmit = isFormValid && !isSubmitting;

  return {
    fields,
    validateSingleField,
    validateAllFields,
    updateField,
    touchField,
    resetForm,
    handleSubmit,
    isFormValid,
    hasErrors,
    isSubmitting,
    submitError,
    submitAttempts,
    canSubmit,
    validationMode,
    setValidationMode,
    lastSubmitTime,
  };
}

// Form field components with validation
interface ValidatedInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  touched?: boolean;
  helpText?: string;
}

export function ValidatedInput({ 
  label, 
  error, 
  touched, 
  helpText, 
  className = '', 
  id,
  ...props 
}: ValidatedInputProps) {
  const hasError = error && touched;
  const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;

  return (
    <div className="mb-4">
      <label htmlFor={inputId} className="block text-sm font-medium text-gray-700 mb-1">
        {label}
        {props.required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <input
        {...props}
        id={inputId}
        className={`
          block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-1 sm:text-sm
          ${hasError 
            ? 'border-red-300 focus:ring-red-500 focus:border-red-500' 
            : 'border-gray-300 focus:ring-indigo-500 focus:border-indigo-500'
          }
          ${className}
        `}
      />
      {hasError && (
        <p className="mt-1 text-sm text-red-600 flex items-center">
          <svg className="h-4 w-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}
      {helpText && !hasError && (
        <p className="mt-1 text-sm text-gray-500">{helpText}</p>
      )}
    </div>
  );
}

interface ValidatedTextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  error?: string;
  touched?: boolean;
  helpText?: string;
}

export function ValidatedTextarea({ 
  label, 
  error, 
  touched, 
  helpText, 
  className = '', 
  id,
  ...props 
}: ValidatedTextareaProps) {
  const hasError = error && touched;
  const textareaId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`;

  return (
    <div className="mb-4">
      <label htmlFor={textareaId} className="block text-sm font-medium text-gray-700 mb-1">
        {label}
        {props.required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <textarea
        {...props}
        id={textareaId}
        className={`
          block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-1 sm:text-sm
          ${hasError 
            ? 'border-red-300 focus:ring-red-500 focus:border-red-500' 
            : 'border-gray-300 focus:ring-indigo-500 focus:border-indigo-500'
          }
          ${className}
        `}
      />
      {hasError && (
        <p className="mt-1 text-sm text-red-600 flex items-center">
          <svg className="h-4 w-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}
      {helpText && !hasError && (
        <p className="mt-1 text-sm text-gray-500">{helpText}</p>
      )}
    </div>
  );
}

interface ValidatedSelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  error?: string;
  touched?: boolean;
  helpText?: string;
  options: { value: string; label: string }[];
}

export function ValidatedSelect({ 
  label, 
  error, 
  touched, 
  helpText, 
  options, 
  className = '', 
  id,
  ...props 
}: ValidatedSelectProps) {
  const hasError = error && touched;
  const selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;

  return (
    <div className="mb-4">
      <label htmlFor={selectId} className="block text-sm font-medium text-gray-700 mb-1">
        {label}
        {props.required && <span className="text-red-500 ml-1">*</span>}
      </label>
      <select
        {...props}
        id={selectId}
        className={`
          block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-1 sm:text-sm
          ${hasError 
            ? 'border-red-300 focus:ring-red-500 focus:border-red-500' 
            : 'border-gray-300 focus:ring-indigo-500 focus:border-indigo-500'
          }
          ${className}
        `}
      >
        <option value="">Select an option</option>
        {options.map(option => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {hasError && (
        <p className="mt-1 text-sm text-red-600 flex items-center">
          <svg className="h-4 w-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
          {error}
        </p>
      )}
      {helpText && !hasError && (
        <p className="mt-1 text-sm text-gray-500">{helpText}</p>
      )}
    </div>
  );
}