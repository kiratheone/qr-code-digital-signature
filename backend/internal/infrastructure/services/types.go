package services

// AuditSeverity represents the severity level of audit events
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

// AuditEventType represents different types of audit events
type AuditEventType string

const (
	// Authentication events
	AuditEventLogin          AuditEventType = "auth.login"
	AuditEventLogout         AuditEventType = "auth.logout"
	AuditEventLoginFailed    AuditEventType = "auth.login_failed"
	AuditEventSessionExpired AuditEventType = "auth.session_expired"
	AuditEventTokenRefresh   AuditEventType = "auth.token_refresh"

	// Document events
	AuditEventDocumentSign   AuditEventType = "document.sign"
	AuditEventDocumentView   AuditEventType = "document.view"
	AuditEventDocumentDelete AuditEventType = "document.delete"
	AuditEventDocumentList   AuditEventType = "document.list"

	// Verification events
	AuditEventVerificationAttempt AuditEventType = "verification.attempt"
	AuditEventVerificationSuccess AuditEventType = "verification.success"
	AuditEventVerificationFailed  AuditEventType = "verification.failed"
	AuditEventQRCodeScan         AuditEventType = "verification.qr_scan"

	// System events
	AuditEventSystemStart    AuditEventType = "system.start"
	AuditEventSystemStop     AuditEventType = "system.stop"
	AuditEventConfigChange   AuditEventType = "system.config_change"
	AuditEventKeyRotation    AuditEventType = "system.key_rotation"
	AuditEventSecurityAlert  AuditEventType = "system.security_alert"

	// Admin events
	AuditEventUserCreate AuditEventType = "admin.user_create"
	AuditEventUserUpdate AuditEventType = "admin.user_update"
	AuditEventUserDelete AuditEventType = "admin.user_delete"
)