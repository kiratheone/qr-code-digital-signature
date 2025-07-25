import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { 
  validateField, 
  useFormValidation, 
  ValidatedInput, 
  ValidatedTextarea, 
  ValidatedSelect,
  ValidationRule 
} from '../FormValidation';

describe('validateField', () => {
  it('validates required fields', () => {
    const rules: ValidationRule = { required: true };
    
    expect(validateField('', rules)).toBe('This field is required');
    expect(validateField('   ', rules)).toBe('This field is required');
    expect(validateField(null, rules)).toBe('This field is required');
    expect(validateField(undefined, rules)).toBe('This field is required');
    expect(validateField('value', rules)).toBeNull();
  });

  it('validates string length', () => {
    const rules: ValidationRule = { minLength: 3, maxLength: 10 };
    
    expect(validateField('ab', rules)).toBe('Must be at least 3 characters');
    expect(validateField('abc', rules)).toBeNull();
    expect(validateField('abcdefghijk', rules)).toBe('Must be no more than 10 characters');
    expect(validateField('abcdefghij', rules)).toBeNull();
  });

  it('validates email format', () => {
    const rules: ValidationRule = { email: true };
    
    expect(validateField('invalid-email', rules)).toBe('Please enter a valid email address');
    expect(validateField('test@', rules)).toBe('Please enter a valid email address');
    expect(validateField('test@example.com', rules)).toBeNull();
  });

  it('validates URL format', () => {
    const rules: ValidationRule = { url: true };
    
    expect(validateField('invalid-url', rules)).toBe('Please enter a valid URL');
    expect(validateField('http://example.com', rules)).toBeNull();
    expect(validateField('https://example.com', rules)).toBeNull();
  });

  it('validates numbers', () => {
    const rules: ValidationRule = { number: true, min: 0, max: 100 };
    
    expect(validateField('not-a-number', rules)).toBe('Please enter a valid number');
    expect(validateField('-1', rules)).toBe('Must be at least 0');
    expect(validateField('101', rules)).toBe('Must be no more than 100');
    expect(validateField('50', rules)).toBeNull();
  });

  it('validates with custom function', () => {
    const rules: ValidationRule = {
      custom: (value) => value === 'forbidden' ? 'This value is not allowed' : null
    };
    
    expect(validateField('forbidden', rules)).toBe('This value is not allowed');
    expect(validateField('allowed', rules)).toBeNull();
  });

  it('validates with regex pattern', () => {
    const rules: ValidationRule = { pattern: /^\d{3}-\d{3}-\d{4}$/ };
    
    expect(validateField('123-456-789', rules)).toBe('Please enter a valid format');
    expect(validateField('123-456-7890', rules)).toBeNull();
  });

  it('skips validation for empty non-required fields', () => {
    const rules: ValidationRule = { minLength: 5, email: true };
    
    expect(validateField('', rules)).toBeNull();
    expect(validateField(null, rules)).toBeNull();
  });
});

// Test component for useFormValidation hook
const TestFormComponent = () => {
  const {
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
  } = useFormValidation({
    email: {
      value: '',
      rules: { required: true, email: true },
    },
    password: {
      value: '',
      rules: { required: true, minLength: 8 },
    },
  });

  const mockSubmit = async (data: Record<string, any>) => {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 100));
    if (data.email === 'fail@example.com') {
      throw new Error('Submission failed');
    }
  };

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    handleSubmit(mockSubmit);
  };

  return (
    <form onSubmit={onSubmit}>
      <input
        data-testid="email"
        value={fields.email.value}
        onChange={(e) => updateField('email', e.target.value)}
        onBlur={() => {
          touchField('email');
          validateSingleField('email', fields.email.value);
        }}
      />
      {fields.email.error && fields.email.touched && (
        <span data-testid="email-error">{fields.email.error}</span>
      )}
      
      <input
        data-testid="password"
        type="password"
        value={fields.password.value}
        onChange={(e) => updateField('password', e.target.value)}
        onBlur={() => {
          touchField('password');
          validateSingleField('password', fields.password.value);
        }}
      />
      {fields.password.error && fields.password.touched && (
        <span data-testid="password-error">{fields.password.error}</span>
      )}
      
      {submitError && (
        <div data-testid="submit-error">{submitError}</div>
      )}
      
      <button type="submit" disabled={!canSubmit}>
        {isSubmitting ? 'Submitting...' : 'Submit'}
      </button>
      <button type="button" onClick={resetForm}>
        Reset
      </button>
      
      <span data-testid="form-valid">{isFormValid.toString()}</span>
      <span data-testid="has-errors">{hasErrors.toString()}</span>
      <span data-testid="is-submitting">{isSubmitting.toString()}</span>
      <span data-testid="submit-attempts">{submitAttempts}</span>
      <span data-testid="can-submit">{canSubmit.toString()}</span>
    </form>
  );
};

