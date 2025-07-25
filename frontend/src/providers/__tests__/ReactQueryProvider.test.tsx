import { render } from '@testing-library/react';
import ReactQueryProvider from '../ReactQueryProvider';

// Mock the ReactQueryDevtools
jest.mock('@tanstack/react-query-devtools', () => ({
  ReactQueryDevtools: ({ initialIsOpen }: { initialIsOpen: boolean }) => 
    <div data-testid="react-query-devtools" />,
}));

describe('ReactQueryProvider', () => {
  const originalNodeEnv = process.env.NODE_ENV;
  
  afterEach(() => {
    process.env.NODE_ENV = originalNodeEnv;
  });
  
  it('renders children correctly', () => {
    const { getByText } = render(
      <ReactQueryProvider>
        <div>Test Child</div>
      </ReactQueryProvider>
    );
    
    expect(getByText('Test Child')).toBeInTheDocument();
  });
  
  it('renders ReactQueryDevtools in development environment', () => {
    process.env.NODE_ENV = 'development';
    
    const { queryByTestId } = render(
      <ReactQueryProvider>
        <div>Test Child</div>
      </ReactQueryProvider>
    );
    
    expect(queryByTestId('react-query-devtools')).toBeInTheDocument();
  });
  
  it('does not render ReactQueryDevtools in production environment', () => {
    process.env.NODE_ENV = 'production';
    
    const { queryByTestId } = render(
      <ReactQueryProvider>
        <div>Test Child</div>
      </ReactQueryProvider>
    );
    
    expect(queryByTestId('react-query-devtools')).not.toBeInTheDocument();
  });
});