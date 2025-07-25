import React from 'react';
import { render, screen, act } from '@testing-library/react';
import { 
  ProgressIndicator, 
  useProgressSteps, 
  AnimatedProgressBar, 
  FormProgress,
  ProgressStep 
} from '../ProgressIndicator';

const mockSteps: ProgressStep[] = [
  { id: 'step1', label: 'Step 1', description: 'First step', status: 'completed' },
  { id: 'step2', label: 'Step 2', description: 'Second step', status: 'active' },
  { id: 'step3', label: 'Step 3', description: 'Third step', status: 'pending' },
];

describe('ProgressIndicator', () => {
  it('renders horizontal progress indicator', () => {
    render(<ProgressIndicator steps={mockSteps} />);
    
    expect(screen.getByText('Step 1')).toBeInTheDocument();
    expect(screen.getByText('Step 2')).toBeInTheDocument();
    expect(screen.getByText('Step 3')).toBeInTheDocument();
  });

  it('renders vertical progress indicator', () => {
    render(<ProgressIndicator steps={mockSteps} variant="vertical" />);
    
    expect(screen.getByText('Step 1')).toBeInTheDocument();
    expect(screen.getByText('Step 2')).toBeInTheDocument();
    expect(screen.getByText('Step 3')).toBeInTheDocument();
  });

  it('shows descriptions when enabled', () => {
    render(<ProgressIndicator steps={mockSteps} showDescriptions={true} />);
    
    expect(screen.getByText('First step')).toBeInTheDocument();
    expect(screen.getByText('Second step')).toBeInTheDocument();
    expect(screen.getByText('Third step')).toBeInTheDocument();
  });

  it('hides labels when disabled', () => {
    render(<ProgressIndicator steps={mockSteps} showLabels={false} />);
    
    expect(screen.queryByText('Step 1')).not.toBeInTheDocument();
    expect(screen.queryByText('Step 2')).not.toBeInTheDocument();
    expect(screen.queryByText('Step 3')).not.toBeInTheDocument();
  });

  it('displays different step statuses correctly', () => {
    const stepsWithDifferentStatuses: ProgressStep[] = [
      { id: 'completed', label: 'Completed', status: 'completed' },
      { id: 'active', label: 'Active', status: 'active' },
      { id: 'error', label: 'Error', status: 'error' },
      { id: 'pending', label: 'Pending', status: 'pending' },
    ];

    render(<ProgressIndicator steps={stepsWithDifferentStatuses} />);
    
    // Check that all steps are rendered
    expect(screen.getByText('Completed')).toBeInTheDocument();
    expect(screen.getByText('Active')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Pending')).toBeInTheDocument();
  });
});

// Test component for useProgressSteps hook
const ProgressStepsTest = () => {
  const {
    steps,
    updateStep,
    setStepStatus,
    nextStep,
    previousStep,
    resetSteps,
    completeAllSteps,
    getCurrentStep,
    getProgress,
  } = useProgressSteps([
    { id: 'step1', label: 'Step 1', status: 'active' },
    { id: 'step2', label: 'Step 2', status: 'pending' },
    { id: 'step3', label: 'Step 3', status: 'pending' },
  ]);

  const currentStep = getCurrentStep();
  const progress = getProgress();

  return (
    <div>
      <div data-testid="current-step">
        {currentStep ? currentStep.label : 'None'}
      </div>
      <div data-testid="progress">{progress}</div>
      
      <button onClick={nextStep}>Next Step</button>
      <button onClick={previousStep}>Previous Step</button>
      <button onClick={resetSteps}>Reset Steps</button>
      <button onClick={completeAllSteps}>Complete All</button>
      <button onClick={() => setStepStatus('step2', 'error')}>Set Step 2 Error</button>
      
      <div data-testid="steps">
        {steps.map(step => (
          <div key={step.id} data-testid={`step-${step.id}`}>
            {step.label}: {step.status}
          </div>
        ))}
      </div>
    </div>
  );
};

describe('useProgressSteps hook', () => {
  it('manages step progression', () => {
    render(<ProgressStepsTest />);
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 1');
    expect(screen.getByTestId('progress')).toHaveTextContent('0');
    
    const nextButton = screen.getByText('Next Step');
    act(() => {
      nextButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 2');
    expect(screen.getByTestId('step-step1')).toHaveTextContent('Step 1: completed');
    expect(screen.getByTestId('step-step2')).toHaveTextContent('Step 2: active');
  });

  it('handles previous step', () => {
    render(<ProgressStepsTest />);
    
    const nextButton = screen.getByText('Next Step');
    const prevButton = screen.getByText('Previous Step');
    
    // Go to next step first
    act(() => {
      nextButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 2');
    
    // Go back to previous step
    act(() => {
      prevButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 1');
    expect(screen.getByTestId('step-step2')).toHaveTextContent('Step 2: pending');
  });

  it('resets all steps', () => {
    render(<ProgressStepsTest />);
    
    const nextButton = screen.getByText('Next Step');
    const resetButton = screen.getByText('Reset Steps');
    
    // Progress through steps
    act(() => {
      nextButton.click();
      nextButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 3');
    
    // Reset steps
    act(() => {
      resetButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('Step 1');
    expect(screen.getByTestId('step-step1')).toHaveTextContent('Step 1: active');
    expect(screen.getByTestId('step-step2')).toHaveTextContent('Step 2: pending');
    expect(screen.getByTestId('step-step3')).toHaveTextContent('Step 3: pending');
  });

  it('completes all steps', () => {
    render(<ProgressStepsTest />);
    
    const completeButton = screen.getByText('Complete All');
    
    act(() => {
      completeButton.click();
    });
    
    expect(screen.getByTestId('current-step')).toHaveTextContent('None');
    expect(screen.getByTestId('progress')).toHaveTextContent('100');
    expect(screen.getByTestId('step-step1')).toHaveTextContent('Step 1: completed');
    expect(screen.getByTestId('step-step2')).toHaveTextContent('Step 2: completed');
    expect(screen.getByTestId('step-step3')).toHaveTextContent('Step 3: completed');
  });

  it('sets step status', () => {
    render(<ProgressStepsTest />);
    
    const errorButton = screen.getByText('Set Step 2 Error');
    
    act(() => {
      errorButton.click();
    });
    
    expect(screen.getByTestId('step-step2')).toHaveTextContent('Step 2: error');
  });
});

describe('AnimatedProgressBar', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders progress bar', () => {
    render(<AnimatedProgressBar progress={50} />);
    
    expect(screen.getByText('Progress')).toBeInTheDocument();
  });

  it('shows percentage when enabled', () => {
    render(<AnimatedProgressBar progress={75} showPercentage={true} />);
    
    // Initially shows 0% due to animation
    expect(screen.getByText('0%')).toBeInTheDocument();
  });

  it('applies different colors', () => {
    const { rerender } = render(<AnimatedProgressBar progress={50} color="green" />);
    
    // Test that component renders without error
    expect(screen.getByText('Progress')).toBeInTheDocument();
    
    rerender(<AnimatedProgressBar progress={50} color="red" />);
    expect(screen.getByText('Progress')).toBeInTheDocument();
  });
});

describe('FormProgress', () => {
  const steps = ['Personal Info', 'Contact Details', 'Review'];

  it('renders form progress', () => {
    render(<FormProgress steps={steps} currentStep={1} />);
    
    expect(screen.getByText('Personal Info')).toBeInTheDocument();
    expect(screen.getByText('Contact Details')).toBeInTheDocument();
    expect(screen.getByText('Review')).toBeInTheDocument();
  });

  it('shows errors for specific steps', () => {
    const errors = { 1: true };
    
    render(<FormProgress steps={steps} currentStep={1} errors={errors} />);
    
    // Component should render without error
    expect(screen.getByText('Personal Info')).toBeInTheDocument();
  });

  it('calculates progress correctly', () => {
    render(<FormProgress steps={steps} currentStep={1} />);
    
    expect(screen.getByText('Progress')).toBeInTheDocument();
    expect(screen.getByText('50%')).toBeInTheDocument();
  });
});