// Performance monitoring utilities for frontend

export interface PerformanceMetrics {
  loadTime: number;
  renderTime: number;
  interactionTime: number;
  bundleSize?: number;
}

export interface WebVitals {
  FCP?: number; // First Contentful Paint
  LCP?: number; // Largest Contentful Paint
  FID?: number; // First Input Delay
  CLS?: number; // Cumulative Layout Shift
  TTFB?: number; // Time to First Byte
}

// Proper interfaces for performance entries
interface FirstInputPerformanceEntry extends PerformanceEntry {
  processingStart: number;
}

interface LayoutShiftPerformanceEntry extends PerformanceEntry {
  hadRecentInput: boolean;
  value: number;
}

class PerformanceMonitor {
  private metrics: Map<string, PerformanceMetrics> = new Map();
  private webVitals: WebVitals = {};

  // Measure page load performance
  measurePageLoad(pageName: string): void {
    if (typeof window === 'undefined') return;

    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    
    if (navigation) {
      const loadTime = navigation.loadEventEnd - navigation.fetchStart;
      const renderTime = navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart;
      
      this.metrics.set(pageName, {
        loadTime,
        renderTime,
        interactionTime: 0,
      });
    }
  }

  // Measure component render time
  measureRender<T>(componentName: string, renderFn: () => T): T {
    const startTime = performance.now();
    const result = renderFn();
    const endTime = performance.now();
    
    const existing = this.metrics.get(componentName) || { loadTime: 0, renderTime: 0, interactionTime: 0 };
    existing.renderTime = endTime - startTime;
    this.metrics.set(componentName, existing);
    
    return result;
  }

  // Measure user interaction time
  measureInteraction(actionName: string, actionFn: () => Promise<void> | void): Promise<void> | void {
    const startTime = performance.now();
    
    const result = actionFn();
    
    if (result instanceof Promise) {
      return result.finally(() => {
        const endTime = performance.now();
        const existing = this.metrics.get(actionName) || { loadTime: 0, renderTime: 0, interactionTime: 0 };
        existing.interactionTime = endTime - startTime;
        this.metrics.set(actionName, existing);
      });
    } else {
      const endTime = performance.now();
      const existing = this.metrics.get(actionName) || { loadTime: 0, renderTime: 0, interactionTime: 0 };
      existing.interactionTime = endTime - startTime;
      this.metrics.set(actionName, existing);
    }
  }

  // Initialize Web Vitals monitoring
  initWebVitals(): void {
    if (typeof window === 'undefined') return;

    // First Contentful Paint
    this.observePerformanceEntry('paint', (entry) => {
      if (entry.name === 'first-contentful-paint') {
        this.webVitals.FCP = entry.startTime;
      }
    });

    // Largest Contentful Paint
    this.observePerformanceEntry('largest-contentful-paint', (entry) => {
      this.webVitals.LCP = entry.startTime;
    });

    // First Input Delay
    this.observePerformanceEntry('first-input', (entry) => {
      const firstInputEntry = entry as FirstInputPerformanceEntry;
      this.webVitals.FID = firstInputEntry.processingStart - firstInputEntry.startTime;
    });

    // Cumulative Layout Shift
    this.observePerformanceEntry('layout-shift', (entry) => {
      const layoutShiftEntry = entry as LayoutShiftPerformanceEntry;
      if (!layoutShiftEntry.hadRecentInput) {
        this.webVitals.CLS = (this.webVitals.CLS || 0) + layoutShiftEntry.value;
      }
    });

    // Time to First Byte
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    if (navigation) {
      this.webVitals.TTFB = navigation.responseStart - navigation.requestStart;
    }
  }

  private observePerformanceEntry(type: string, callback: (entry: PerformanceEntry) => void): void {
    try {
      const observer = new PerformanceObserver((list) => {
        list.getEntries().forEach(callback);
      });
      observer.observe({ type, buffered: true });
    } catch (error) {
      console.warn(`Performance observer for ${type} not supported:`, error);
    }
  }

  // Get all metrics
  getMetrics(): Map<string, PerformanceMetrics> {
    return new Map(this.metrics);
  }

  // Get Web Vitals
  getWebVitals(): WebVitals {
    return { ...this.webVitals };
  }

  // Log performance summary
  logSummary(): void {
    console.group('Performance Summary');
    
    console.log('Web Vitals:', this.webVitals);
    
    console.log('Component Metrics:');
    this.metrics.forEach((metrics, name) => {
      console.log(`${name}:`, {
        loadTime: `${metrics.loadTime.toFixed(2)}ms`,
        renderTime: `${metrics.renderTime.toFixed(2)}ms`,
        interactionTime: `${metrics.interactionTime.toFixed(2)}ms`,
      });
    });
    
    console.groupEnd();
  }

  // Send metrics to analytics (placeholder)
  sendToAnalytics(): void {
    const data = {
      webVitals: this.webVitals,
      metrics: Object.fromEntries(this.metrics),
      timestamp: Date.now(),
      userAgent: navigator.userAgent,
      url: window.location.href,
    };

    // In a real application, you would send this to your analytics service
    console.log('Analytics data:', data);
    
    // Example: send to your backend
    // fetch('/api/analytics/performance', {
    //   method: 'POST',
    //   headers: { 'Content-Type': 'application/json' },
    //   body: JSON.stringify(data),
    // }).catch(console.error);
  }

  // Clear all metrics
  clear(): void {
    this.metrics.clear();
    this.webVitals = {};
  }
}

// Create singleton instance
export const performanceMonitor = new PerformanceMonitor();

// React hook for performance monitoring
export function usePerformanceMonitor() {
  return {
    measureRender: performanceMonitor.measureRender.bind(performanceMonitor),
    measureInteraction: performanceMonitor.measureInteraction.bind(performanceMonitor),
    measurePageLoad: performanceMonitor.measurePageLoad.bind(performanceMonitor),
    getMetrics: performanceMonitor.getMetrics.bind(performanceMonitor),
    getWebVitals: performanceMonitor.getWebVitals.bind(performanceMonitor),
    logSummary: performanceMonitor.logSummary.bind(performanceMonitor),
  };
}

// Initialize Web Vitals monitoring on import
if (typeof window !== 'undefined') {
  performanceMonitor.initWebVitals();
}