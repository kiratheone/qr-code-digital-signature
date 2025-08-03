/**
 * Hooks exports
 * Provides clean imports for all custom hooks
 */

// Document operations hooks
export {
  useDocumentOperations,
  useDocument,
  useDocumentDownload,
  useFileValidation,
  documentKeys,
} from './useDocumentOperations';

// Authentication hooks
export {
  useAuthOperations,
  useAuthValidation,
  useRequireAuth,
  authKeys,
} from './useAuthOperations';

// Verification hooks
export {
  useVerificationOperations,
  useQRCodeOperations,
  useVerificationDisplay,
  useVerificationFlow,
  verificationKeys,
} from './useVerificationOperations';