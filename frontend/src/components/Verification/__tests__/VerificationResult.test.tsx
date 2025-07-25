import React from 'react';
import { render, screen } from '@testing-library/react';
import VerificationResult from '../VerificationResult';
import { VerificationResponse } from '@/types/verification';

// Mock the date utils
jest.mock('@/utils/dateUtils', () => ({
  formatDate: (dateString: string) => new Date(dateString).toLocaleDateString(),
}));

describe('VerificationResult', () => {
  const baseResult: VerificationResponse = {
    documentId: 'test-doc-id',
    filename: 'test-document.pdf',
    issuer: 'Test Organization',
    createdAt: '2024-01-15T10:30:00Z',
    status: 'valid',
    message: 'Document is valid and has not been modified',
  };

  it('renders valid verification result correctly', () => {
    const validResult: VerificationResponse = {
      ...baseResult,
      status: 'valid',
      message: 'Document is valid and has not been modified',
    };

    render(<VerificationResult result={validResult} />);

    expect(screen.getByText('✅ Document is valid')).toBeInTheDocument();
    expect(screen.getByText('Document is valid and has not been modified')).toBeInTheDocument();
    expect(screen.getByText('test-document.pdf')).toBeInTheDocument();
    expect(screen.getByText('Test Organization')).toBeInTheDocument();
    expect(screen.getByText('✅ Valid')).toBeInTheDocument();
  });

  it('renders modified verification result correctly', () => {
    const modifiedResult: VerificationResponse = {
      ...baseResult,
      status: 'modified',
      message: 'QR code is valid but document content has been changed',
    };

    render(<VerificationResult result={modifiedResult} />);

    expect(screen.getByText('⚠️ QR valid, but file content has changed')).toBeInTheDocument();
    expect(screen.getByText('QR code is valid but document content has been changed')).toBeInTheDocument();
    expect(screen.getByText('⚠️ Modified')).toBeInTheDocument();
  });

  it('renders invalid verification result correctly', () => {
    const invalidResult: VerificationResponse = {
      ...baseResult,
      status: 'invalid',
      message: 'Digital signature is invalid or document has been tampered with',
    };

    render(<VerificationResult result={invalidResult} />);

    expect(screen.getByText('❌ QR invalid / signature incorrect')).toBeInTheDocument();
    expect(screen.getByText('Digital signature is invalid or document has been tampered with')).toBeInTheDocument();
    expect(screen.getByText('❌ Invalid')).toBeInTheDocument();
  });

  it('renders pending verification result correctly', () => {
    const pendingResult: VerificationResponse = {
      ...baseResult,
      status: 'pending',
      message: 'Verification is in progress',
    };

    render(<VerificationResult result={pendingResult} />);

    expect(screen.getByText('Verification pending')).toBeInTheDocument();
    expect(screen.getByText('Verification is in progress')).toBeInTheDocument();
    expect(screen.getByText('Pending')).toBeInTheDocument();
  });

  it('displays document information correctly', () => {
    render(<VerificationResult result={baseResult} />);

    expect(screen.getByText('Document Information')).toBeInTheDocument();
    expect(screen.getByText('Document Name')).toBeInTheDocument();
    expect(screen.getByText('test-document.pdf')).toBeInTheDocument();
    expect(screen.getByText('Issuer')).toBeInTheDocument();
    expect(screen.getByText('Test Organization')).toBeInTheDocument();
    expect(screen.getByText('Creation Date')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
  });

  it('displays additional details when provided', () => {
    const resultWithDetails: VerificationResponse = {
      ...baseResult,
      details: 'Hash comparison: MATCH\nSignature verification: VALID\nTimestamp: Valid',
    };

    render(<VerificationResult result={resultWithDetails} />);

    expect(screen.getByText('Additional Details')).toBeInTheDocument();
    expect(screen.getByText('Hash comparison: MATCH\nSignature verification: VALID\nTimestamp: Valid')).toBeInTheDocument();
  });

  it('does not display additional details section when not provided', () => {
    render(<VerificationResult result={baseResult} />);

    expect(screen.queryByText('Additional Details')).not.toBeInTheDocument();
  });

  it('applies correct styling for valid status', () => {
    const validResult: VerificationResponse = {
      ...baseResult,
      status: 'valid',
    };

    render(<VerificationResult result={validResult} />);

    const resultContainer = screen.getByTestId('verification-result');
    expect(resultContainer).toHaveClass('text-green-700', 'bg-green-50', 'border-green-200');
  });

  it('applies correct styling for modified status', () => {
    const modifiedResult: VerificationResponse = {
      ...baseResult,
      status: 'modified',
    };

    render(<VerificationResult result={modifiedResult} />);

    const resultContainer = screen.getByTestId('verification-result');
    expect(resultContainer).toHaveClass('text-yellow-700', 'bg-yellow-50', 'border-yellow-200');
  });

  it('applies correct styling for invalid status', () => {
    const invalidResult: VerificationResponse = {
      ...baseResult,
      status: 'invalid',
    };

    render(<VerificationResult result={invalidResult} />);

    const resultContainer = screen.getByTestId('verification-result');
    expect(resultContainer).toHaveClass('text-red-700', 'bg-red-50', 'border-red-200');
  });

  it('applies correct styling for pending status', () => {
    const pendingResult: VerificationResponse = {
      ...baseResult,
      status: 'pending',
    };

    render(<VerificationResult result={pendingResult} />);

    const resultContainer = screen.getByTestId('verification-result');
    expect(resultContainer).toHaveClass('text-gray-700', 'bg-gray-50', 'border-gray-200');
  });

  it('displays formatted creation date', () => {
    render(<VerificationResult result={baseResult} />);

    // The date should be formatted by the mocked formatDate function
    expect(screen.getByText('1/15/2024')).toBeInTheDocument();
  });
});