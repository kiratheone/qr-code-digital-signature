/**
 * @jest-environment jsdom
 */

import { performanceMonitor, usePerformanceMonitor } from '../performance';
import { renderHook } from '@testing-library/react';

// Mock performance API
const mockPerformance = {
  now: jest.fn(() => 1000),
  getEntriesByType: jest.fn(() => []),
};

// Mock PerformanceObserver
const mockPerformanceObserver = jest.fn();
mockPerformanceObserver.prototype.observe = jest.fn();
mockPerformanceObserver.prototype.disconnect = jest.fn();

// Setup global mocks
Object.defineProperty(global, 'performance', {
  value: mockPerformance,
  writable: true,
});

Object.defineProperty(global, 'PerformanceObserver', {
  value: mockPerformanceObserver,
  writable: true,
});

Object.defineProperty(global, 'navigator', {
  value: {
    userAgent: 'test-user-agent',
  },
  writable: true,
});

Object.defineProperty(global, 'window', {
  value: {
    location: {
      href: 'http://localhost:3000/test',
    },
  },
  writable: true,
});

describe('Performance Monitor', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    performanceMonitor.clear();
    mockPerformance.now.mockReturnValue(1000);
  });

  describe('measurePageLoad', () => {
    it('should measure page load performance', () => {
      const mockNavigation = {
        fetchStart: 100,
        loadEventEnd: 500,
        domContentLoadedEventStart: 200,
        domContentLoadedEventEnd: 300,
      };

      mockPerformance.getEntriesByType.mockReturnValue([mockNavigation]);

      performanceMonitor.measurePageLoad('/test-page');

      const metrics = performanceMonitor.getMetrics();
      const pageMetrics = metrics.get('/test-page');

      expect(pageMetrics).toBeDefined();
      expect(pageMetrics?.loadTime).toBe(400); // 500 - 100
      expect(pageMetrics?.renderTime).toBe(100); // 300 - 200
    });

    it('should handle missing navigation timing', () => {
      mockPerformance.getEntriesByType.mockReturnValue([]);

      performanceMonitor.measurePageLoad('/test-page');

      const metrics = performanceMonitor.getMetrics();
      expect(metrics.size).toBe(0);
    });
  });

  describe('measureRender', () => {
    it('should measure render time', () => {
      let callCount = 0;
      mockPerformance.now.mockImplementation(() => {
        callCount++;
        return callCount === 1 ? 1000 : 1050; // 50ms difference
      });

      const result = performanceMonitor.measureRender('TestComponent', () => {
        return 'rendered';
      });

      expect(result).toBe('rendered');

      const metrics = performanceMonitor.getMetrics();
      const componentMetrics = metrics.get('TestComponent');

      expect(componentMetrics).toBeDefined();
      expect(componentMetrics?.renderTime).toBe(50);
    });

    it('should handle render function errors', () => {
      const errorFn = () => {
        throw new Error('Render error');
      };

      expect(() => {
        performanceMonitor.measureRender('ErrorComponent', errorFn);
      }).toThrow('Render error');
    });
  });

  describe('measureInteraction', () => {
    it('should measure synchronous interaction time', () => {
      let callCount = 0;
      mockPerformance.now.mockImplementation(() => {
        callCount++;
        return callCount === 1 ? 1000 : 1100; // 100ms difference
      });

      performanceMonitor.measureInteraction('click-button', () => {
        // Simulate some work
      });

      const metrics = performanceMonitor.getMetrics();
      const interactionMetrics = metrics.get('click-button');

      expect(interactionMetrics).toBeDefined();
      expect(interactionMetrics?.interactionTime).toBe(100);
    });

    it('should measure asynchronous interaction time', async () => {
      let callCount = 0;
      mockPerformance.now.mockImplementation(() => {
        callCount++;
        return callCount === 1 ? 1000 : 1200; // 200ms difference
      });

      const asyncAction = async () => {
        return new Promise(resolve => setTimeout(resolve, 10));
      };

      await performanceMonitor.measureInteraction('async-action', asyncAction);

      const metrics = performanceMonitor.getMetrics();
      const interactionMetrics = metrics.get('async-action');

      expect(interactionMetrics).toBeDefined();
      expect(interactionMetrics?.interactionTime).toBe(200);
    });

    it('should handle async interaction errors', async () => {
      const errorAsyncFn = async () => {
        throw new Error('Async error');
      };

      await expect(
        performanceMonitor.measureInteraction('error-async', errorAsyncFn)
      ).rejects.toThrow('Async error');
    });
  });

  describe('initWebVitals', () => {
    it('should initialize web vitals monitoring', () => {
      performanceMonitor.initWebVitals();

      // Should have attempted to create PerformanceObserver instances
      expect(mockPerformanceObserver).toHaveBeenCalled();
    });

    it('should handle PerformanceObserver errors gracefully', () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      // Mock PerformanceObserver to throw
      mockPerformanceObserver.mockImplementation(() => {
        throw new Error('Observer not supported');
      });

      expect(() => {
        performanceMonitor.initWebVitals();
      }).not.toThrow();

      consoleSpy.mockRestore();
    });
  });

  describe('getMetrics and getWebVitals', () => {
    it('should return current metrics', () => {
      performanceMonitor.measureRender('TestComponent', () => 'test');

      const metrics = performanceMonitor.getMetrics();
      expect(metrics).toBeInstanceOf(Map);
      expect(metrics.has('TestComponent')).toBe(true);
    });

    it('should return current web vitals', () => {
      const webVitals = performanceMonitor.getWebVitals();
      expect(typeof webVitals).toBe('object');
    });
  });

  describe('logSummary', () => {
    it('should log performance summary', () => {
      const consoleSpy = jest.spyOn(console, 'group').mockImplementation();
      const consoleLogSpy = jest.spyOn(console, 'log').mockImplementation();
      const consoleGroupEndSpy = jest.spyOn(console, 'groupEnd').mockImplementation();

      performanceMonitor.measureRender('TestComponent', () => 'test');
      performanceMonitor.logSummary();

      expect(consoleSpy).toHaveBeenCalledWith('Performance Summary');
      expect(consoleLogSpy).toHaveBeenCalled();
      expect(consoleGroupEndSpy).toHaveBeenCalled();

      consoleSpy.mockRestore();
      consoleLogSpy.mockRestore();
      consoleGroupEndSpy.mockRestore();
    });
  });

  describe('sendToAnalytics', () => {
    it('should prepare analytics data', () => {
      const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

      performanceMonitor.measureRender('TestComponent', () => 'test');
      performanceMonitor.sendToAnalytics();

      expect(consoleSpy).toHaveBeenCalledWith('Analytics data:', expect.any(Object));

      consoleSpy.mockRestore();
    });
  });

  describe('clear', () => {
    it('should clear all metrics and web vitals', () => {
      performanceMonitor.measureRender('TestComponent', () => 'test');
      
      let metrics = performanceMonitor.getMetrics();
      expect(metrics.size).toBeGreaterThan(0);

      performanceMonitor.clear();

      metrics = performanceMonitor.getMetrics();
      const webVitals = performanceMonitor.getWebVitals();

      expect(metrics.size).toBe(0);
      expect(Object.keys(webVitals)).toHaveLength(0);
    });
  });
});

describe('usePerformanceMonitor hook', () => {
  it('should return performance monitoring functions', () => {
    const { result } = renderHook(() => usePerformanceMonitor());

    expect(result.current).toHaveProperty('measureRender');
    expect(result.current).toHaveProperty('measureInteraction');
    expect(result.current).toHaveProperty('measurePageLoad');
    expect(result.current).toHaveProperty('getMetrics');
    expect(result.current).toHaveProperty('getWebVitals');
    expect(result.current).toHaveProperty('logSummary');

    expect(typeof result.current.measureRender).toBe('function');
    expect(typeof result.current.measureInteraction).toBe('function');
    expect(typeof result.current.measurePageLoad).toBe('function');
    expect(typeof result.current.getMetrics).toBe('function');
    expect(typeof result.current.getWebVitals).toBe('function');
    expect(typeof result.current.logSummary).toBe('function');
  });

  it('should work with hook functions', () => {
    const { result } = renderHook(() => usePerformanceMonitor());

    const testResult = result.current.measureRender('HookTest', () => 'hook-result');
    expect(testResult).toBe('hook-result');

    const metrics = result.current.getMetrics();
    expect(metrics.has('HookTest')).toBe(true);
  });
});