import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { FormErrorHandler, useFormErrorHandler } from '../FormErrorHandler';
import { NotificationProvider } from '../Notifications';

// Test wrapper component
const TestWrapper = ({ children }: { children: React.ReactNode }) => (
  <NotificationProvider>
    {children}
  </NotificationProvider>
);

// Mock form component that works with FormErrorHandler
const MockFormWithHandler = ({ 
  shouldFail = false, 
  failureType = 'server',
  data = { test: 'data' },
  onSuccess,
  onError
}: {
  shouldFail?: boolean;
  failureType?: 'server' | 'network' | 'validation' | 'timeout';
  data?: any;
  onSuccess?: (data: any) => void;
  onError?: (error: unknown) => void;
}) => {
  const mockSubmit = async (formData: any) => {
    if (shouldFail) {
      let error: any;
      switch (failureType) {
        case 'network':
          error = new Error('Network error');
          break;
        case 'validation':
          error = {
            status: 400,
            message: 'Validation failed',
            validation_errors: [
              { field: 'email', message: 'Email is required', code: 'required' },
              { field: 'password', message: 'Password too short', code: 'min_length' }
            ]
          };
          break;
        case 'timeout':
          error = { status: 408, message: 'Request timeout' };
          break;
        default:
          error = { status: 500, message: 'Server error' };
      }
      throw error;
    }
    return formData;
  };

  return (
    <FormErrorHandler onSubmit={mockSubmit} onSuccess={onSuccess} onError={onError} showNotifications={false}>
      <form data-testid="mock-form">
        <input name="test" defaultValue="value" />
        <button type="submit">Submit</button>
      </form>
    </FormErrorHandler>
  );
};

describe('FormErrorHandler', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders children without errors', () => {
    render(
      <TestWrapper>
        <MockFormWithHandler />
      </TestWrapper>
    );
    
    expect(screen.getByTestId('mock-form')).toBeInTheDocument();
    expect(screen.getByText('Submit')).toBeInTheDocument();
  });

  it('handles successful form submission', async () => {
    const mockSuccess = jest.fn();
    
    render(
      <TestWrapper>
        <MockFormWithHandler onSuccess={mockSuccess} />
      </TestWrapper>
    );
    
    const submitButton = screen.getByText('Submit');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockSuccess).toHaveBeenCalled();
    });
  });

  it('displays server errors', async () => {
    render(
      <TestWrapper>
        <MockFormWithHandler shouldFail={true} failureType="server" />
      </TestWrapper>
    );
    
    const submitButton = screen.getByText('Submit');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Form Error')).toBeInTheDocument();
      expect(screen.getByText('Server error')).toBeInTheDocument();
    });
  });

  it('displays validation errors with field names', async () => {
    render(
      <TestWrapper>
        <MockFormWithHandler shouldFail={true} failureType="validation" />
      </TestWrapper>
    );
    
    const submitButton = screen.getByText('Submit');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('email Error')).toBeInTheDocument();
      expect(screen.getByText('Email is required')).toBeInTheDocument();
      expect(screen.getByText('password Error')).toBeInTheDocument();
      expect(screen.getByText('Password too short')).toBeInTheDocument();
    });
  });

  it('calls onError callback on failure', async () => {
    const mockError = jest.fn();
    
    render(
      <TestWrapper>
        <MockFormWithHandler shouldFail={true} onError={mockError} />
      </TestWrapper>
    );
    
    const submitButton = screen.getByText('Submit');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockError).toHaveBeenCalledWith(expect.any(Error));
    });
  });
});

// Test component for useFormErrorHandler hook
const FormErrorHookTest = () => {
  const { errorState, handleFormSubmit, clearErrors } = useFormErrorHandler({
    maxRetries: 2,
    showNotifications: false,
  });

  const mockSubmit = async (data: any) => {
    if (data.shouldFail) {
      throw new Error('Test error');
    }
  };

  const handleSubmit = (shouldFail: boolean) => {
    handleFormSubmit(mockSubmit, { shouldFail });
  };

  return (
    <div>
      <button onClick={() => handleSubmit(false)}>Submit Success</button>
      <button onClick={() => handleSubmit(true)}>Submit Fail</button>
      <button onClick={clearErrors}>Clear Errors</button>
      
      <div data-testid="error-count">{errorState.errors.length}</div>
      <div data-testid="is-submitting">{errorState.isSubmitting.toString()}</div>
      <div data-testid="submit-attempts">{errorState.submitAttempts}</div>
      <div data-testid="can-retry">{errorState.canRetry.toString()}</div>
      
      {errorState.errors.map((error, index) => (
        <div key={index} data-testid={`error-${index}`}>
          {error.message}
        </div>
      ))}
    </div>
  );
};

describe('useFormErrorHandler hook', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('handles successful submission', async () => {
    render(
      <TestWrapper>
        <FormErrorHookTest />
      </TestWrapper>
    );
    
    const successButton = screen.getByText('Submit Success');
    fireEvent.click(successButton);
    
    expect(screen.getByTestId('is-submitting')).toHaveTextContent('true');
    
    await waitFor(() => {
      expect(screen.getByTestId('is-submitting')).toHaveTextContent('false');
      expect(screen.getByTestId('error-count')).toHaveTextContent('0');
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('0');
    });
  });

  it('handles failed submission', async () => {
    render(
      <TestWrapper>
        <FormErrorHookTest />
      </TestWrapper>
    );
    
    const failButton = screen.getByText('Submit Fail');
    fireEvent.click(failButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('error-count')).toHaveTextContent('1');
      expect(screen.getByTestId('error-0')).toHaveTextContent('Test error');
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('1');
    });
  });

  it('clears errors', async () => {
    render(
      <TestWrapper>
        <FormErrorHookTest />
      </TestWrapper>
    );
    
    const failButton = screen.getByText('Submit Fail');
    fireEvent.click(failButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('error-count')).toHaveTextContent('1');
    });
    
    const clearButton = screen.getByText('Clear Errors');
    fireEvent.click(clearButton);
    
    expect(screen.getByTestId('error-count')).toHaveTextContent('0');
    expect(screen.getByTestId('can-retry')).toHaveTextContent('false');
  });

  it('tracks submission attempts', async () => {
    render(
      <TestWrapper>
        <FormErrorHookTest />
      </TestWrapper>
    );
    
    const failButton = screen.getByText('Submit Fail');
    
    // First attempt
    fireEvent.click(failButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('1');
    });
    
    // Second attempt
    fireEvent.click(failButton);
    
    await waitFor(() => {
      expect(screen.getByTestId('submit-attempts')).toHaveTextContent('2');
    });
  });
});