/**
 * Type definitions exports
 * Provides clean imports for all type definitions
 */

// Document types
export type {
  Document,
  SignDocumentRequest,
  SignDocumentResponse,
  DocumentList,
} from './document';

// Authentication types
export type {
  User,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  AuthState,
} from './auth';

// Verification types
export type {
  VerificationInfo,
  VerifyDocumentRequest,
  VerificationResult,
  VerificationStatus,
} from './verification';