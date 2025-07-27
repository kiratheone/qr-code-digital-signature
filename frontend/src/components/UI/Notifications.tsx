'use client';

import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';

export type NotificationType = 'success' | 'error' | 'warning' | 'info';

export interface Notification {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  duration?: number;
  persistent?: boolean;
  action?: {
    label: string;
    onClick: () => void;
  };
}

interface NotificationContextType {
  notifications: Notification[];
  addNotification: (notification: Omit<Notification, 'id'>) => string;
  removeNotification: (id: string) => void;
  clearAllNotifications: () => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export function useNotifications() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
}

interface NotificationProviderProps {
  children: React.ReactNode;
  maxNotifications?: number;
}

export function NotificationProvider({ 
  children, 
  maxNotifications = 5 
}: NotificationProviderProps) {
  const [notifications, setNotifications] = useState<Notification[]>([]);


  // Define removeNotification first to avoid circular dependency
  const removeNotification = useCallback((id: string) => {
    setNotifications(prev => prev.filter(notification => notification.id !== id));
  }, []);

  const addNotification = useCallback((notification: Omit<Notification, 'id'>) => {
    const id = Math.random().toString(36).substr(2, 9);
    const newNotification: Notification = {
      ...notification,
      id,
      duration: notification.duration ?? 5000,
    };

    setNotifications(prev => {
      const updated = [newNotification, ...prev];
      return updated.slice(0, maxNotifications);
    });

    // Auto-remove notification after duration (unless persistent)
    if (!notification.persistent && newNotification.duration && newNotification.duration > 0) {
      setTimeout(() => {
        setNotifications(prev => prev.filter(notification => notification.id !== id));
      }, newNotification.duration);
    }

    return id;
  }, [maxNotifications]);

  const clearAllNotifications = useCallback(() => {
    setNotifications([]);
  }, []);

  return (
    <NotificationContext.Provider value={{
      notifications,
      addNotification,
      removeNotification,
      clearAllNotifications,
    }}>
      {children}
      <NotificationContainer />
    </NotificationContext.Provider>
  );
}

function NotificationContainer() {
  const { notifications } = useNotifications();

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2 max-w-sm w-full">
      {notifications.map(notification => (
        <NotificationItem key={notification.id} notification={notification} />
      ))}
    </div>
  );
}

interface NotificationItemProps {
  notification: Notification;
}

function NotificationItem({ notification }: NotificationItemProps) {
  const { removeNotification } = useNotifications();
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Trigger animatio
    const timer = setTimeout(() => setIsVisible(true), 10);
    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(() => removeNotification(notification.id), 300);
  };

  const getNotificationStyles = () => {
    const baseStyles = "rounded-lg shadow-lg p-4 border-l-4 transition-all duration-300 transform";
    
    switch (notification.type) {
      case 'success':
        return `${baseStyles} bg-green-50 border-green-400`;
      case 'error':
        return `${baseStyles} bg-red-50 border-red-400`;
      case 'warning':
        return `${baseStyles} bg-yellow-50 border-yellow-400`;
      case 'info':
        return `${baseStyles} bg-blue-50 border-blue-400`;
      default:
        return `${baseStyles} bg-gray-50 border-gray-400`;
    }
  };

  const getIconColor = () => {
    switch (notification.type) {
      case 'success':
        return 'text-green-400';
      case 'error':
        return 'text-red-400';
      case 'warning':
        return 'text-yellow-400';
      case 'info':
        return 'text-blue-400';
      default:
        return 'text-gray-400';
    }
  };

  const getIcon = () => {
    switch (notification.type) {
      case 'success':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
        );
      case 'error':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
          </svg>
        );
      case 'warning':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
        );
      case 'info':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        );
      default:
        return null;
    }
  };

  return (
    <div
      className={`${getNotificationStyles()} ${
        isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
      }`}
    >
      <div className="flex items-start">
        <div className={`flex-shrink-0 ${getIconColor()}`}>
          {getIcon()}
        </div>
        <div className="ml-3 flex-1">
          <h3 className={`text-sm font-medium ${
            notification.type === 'success' ? 'text-green-800' :
            notification.type === 'error' ? 'text-red-800' :
            notification.type === 'warning' ? 'text-yellow-800' :
            'text-blue-800'
          }`}>
            {notification.title}
          </h3>
          {notification.message && (
            <p className={`mt-1 text-sm ${
              notification.type === 'success' ? 'text-green-700' :
              notification.type === 'error' ? 'text-red-700' :
              notification.type === 'warning' ? 'text-yellow-700' :
              'text-blue-700'
            }`}>
              {notification.message}
            </p>
          )}
          {notification.action && (
            <div className="mt-2">
              <button
                onClick={notification.action.onClick}
                className={`text-sm font-medium underline ${
                  notification.type === 'success' ? 'text-green-800 hover:text-green-900' :
                  notification.type === 'error' ? 'text-red-800 hover:text-red-900' :
                  notification.type === 'warning' ? 'text-yellow-800 hover:text-yellow-900' :
                  'text-blue-800 hover:text-blue-900'
                }`}
              >
                {notification.action.label}
              </button>
            </div>
          )}
        </div>
        <div className="ml-4 flex-shrink-0">
          <button
            onClick={handleClose}
            className={`inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2 ${
              notification.type === 'success' ? 'text-green-500 hover:bg-green-100 focus:ring-green-600' :
              notification.type === 'error' ? 'text-red-500 hover:bg-red-100 focus:ring-red-600' :
              notification.type === 'warning' ? 'text-yellow-500 hover:bg-yellow-100 focus:ring-yellow-600' :
              'text-blue-500 hover:bg-blue-100 focus:ring-blue-600'
            }`}
          >
            <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}

// Convenience hooks for different notification types
export function useNotificationHelpers() {
  const { addNotification } = useNotifications();

  const showSuccess = useCallback((title: string, message?: string, options?: Partial<Notification>) => {
    return addNotification({ type: 'success', title, message, ...options });
  }, [addNotification]);

  const showError = useCallback((title: string, message?: string, options?: Partial<Notification>) => {
    return addNotification({ type: 'error', title, message, persistent: true, ...options });
  }, [addNotification]);

  const showWarning = useCallback((title: string, message?: string, options?: Partial<Notification>) => {
    return addNotification({ type: 'warning', title, message, ...options });
  }, [addNotification]);

  const showInfo = useCallback((title: string, message?: string, options?: Partial<Notification>) => {
    return addNotification({ type: 'info', title, message, ...options });
  }, [addNotification]);

  return {
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };
}