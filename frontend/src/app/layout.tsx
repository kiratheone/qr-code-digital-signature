import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import ReactQueryProvider from '@/providers/ReactQueryProvider'
import { ErrorBoundary } from '@/components/ErrorBoundary/ErrorBoundary'
import { NotificationProvider } from '@/components/UI/Notifications'
import PerformanceMonitor from '@/components/Performance/PerformanceMonitor'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Digital Signature System',
  description: 'QR Code-based digital signature system for PDF documents',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="min-h-screen bg-gray-50">
          <ErrorBoundary
            onError={(error, errorInfo) => {
              // Log error to monitoring service in production
              if (process.env.NODE_ENV === 'production') {
                console.error('Global error boundary caught:', error, errorInfo);
                // TODO: Send to error reporting service
              }
            }}
          >
            <NotificationProvider>
              <ReactQueryProvider>
                <PerformanceMonitor>
                  {children}
                </PerformanceMonitor>
              </ReactQueryProvider>
            </NotificationProvider>
          </ErrorBoundary>
        </div>
      </body>
    </html>
  )
}