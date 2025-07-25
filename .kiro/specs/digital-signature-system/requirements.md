# Requirements Document

## Introduction

A QR Code-based digital signature system for PDF documents that enables automatic validation of document authenticity. This system replaces manual signatures with secure digital signature technology, using document hashes and QR Codes for verification. The application consists of a Next.js frontend, Golang backend, PostgreSQL database, and can be deployed using Docker Compose.

## Requirements

### Requirement 1

**User Story:** As a system user, I want to digitally sign PDF documents, so that documents have verifiable validity and authenticity.

#### Acceptance Criteria

1. WHEN user uploads PDF document THEN system SHALL generate SHA-256 hash of the document
2. WHEN document hash has been created THEN system SHALL sign the hash using private key
3. WHEN digital signature has been created THEN system SHALL store signature and document metadata to database
4. WHEN signature is stored THEN system SHALL generate QR Code containing doc_id, hash, and signature
5. IF document is successfully signed THEN system SHALL display success confirmation to user

### Requirement 2

**User Story:** As a system user, I want QR Code to be automatically injected into PDF documents, so that final documents can be distributed with integrated validation.

#### Acceptance Criteria

1. WHEN QR Code has been created THEN system SHALL inject QR Code to the last page of PDF by default
2. WHEN user selects custom position THEN system SHALL be able to inject QR Code at specified position
3. WHEN QR Code has been injected THEN system SHALL generate final PDF that can be downloaded
4. WHEN final PDF is created THEN system SHALL maintain original document quality and format
5. IF QR injection process fails THEN system SHALL display clear error message

### Requirement 3

**User Story:** As a system administrator, I want to manage signed documents, so that I can track and organize documents efficiently.

#### Acceptance Criteria

1. WHEN user accesses management page THEN system SHALL display list of all signed documents
2. WHEN documents are displayed THEN system SHALL show metadata: doc_id, filename, issuer, creation time, document hash, and signature
3. WHEN user searches for documents THEN system SHALL provide search functionality based on filename or issuer
4. WHEN user selects a document THEN system SHALL display complete document details
5. IF document needs to be deleted THEN system SHALL provide deletion function with confirmation

### Requirement 4

**User Story:** As a document verifier, I want to verify document authenticity through QR Code, so that I can ensure documents have not been altered or forged.

#### Acceptance Criteria

1. WHEN QR Code is scanned THEN system SHALL open verification page with document metadata
2. WHEN verification page opens THEN system SHALL display information: document title, issuer, and creation date
3. WHEN verification page is displayed THEN system SHALL provide "Upload Document for Verification" button
4. WHEN document is uploaded for verification THEN system SHALL calculate hash of uploaded document
5. WHEN hash is calculated THEN system SHALL compare with original hash in database
6. WHEN comparison is complete THEN system SHALL validate digital signature using public key

### Requirement 5

**User Story:** As a document verifier, I want to receive clear and accurate validation results, so that I can make decisions based on document authenticity status.

#### Acceptance Criteria

1. WHEN validation completes AND hash matches AND signature is valid THEN system SHALL display status "✅ Document is valid"
2. WHEN validation completes AND QR is valid AND hash does not match THEN system SHALL display status "⚠️ QR valid, but file content has changed"
3. WHEN validation completes AND signature is invalid THEN system SHALL display status "❌ QR invalid / signature incorrect"
4. WHEN validation results are displayed THEN system SHALL provide detailed explanation about validation status
5. IF error occurs during validation process THEN system SHALL display informative error message

### Requirement 6

**User Story:** As a system administrator, I want the system to have high security, so that private keys and sensitive data are protected from unauthorized access.

#### Acceptance Criteria

1. WHEN system runs THEN private key SHALL be stored in secure environment variables
2. WHEN documents are processed THEN system SHALL be able to operate without storing original documents (only hash)
3. WHEN signature is created THEN process SHALL only be performed by authorized users on official server
4. WHEN access to private key is required THEN system SHALL use strong authentication mechanisms
5. IF unauthorized access attempts occur THEN system SHALL log and block such access
6. WHEN user authentication is required THEN system SHALL use secure session management

### Requirement 7

**User Story:** As a system user, I want the application to have optimal performance and user-friendly interface, so that it can be used easily and efficiently.

#### Acceptance Criteria

1. WHEN user uploads document THEN system SHALL process within maximum 30 seconds for documents up to 50MB
2. WHEN page loads THEN system SHALL display loading indicator during processing
3. WHEN user interacts with UI THEN system SHALL provide clear visual feedback
4. WHEN system displays document list THEN system SHALL use pagination for optimal performance
5. IF error occurs THEN system SHALL display user-friendly error messages with solution suggestions

### Requirement 8

**User Story:** As a system user, I want to authenticate with the system, so that only authorized users can sign documents and access management features.

#### Acceptance Criteria

1. WHEN user accesses document signing page THEN system SHALL require authentication
2. WHEN user provides valid credentials THEN system SHALL grant access to signing functionality
3. WHEN user session expires THEN system SHALL redirect to login page
4. WHEN user logs out THEN system SHALL invalidate session and redirect to login
5. IF invalid credentials are provided THEN system SHALL display clear error message
6. WHEN user accesses management features THEN system SHALL verify user authorization

### Requirement 9

**User Story:** As a developer, I want the system to be easily deployable using Docker, so that deployment and maintenance become more efficient.

#### Acceptance Criteria

1. WHEN system is deployed THEN docker-compose SHALL be able to run entire application stack
2. WHEN containers are running THEN Next.js frontend SHALL be able to communicate with Golang backend
3. WHEN application runs THEN PostgreSQL database SHALL connect with backend
4. WHEN environment variables are set THEN system SHALL be able to access required configurations
5. IF deployment issues occur THEN system SHALL provide informative error logs