describe('useFormValidation', () => {
  it('updates field values', () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    
    expect(emailInput).toHaveValue('test@example.com');
  });

  it('validates single field on blur', async () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });
    fireEvent.blur(emailInput);
    
    await waitFor(() => {
      expect(screen.getByTestId('email-error')).toHaveTextContent('Please enter a valid email address');
    });
  });

  it('validates all fields on submit', async () => {
    render(<TestFormComponent />);
    
    const form = screen.getByRole('form');
    fireEvent.submit(form);
    
    await waitFor(() => {
      expect(screen.getByTestId('email-error')).toHaveTextContent('This field is required');
      expect(screen.getByTestId('password-error')).toHaveTextContent('This field is required');
    });
  });

  it('tracks form validity', async () => {
    render(<TestFormComponent />);
    
    expect(screen.getByTestId('form-valid')).toHaveTextContent('false');
    
    const emailInput = screen.getByTestId('email');
    const passwordInput = screen.getByTestId('password');
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    await waitFor(() => {
      expect(screen.getByTestId('form-valid')).toHaveTextContent('true');
    });
  });

  it('resets form', async () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    const resetButton = screen.getByText('Reset');
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.blur(emailInput);
    
    fireEvent.click(resetButton);
    
    await waitFor(() => {
      expect(emailInput).toHaveValue('');
      expect(screen.queryByTestId('email-error')).not.toBeInTheDocument();
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('0');
    });
  });

  it('handles form submission successfully', async () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    const passwordInput = screen.getByTestId('password');
    const submitButton = screen.getByText('Submit');
    
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    await waitFor(() => {
      expect(screen.getByTestId('can-submit')).toHaveTextContent('true');
    });
    
    fireEvent.click(submitButton);
    
    expect(screen.getByTestId('is-submitting')).toHaveTextContent('true');
    expect(screen.getByText('Submitting...')).toBeInTheDocument();
    
    await waitFor(() => {
      expect(screen.getByTestId('is-submitting')).toHaveTextContent('false');
      expect(screen.getByText('Submit')).toBeInTheDocument();
    });
  });

  it('handles form submission failure', async () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    const passwordInput = screen.getByTestId('password');
    const submitButton = screen.getByText('Submit');
    
    fireEvent.change(emailInput, { target: { value: 'fail@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    await waitFor(() => {
      expect(screen.getByTestId('can-submit')).toHaveTextContent('true');
    });
    
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('submit-error')).toHaveTextContent('Submission failed');
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('1');
    });
  });

  it('prevents submission when form is invalid', async () => {
    render(<TestFormComponent />);
    
    const submitButton = screen.getByText('Submit');
    
    expect(screen.getByTestId('can-submit')).toHaveTextContent('false');
    expect(submitButton).toBeDisabled();
    
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('submit-error')).toHaveTextContent('Please fix 2 errors before submitting.');
    });
  });

  it('clears submit error when validation errors are fixed', async () => {
    render(<TestFormComponent />);
    
    const emailInput = screen.getByTestId('email');
    const submitButton = screen.getByText('Submit');
    
    // Try to submit invalid form
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('submit-error')).toBeInTheDocument();
    });
    
    // Fix validation error
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    
    await waitFor(() => {
      expect(screen.queryByTestId('submit-error')).not.toBeInTheDocument();
    });
  });
});

describe('ValidatedInput', () => {
  it('renders input with label', () => {
    render(
      <ValidatedInput
        label="Email"
        value=""
        onChange={() => {}}
      />
    );
    
    expect(screen.getByLabelText('Email')).toBeInTheDocument();
  });

  it('shows required indicator', () => {
    render(
      <ValidatedInput
        label="Email"
        value=""
        onChange={() => {}}
        required
      />
    );
    
    expect(screen.getByText('*')).toBeInTheDocument();
  });

  it('shows error message when touched', () => {
    render(
      <ValidatedInput
        label="Email"
        value=""
        onChange={() => {}}
        error="This field is required"
        touched={true}
      />
    );
    
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });

  it('shows help text when no error', () => {
    render(
      <ValidatedInput
        label="Email"
        value=""
        onChange={() => {}}
        helpText="Enter your email address"
      />
    );
    
    expect(screen.getByText('Enter your email address')).toBeInTheDocument();
  });

  it('applies error styling', () => {
    render(
      <ValidatedInput
        label="Email"
        value=""
        onChange={() => {}}
        error="This field is required"
        touched={true}
      />
    );
    
    const input = screen.getByLabelText('Email');
    expect(input).toHaveClass('border-red-300');
  });
});

describe('ValidatedTextarea', () => {
  it('renders textarea with label', () => {
    render(
      <ValidatedTextarea
        label="Description"
        value=""
        onChange={() => {}}
      />
    );
    
    expect(screen.getByLabelText('Description')).toBeInTheDocument();
  });

  it('shows error message when touched', () => {
    render(
      <ValidatedTextarea
        label="Description"
        value=""
        onChange={() => {}}
        error="This field is required"
        touched={true}
      />
    );
    
    expect(screen.getByText('This field is required')).toBeInTheDocument();
  });
});

describe('ValidatedSelect', () => {
  const options = [
    { value: 'option1', label: 'Option 1' },
    { value: 'option2', label: 'Option 2' },
  ];

  it('renders select with options', () => {
    render(
      <ValidatedSelect
        label="Choose Option"
        value=""
        onChange={() => {}}
        options={options}
      />
    );
    
    expect(screen.getByLabelText('Choose Option')).toBeInTheDocument();
    expect(screen.getByText('Option 1')).toBeInTheDocument();
    expect(screen.getByText('Option 2')).toBeInTheDocument();
  });

  it('shows error message when touched', () => {
    render(
      <ValidatedSelect
        label="Choose Option"
        value=""
        onChange={() => {}}
        options={options}
        error="Please select an option"
        touched={true}
      />
    );
    
    expect(screen.getByText('Please select an option')).toBeInTheDocument();
  });
});