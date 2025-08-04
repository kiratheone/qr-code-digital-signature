/**
 * Form Validation Hook
 * Provides simple form validation with error handling
 */

import { useState, useCallback } from 'react';

export interface ValidationRule<T = any> {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: T) => string | null;
}

export interface ValidationRules {
  [key: string]: ValidationRule;
}

export interface ValidationErrors {
  [key: string]: string;
}

export interface FormValidationResult {
  isValid: boolean;
  errors: ValidationErrors;
}

export function useFormValidation<T extends Record<string, any>>(
  initialValues: T,
  rules: ValidationRules
) {
  const [values, setValues] = useState<T>(initialValues);
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  // Validate a single field
  const validateField = useCallback((name: string, value: any): string | null => {
    const rule = rules[name];
    if (!rule) return null;

    // Required validation
    if (rule.required && (!value || (typeof value === 'string' && !value.trim()))) {
      return `${formatFieldName(name)} is required`;
    }

    // Skip other validations if field is empty and not required
    if (!value || (typeof value === 'string' && !value.trim())) {
      return null;
    }

    // Custom validation (run first for better error messages)
    if (rule.custom) {
      const customError = rule.custom(value);
      if (customError) {
        return customError;
      }
    }

    // String validations
    if (typeof value === 'string') {
      // Min length validation
      if (rule.minLength && value.length < rule.minLength) {
        return `${formatFieldName(name)} must be at least ${rule.minLength} characters`;
      }

      // Max length validation
      if (rule.maxLength && value.length > rule.maxLength) {
        return `${formatFieldName(name)} must be less than ${rule.maxLength} characters`;
      }

      // Pattern validation
      if (rule.pattern && !rule.pattern.test(value)) {
        return `${formatFieldName(name)} format is invalid`;
      }
    }

    return null;
  }, [rules]);

  // Validate all fields
  const validateAll = useCallback((): FormValidationResult => {
    const newErrors: ValidationErrors = {};
    let isValid = true;

    Object.keys(rules).forEach((fieldName) => {
      const error = validateField(fieldName, values[fieldName]);
      if (error) {
        newErrors[fieldName] = error;
        isValid = false;
      }
    });

    setErrors(newErrors);
    return { isValid, errors: newErrors };
  }, [values, validateField, rules]);

  // Update field value
  const setValue = useCallback((name: string, value: any) => {
    setValues(prev => ({ ...prev, [name]: value }));
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  }, [errors]);

  // Update multiple values
  const setMultipleValues = useCallback((newValues: Partial<T>) => {
    setValues(prev => ({ ...prev, ...newValues }));
  }, []);

  // Mark field as touched
  const setFieldTouched = useCallback((name: string, isTouched = true) => {
    setTouched(prev => ({ ...prev, [name]: isTouched }));
    
    // Validate field when it's touched
    if (isTouched) {
      const error = validateField(name, values[name]);
      if (error) {
        setErrors(prev => ({ ...prev, [name]: error }));
      }
    }
  }, [validateField, values]);

  // Handle field blur
  const handleBlur = useCallback((name: string) => {
    setFieldTouched(name, true);
  }, [setFieldTouched]);

  // Handle field change
  const handleChange = useCallback((name: string, value: any) => {
    setValue(name, value);
    setFieldTouched(name, true);
  }, [setValue, setFieldTouched]);

  // Reset form
  const reset = useCallback(() => {
    setValues(initialValues);
    setErrors({});
    setTouched({});
  }, [initialValues]);

  // Clear errors
  const clearErrors = useCallback(() => {
    setErrors({});
  }, []);

  // Get field error (only show if touched)
  const getFieldError = useCallback((name: string): string | null => {
    return touched[name] ? errors[name] || null : null;
  }, [errors, touched]);

  // Check if field has error
  const hasFieldError = useCallback((name: string): boolean => {
    return !!touched[name] && !!errors[name];
  }, [errors, touched]);

  // Check if form is valid
  const isFormValid = Object.keys(errors).length === 0;

  return {
    values,
    errors,
    touched,
    isFormValid,
    setValue,
    setMultipleValues,
    setFieldTouched,
    handleBlur,
    handleChange,
    validateField,
    validateAll,
    reset,
    clearErrors,
    getFieldError,
    hasFieldError,
  };
}

// Helper function to format field names for error messages
function formatFieldName(name: string): string {
  return name
    .replace(/([A-Z])/g, ' $1')
    .replace(/^./, str => str.toUpperCase())
    .trim();
}

// Common validation rules
export const commonRules = {
  email: {
    pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    custom: (value: string) => {
      if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
        return 'Please enter a valid email address';
      }
      return null;
    },
  },
  password: {
    minLength: 8,
    custom: (value: string) => {
      if (!value) return null;
      
      const errors: string[] = [];
      
      if (value.length < 8) {
        errors.push('at least 8 characters');
      }
      if (!/[A-Z]/.test(value)) {
        errors.push('one uppercase letter');
      }
      if (!/[a-z]/.test(value)) {
        errors.push('one lowercase letter');
      }
      if (!/\d/.test(value)) {
        errors.push('one number');
      }
      
      if (errors.length > 0) {
        return `Password must contain ${errors.join(', ')}`;
      }
      
      return null;
    },
  },
  username: {
    minLength: 3,
    maxLength: 50,
    pattern: /^[a-zA-Z0-9_-]+$/,
    custom: (value: string) => {
      if (value && !/^[a-zA-Z0-9_-]+$/.test(value)) {
        return 'Username can only contain letters, numbers, hyphens, and underscores';
      }
      return null;
    },
  },
  required: {
    required: true,
  },
};