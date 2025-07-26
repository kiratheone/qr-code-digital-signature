import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import DocumentUploadForm from '../DocumentUploadForm';

// Mock react-dropzone
jest.mock('react-dropzone', () => ({
  useDropzone: () => ({
    getRootProps: () => ({}),
    getInputProps: () => ({}),
    isDragActive: false,
    isDragReject: false,
    fileRejections: [],
  }),
}));

describe('DocumentUploadForm', () => {
  const mockOnSubmit = jest.fn();
  
  beforeEach(() => {
    mockOnSubmit.mockReset();
  });
  
  it('renders the form correctly', () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    expect(screen.getByText('Upload Document for Signing')).toBeInTheDocument();
    expect(screen.getByText('Drag & drop your PDF here')).toBeInTheDocument();
    expect(screen.getByTestId('issuer-input')).toBeInTheDocument();
    expect(screen.getByTestId('description-input')).toBeInTheDocument();
    expect(screen.getByTestId('submit-button')).toBeDisabled();
  });
  
  it('shows validation error when issuer is not provided', async () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Submit without issuer
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText(/Issuer name is required/i)).toBeInTheDocument();
    });
    
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });
  
  it('calls onSubmit with correct data when form is valid', async () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Fill issuer field
    const issuerInput = screen.getByTestId('issuer-input');
    fireEvent.change(issuerInput, { target: { value: 'Test Organization' } });
    
    // Fill description field
    const descriptionInput = screen.getByTestId('description-input');
    fireEvent.change(descriptionInput, { target: { value: 'Test Description' } });
    
    // Submit form
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        file,
        issuer: 'Test Organization',
        description: 'Test Description',
        position: {
          page: undefined,
          x: undefined,
          y: undefined
        }
      });
    });
  });
  
  it('shows loading state during submission', async () => {
    mockOnSubmit.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
    
    render(<DocumentUploadForm onSubmit={mockOnSubmit} isLoading={true} />);
    
    expect(screen.getByTestId('submit-button')).toBeDisabled();
    expect(screen.getByText(/Uploading/i)).toBeInTheDocument();
  });
  
  it('shows custom position options when checkbox is checked', async () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Initially, position options should be hidden
    expect(screen.queryByTestId('page-input')).not.toBeInTheDocument();
    
    // Check the custom position checkbox
    const checkbox = screen.getByTestId('custom-position-checkbox');
    fireEvent.click(checkbox);
    
    // Position options should now be visible
    expect(screen.getByTestId('page-input')).toBeInTheDocument();
    expect(screen.getByTestId('x-position-input')).toBeInTheDocument();
    expect(screen.getByTestId('y-position-input')).toBeInTheDocument();
  });
  
  it('includes custom position data in form submission', async () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Fill issuer field
    const issuerInput = screen.getByTestId('issuer-input');
    fireEvent.change(issuerInput, { target: { value: 'Test Organization' } });
    
    // Check the custom position checkbox
    const checkbox = screen.getByTestId('custom-position-checkbox');
    fireEvent.click(checkbox);
    
    // Fill position fields
    const pageInput = screen.getByTestId('page-input');
    fireEvent.change(pageInput, { target: { value: '2' } });
    
    const xInput = screen.getByTestId('x-position-input');
    fireEvent.change(xInput, { target: { value: '50' } });
    
    const yInput = screen.getByTestId('y-position-input');
    fireEvent.change(yInput, { target: { value: '75' } });
    
    // Submit form
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        file,
        issuer: 'Test Organization',
        position: {
          page: 2,
          x: 50,
          y: 75
        }
      });
    });
  });
  
  it('shows progress indicator during upload', async () => {
    // Mock a delayed submission
    mockOnSubmit.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 500)));
    
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Fill issuer field
    const issuerInput = screen.getByTestId('issuer-input');
    fireEvent.change(issuerInput, { target: { value: 'Test Organization' } });
    
    // Submit form
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    // Progress indicator should appear
    await waitFor(() => {
      expect(screen.getByText(/Uploading document/i)).toBeInTheDocument();
    });
    
    // Progress bar should be visible
    const progressBar = document.querySelector('.bg-blue-600');
    expect(progressBar).toBeInTheDocument();
  });
  
  it('shows success message after successful upload', async () => {
    mockOnSubmit.mockResolvedValue(undefined);
    
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Fill issuer field
    const issuerInput = screen.getByTestId('issuer-input');
    fireEvent.change(issuerInput, { target: { value: 'Test Organization' } });
    
    // Submit form
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    // Success message should appear
    await waitFor(() => {
      expect(screen.getByText(/Document uploaded successfully/i)).toBeInTheDocument();
    });
  });
  
  it('validates issuer name length', async () => {
    render(<DocumentUploadForm onSubmit={mockOnSubmit} />);
    
    // Mock file selection
    const fileInput = screen.getByTestId('file-input');
    const file = new File(['dummy content'], 'test.pdf', { type: 'application/pdf' });
    Object.defineProperty(fileInput, 'files', { value: [file] });
    fireEvent.change(fileInput);
    
    // Fill issuer field with too long text
    const issuerInput = screen.getByTestId('issuer-input');
    const tooLongIssuer = 'A'.repeat(256); // 256 characters
    fireEvent.change(issuerInput, { target: { value: tooLongIssuer } });
    
    // Submit form
    const submitButton = screen.getByTestId('submit-button');
    fireEvent.click(submitButton);
    
    // Error message should appear
    await waitFor(() => {
      expect(screen.getByText(/Issuer name cannot exceed 255 characters/i)).toBeInTheDocument();
    });
    
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });
});
