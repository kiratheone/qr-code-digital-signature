'use client';

import React, { useState, useEffect } from 'react';
import { LoadingSpinner } from './LoadingSpinner';

export interface ProgressStep {
  id: string;
  label: string;
  description?: string;
  status: 'pending' | 'active' | 'completed' | 'error';
}

interface ProgressIndicatorProps {
  steps: ProgressStep[];
  currentStep?: string;
  className?: string;
  variant?: 'horizontal' | 'vertical';
  showLabels?: boolean;
  showDescriptions?: boolean;
}

export function ProgressIndicator({
  steps,
  currentStep,
  className = '',
  variant = 'horizontal',
  showLabels = true,
  showDescriptions = false,
}: ProgressIndicatorProps) {
  const getStepIcon = (step: ProgressStep) => {
    switch (step.status) {
      case 'completed':
        return (
          <div className="flex items-center justify-center w-8 h-8 bg-green-500 rounded-full">
            <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'active':
        return (
          <div className="flex items-center justify-center w-8 h-8 bg-blue-500 rounded-full">
            <LoadingSpinner size="sm" color="white" />
          </div>
        );
      case 'error':
        return (
          <div className="flex items-center justify-center w-8 h-8 bg-red-500 rounded-full">
            <svg className="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'pending':
      default:
        return (
          <div className="flex items-center justify-center w-8 h-8 bg-gray-300 rounded-full">
            <div className="w-3 h-3 bg-gray-500 rounded-full" />
          </div>
        );
    }
  };

  const getStepTextColor = (step: ProgressStep) => {
    switch (step.status) {
      case 'completed':
        return 'text-green-600';
      case 'active':
        return 'text-blue-600';
      case 'error':
        return 'text-red-600';
      case 'pending':
      default:
        return 'text-gray-500';
    }
  };

  const getConnectorColor = (index: number) => {
    if (index >= steps.length - 1) return '';
    
    const currentStepStatus = steps[index].status;
    const nextStepStatus = steps[index + 1].status;
    
    if (currentStepStatus === 'completed') {
      return 'bg-green-500';
    } else if (currentStepStatus === 'active' || currentStepStatus === 'error') {
      return 'bg-gray-300';
    }
    
    return 'bg-gray-300';
  };

  if (variant === 'vertical') {
    return (
      <div className={`space-y-4 ${className}`}>
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-start">
            <div className="flex flex-col items-center">
              {getStepIcon(step)}
              {index < steps.length - 1 && (
                <div className={`w-0.5 h-8 mt-2 ${getConnectorColor(index)}`} />
              )}
            </div>
            
            {showLabels && (
              <div className="ml-4 flex-1">
                <div className={`text-sm font-medium ${getStepTextColor(step)}`}>
                  {step.label}
                </div>
                {showDescriptions && step.description && (
                  <div className="text-xs text-gray-500 mt-1">
                    {step.description}
                  </div>
                )}
              </div>
            )}
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className={`flex items-center ${className}`}>
      {steps.map((step, index) => (
        <React.Fragment key={step.id}>
          <div className="flex flex-col items-center">
            {getStepIcon(step)}
            
            {showLabels && (
              <div className="mt-2 text-center">
                <div className={`text-xs font-medium ${getStepTextColor(step)}`}>
                  {step.label}
                </div>
                {showDescriptions && step.description && (
                  <div className="text-xs text-gray-500 mt-1 max-w-20">
                    {step.description}
                  </div>
                )}
              </div>
            )}
          </div>
          
          {index < steps.length - 1 && (
            <div className={`flex-1 h-0.5 mx-4 ${getConnectorColor(index)}`} />
          )}
        </React.Fragment>
      ))}
    </div>
  );
}

// Hook for managing progress steps
export function useProgressSteps(initialSteps: ProgressStep[]) {
  const [steps, setSteps] = useState<ProgressStep[]>(initialSteps);
  const [currentStepId, setCurrentStepId] = useState<string | null>(null);

  const updateStep = (stepId: string, updates: Partial<ProgressStep>) => {
    setSteps(prev => prev.map(step => 
      step.id === stepId ? { ...step, ...updates } : step
    ));
  };

  const setStepStatus = (stepId: string, status: ProgressStep['status']) => {
    updateStep(stepId, { status });
  };

  const nextStep = () => {
    const currentIndex = steps.findIndex(step => step.status === 'active');
    if (currentIndex >= 0 && currentIndex < steps.length - 1) {
      setSteps(prev => prev.map((step, index) => {
        if (index === currentIndex) {
          return { ...step, status: 'completed' };
        } else if (index === currentIndex + 1) {
          return { ...step, status: 'active' };
        }
        return step;
      }));
    }
  };

  const previousStep = () => {
    const currentIndex = steps.findIndex(step => step.status === 'active');
    if (currentIndex > 0) {
      setSteps(prev => prev.map((step, index) => {
        if (index === currentIndex) {
          return { ...step, status: 'pending' };
        } else if (index === currentIndex - 1) {
          return { ...step, status: 'active' };
        }
        return step;
      }));
    }
  };

  const resetSteps = () => {
    setSteps(prev => prev.map((step, index) => ({
      ...step,
      status: index === 0 ? 'active' : 'pending'
    })));
  };

  const completeAllSteps = () => {
    setSteps(prev => prev.map(step => ({ ...step, status: 'completed' })));
  };

  const getCurrentStep = () => {
    return steps.find(step => step.status === 'active') || null;
  };

  const getProgress = () => {
    const completedSteps = steps.filter(step => step.status === 'completed').length;
    return (completedSteps / steps.length) * 100;
  };

  return {
    steps,
    currentStepId,
    updateStep,
    setStepStatus,
    nextStep,
    previousStep,
    resetSteps,
    completeAllSteps,
    getCurrentStep,
    getProgress,
  };
}

// Animated progress bar component
interface AnimatedProgressBarProps {
  progress: number;
  duration?: number;
  color?: 'blue' | 'green' | 'red' | 'yellow';
  showPercentage?: boolean;
  className?: string;
}

export function AnimatedProgressBar({
  progress,
  duration = 1000,
  color = 'blue',
  showPercentage = false,
  className = '',
}: AnimatedProgressBarProps) {
  const [animatedProgress, setAnimatedProgress] = useState(0);

  useEffect(() => {
    const startTime = Date.now();
    const startProgress = animatedProgress;
    const progressDiff = progress - startProgress;

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progressRatio = Math.min(elapsed / duration, 1);
      
      // Easing function (ease-out)
      const easedProgress = 1 - Math.pow(1 - progressRatio, 3);
      const currentProgress = startProgress + (progressDiff * easedProgress);
      
      setAnimatedProgress(currentProgress);
      
      if (progressRatio < 1) {
        requestAnimationFrame(animate);
      }
    };

    requestAnimationFrame(animate);
  }, [progress, duration, animatedProgress]);

  const colorClasses = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    red: 'bg-red-500',
    yellow: 'bg-yellow-500',
  };

  return (
    <div className={`w-full ${className}`}>
      <div className="flex justify-between items-center mb-1">
        <span className="text-sm font-medium text-gray-700">Progress</span>
        {showPercentage && (
          <span className="text-sm text-gray-500">
            {Math.round(animatedProgress)}%
          </span>
        )}
      </div>
      <div className="w-full bg-gray-200 rounded-full h-2">
        <div
          className={`h-2 rounded-full transition-all duration-300 ${colorClasses[color]}`}
          style={{ width: `${Math.max(0, Math.min(100, animatedProgress))}%` }}
        />
      </div>
    </div>
  );
}

// Multi-step form progress component
interface FormProgressProps {
  steps: string[];
  currentStep: number;
  errors?: Record<number, boolean>;
  className?: string;
}

export function FormProgress({
  steps,
  currentStep,
  errors = {},
  className = '',
}: FormProgressProps) {
  const progressSteps: ProgressStep[] = steps.map((label, index) => ({
    id: `step-${index}`,
    label: `${index + 1}`,
    description: label,
    status: 
      errors[index] ? 'error' :
      index < currentStep ? 'completed' :
      index === currentStep ? 'active' :
      'pending'
  }));

  return (
    <div className={className}>
      <ProgressIndicator
        steps={progressSteps}
        variant="horizontal"
        showLabels={true}
        showDescriptions={true}
      />
      <div className="mt-4">
        <AnimatedProgressBar
          progress={(currentStep / (steps.length - 1)) * 100}
          showPercentage={true}
          color={errors[currentStep] ? 'red' : 'blue'}
        />
      </div>
    </div>
  );
}