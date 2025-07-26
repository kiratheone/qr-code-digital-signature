'use client';

import { useEffect, ReactNode } from 'react';
import { usePathname } from 'next/navigation';
import { performanceMonitor } from '@/utils/performance';

interface PerformanceMonitorProps {
  children: ReactNode;
}

export default function PerformanceMonitor({ children }: PerformanceMonitorProps) {
  const pathname = usePathname();

  useEffect(() => {
    // Measure page load performance
    performanceMonitor.measurePageLoad(pathname);

    // Log performance summary in development
    if (process.env.NODE_ENV === 'development') {
      const timer = setTimeout(() => {
        performanceMonitor.logSummary();
      }, 2000); // Wait 2 seconds for metrics to stabilize

      return () => clearTimeout(timer);
    }

    // Send metrics to analytics in production
    if (process.env.NODE_ENV === 'production') {
      const timer = setTimeout(() => {
        performanceMonitor.sendToAnalytics();
      }, 5000); // Wait 5 seconds for all metrics

      return () => clearTimeout(timer);
    }
  }, [pathname]);

  // Monitor route changes
  useEffect(() => {
    performance.now(); // Mark start time

    return () => {
      performance.now(); // Mark end time
      performanceMonitor.measureInteraction(`route-change-${pathname}`, () => {
        // Route change completed
      });
    };
  }, [pathname]);

  return <>{children}</>;
}