/**
 * Formats a date string into a human-readable format
 * @param dateString ISO date string
 * @returns Formatted date string
 */
export function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date);
  } catch (error) {
    return dateString;
  }
}

/**
 * Returns a relative time string (e.g., "2 hours ago")
 * @param dateString ISO date string
 * @returns Relative time string
 */
export function getRelativeTimeString(dateString: string): string {
  try {
    const date = new Date(dateString);
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);
    
    // Less than a minute
    if (diffInSeconds < 60) {
      return 'just now';
    }
    
    // Less than an hour
    if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60);
      return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    }
    
    // Less than a day
    if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600);
      return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    }
    
    // Less than a week
    if (diffInSeconds < 604800) {
      const days = Math.floor(diffInSeconds / 86400);
      return `${days} day${days > 1 ? 's' : ''} ago`;
    }
    
    // Default to formatted date
    return formatDate(dateString);
  } catch (error) {
    return dateString;
  }
}

/**
 * Checks if a date is in the past
 * @param dateString ISO date string
 * @returns True if the date is in the past
 */
export function isDateInPast(dateString: string): boolean {
  try {
    const date = new Date(dateString);
    const now = new Date();
    return date < now;
  } catch (error) {
    return false;
  }
}

/**
 * Formats a date as YYYY-MM-DD
 * @param date Date object or ISO date string
 * @returns Formatted date string
 */
export function formatDateYYYYMMDD(date: Date | string): string {
  try {
    const d = typeof date === 'string' ? new Date(date) : date;
    return d.toISOString().split('T')[0];
  } catch (error) {
    return '';
  }
}