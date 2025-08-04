/**
 * useFormValidation Hook Tests
 */

import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { useFormValidation, commonRules } from '../useFormValidation';

describe('useFormValidation', () => {
  const initialValues = {
    username: '',
    email: '',
    password: '',
  };

  const rules = {
    username: { required: true, minLength: 3 },
    email: { required: true, ...commonRules.email },
    password: { required: true, ...commonRules.password },
  };

  it('initializes with provided values', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    expect(result.current.values).toEqual(initialValues);
    expect(result.current.errors).toEqual({});
    expect(result.current.touched).toEqual({});
    expect(result.current.isFormValid).toBe(true);
  });

  it('validates required fields', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      const validation = result.current.validateAll();
      expect(validation.isValid).toBe(false);
      expect(validation.errors.username).toBe('Username is required');
      expect(validation.errors.email).toBe('Email is required');
      expect(validation.errors.password).toBe('Password is required');
    });
  });

  it('validates minimum length', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setValue('username', 'ab');
      const error = result.current.validateField('username', 'ab');
      expect(error).toBe('Username must be at least 3 characters');
    });
  });

  it('validates email format', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setValue('email', 'invalid-email');
      const error = result.current.validateField('email', 'invalid-email');
      expect(error).toBe('Please enter a valid email address');
    });
  });

  it('validates password strength', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setValue('password', 'weak');
      const error = result.current.validateField('password', 'weak');
      expect(error).toBe('Password must contain at least 8 characters, one uppercase letter, one number');
    });
  });

  it('updates field values', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setValue('username', 'testuser');
    });
    
    expect(result.current.values.username).toBe('testuser');
  });

  it('handles field touch state', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setFieldTouched('username', true);
    });
    
    expect(result.current.touched.username).toBe(true);
  });

  it('shows field errors only when touched', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    // Field not touched, should not show error
    expect(result.current.getFieldError('username')).toBeNull();
    expect(result.current.hasFieldError('username')).toBe(false);
    
    act(() => {
      result.current.setFieldTouched('username', true);
    });
    
    // Field touched but empty, should show error
    expect(result.current.getFieldError('username')).toBe('Username is required');
    expect(result.current.hasFieldError('username')).toBe(true);
  });

  it('clears errors when field becomes valid', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    // First touch the field to show error
    act(() => {
      result.current.setFieldTouched('username', true);
    });
    
    expect(result.current.getFieldError('username')).toBe('Username is required');
    
    // Then set valid value
    act(() => {
      result.current.setValue('username', 'validuser');
    });
    
    expect(result.current.getFieldError('username')).toBeNull();
    expect(result.current.hasFieldError('username')).toBe(false);
  });

  it('handles change events', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.handleChange('username', 'newvalue');
    });
    
    expect(result.current.values.username).toBe('newvalue');
    expect(result.current.touched.username).toBe(true);
  });

  it('handles blur events', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.handleBlur('username');
    });
    
    expect(result.current.touched.username).toBe(true);
  });

  it('resets form to initial state', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.setValue('username', 'testuser');
      result.current.setFieldTouched('username', true);
      result.current.reset();
    });
    
    expect(result.current.values).toEqual(initialValues);
    expect(result.current.errors).toEqual({});
    expect(result.current.touched).toEqual({});
  });

  it('clears all errors', () => {
    const { result } = renderHook(() => useFormValidation(initialValues, rules));
    
    act(() => {
      result.current.validateAll();
      result.current.clearErrors();
    });
    
    expect(result.current.errors).toEqual({});
  });

  it('validates with custom validation rules', () => {
    const customRules = {
      customField: {
        custom: (value: string) => {
          if (value === 'forbidden') {
            return 'This value is not allowed';
          }
          return null;
        },
      },
    };

    const { result } = renderHook(() => 
      useFormValidation({ customField: '' }, customRules)
    );
    
    act(() => {
      const error = result.current.validateField('customField', 'forbidden');
      expect(error).toBe('This value is not allowed');
    });
  });

  it('validates maximum length', () => {
    const maxLengthRules = {
      shortField: { maxLength: 5 },
    };

    const { result } = renderHook(() => 
      useFormValidation({ shortField: '' }, maxLengthRules)
    );
    
    act(() => {
      const error = result.current.validateField('shortField', 'toolongvalue');
      expect(error).toBe('Short Field must be less than 5 characters');
    });
  });

  it('validates pattern matching', () => {
    const patternRules = {
      alphanumeric: { pattern: /^[a-zA-Z0-9]+$/ },
    };

    const { result } = renderHook(() => 
      useFormValidation({ alphanumeric: '' }, patternRules)
    );
    
    act(() => {
      const error = result.current.validateField('alphanumeric', 'invalid@value');
      expect(error).toBe('Alphanumeric format is invalid');
    });
  });

  it('skips validation for empty non-required fields', () => {
    const optionalRules = {
      optional: { minLength: 5 }, // Not required, but has minLength
    };

    const { result } = renderHook(() => 
      useFormValidation({ optional: '' }, optionalRules)
    );
    
    act(() => {
      const error = result.current.validateField('optional', '');
      expect(error).toBeNull();
    });
  });

  it('validates non-empty optional fields', () => {
    const optionalRules = {
      optional: { minLength: 5 }, // Not required, but has minLength
    };

    const { result } = renderHook(() => 
      useFormValidation({ optional: '' }, optionalRules)
    );
    
    act(() => {
      const error = result.current.validateField('optional', 'abc');
      expect(error).toBe('Optional must be at least 5 characters');
    });
  });
});

describe('commonRules', () => {
  it('validates email format correctly', () => {
    const validEmails = ['test@example.com', 'user.name@domain.co.uk', 'test+tag@example.org'];
    const invalidEmails = ['invalid', '@example.com', 'test@', 'test.example.com'];

    validEmails.forEach(email => {
      expect(commonRules.email.custom!(email)).toBeNull();
    });

    invalidEmails.forEach(email => {
      expect(commonRules.email.custom!(email)).toBe('Please enter a valid email address');
    });
  });

  it('validates password strength correctly', () => {
    const validPasswords = ['Password123', 'MySecure1Pass', 'Test123ABC'];
    const invalidPasswords = ['weak', 'password', 'PASSWORD', '12345678', 'NoNumber'];

    validPasswords.forEach(password => {
      expect(commonRules.password.custom!(password)).toBeNull();
    });

    invalidPasswords.forEach(password => {
      expect(commonRules.password.custom!(password)).toContain('Password must contain');
    });
  });

  it('validates username format correctly', () => {
    const validUsernames = ['user123', 'test_user', 'user-name', 'TestUser'];
    const invalidUsernames = ['user@name', 'user name', 'user.name', 'user#123'];

    validUsernames.forEach(username => {
      expect(commonRules.username.custom!(username)).toBeNull();
    });

    invalidUsernames.forEach(username => {
      expect(commonRules.username.custom!(username)).toBe(
        'Username can only contain letters, numbers, hyphens, and underscores'
      );
    });
  });
});