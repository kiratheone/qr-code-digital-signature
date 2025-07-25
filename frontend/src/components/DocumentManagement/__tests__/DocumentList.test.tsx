import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import DocumentList from '../DocumentList';

// Mock date-fns to avoid issues with dynamic time formatting
jest.mock('date-fns', () => ({
  formatDistanceToNow: jest.fn(() => '2 days ago'),
}));

describe('DocumentList Component', () => {
  const mockDocuments = [
    {
      id: '123e4567-e89b-12d3-a456-426614174000',
      filename: 'test-document.pdf',
      issuer: 'Test Issuer',
      documentHash: 'abc123hash',
      createdAt: '2023-01-01T12:00:00Z',
      status: 'active',
    },
    {
      id: '223e4567-e89b-12d3-a456-426614174001',
      filename: 'another-document.pdf',
      issuer: 'Another Issuer',
      documentHash: 'def456hash',
      createdAt: '2023-01-02T12:00:00Z',
      status: 'active',
    },
  ];

  const mockHandlers = {
    onViewDetails: jest.fn(),
    onDelete: jest.fn(),
  };

  test('renders loading state correctly', () => {
    render(
      <DocumentList
        documents={[]}
        isLoading={true}
        onViewDetails={mockHandlers.onViewDetails}
        onDelete={mockHandlers.onDelete}
      />
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  test('renders empty state correctly', () => {
    render(
      <DocumentList
        documents={[]}
        isLoading={false}
        onViewDetails={mockHandlers.onViewDetails}
        onDelete={mockHandlers.onDelete}
      />
    );

    expect(screen.getByText('No documents found')).toBeInTheDocument();
    expect(screen.getByText('Get started by signing a new document.')).toBeInTheDocument();
  });

  test('renders document list correctly', () => {
    render(
      <DocumentList
        documents={mockDocuments}
        isLoading={false}
        onViewDetails={mockHandlers.onViewDetails}
        onDelete={mockHandlers.onDelete}
      />
    );

    expect(screen.getByText('test-document.pdf')).toBeInTheDocument();
    expect(screen.getByText('another-document.pdf')).toBeInTheDocument();
    expect(screen.getByText('Test Issuer')).toBeInTheDocument();
    expect(screen.getByText('Another Issuer')).toBeInTheDocument();
  });

  test('calls onViewDetails when View button is clicked', () => {
    render(
      <DocumentList
        documents={mockDocuments}
        isLoading={false}
        onViewDetails={mockHandlers.onViewDetails}
        onDelete={mockHandlers.onDelete}
      />
    );

    fireEvent.click(screen.getAllByText('View')[0]);
    expect(mockHandlers.onViewDetails).toHaveBeenCalledWith(mockDocuments[0]);
  });

  test('calls onDelete when Delete button is clicked', () => {
    render(
      <DocumentList
        documents={mockDocuments}
        isLoading={false}
        onViewDetails={mockHandlers.onViewDetails}
        onDelete={mockHandlers.onDelete}
      />
    );

    fireEvent.click(screen.getAllByText('Delete')[0]);
    expect(mockHandlers.onDelete).toHaveBeenCalledWith(mockDocuments[0]);
  });
});