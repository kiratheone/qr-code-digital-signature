package di

import (
	"context"
	"digital-signature-system/internal/config"
	"digital-signature-system/internal/domain/repositories"
	"digital-signature-system/internal/domain/services"
	"digital-signature-system/internal/domain/usecases"
	infraServices "digital-signature-system/internal/infrastructure/services"
	"time"

	"gorm.io/gorm"
)

// auditServiceAdapter adapts the infrastructure AuditService to the domain interface
type auditServiceAdapter struct {
	service *infraServices.AuditService
}

// LogAuthEvent adapts the method signature from string to AuditEventType
func (a *auditServiceAdapter) LogAuthEvent(ctx context.Context, eventType string, userID, ip string, success bool, details map[string]interface{}) {
	a.service.LogAuthEvent(ctx, infraServices.AuditEventType(eventType), userID, ip, success, details)
}

// Container holds all dependencies
type Container struct {
	config *config.Config
	db     *gorm.DB

	// Repositories
	userRepo            repositories.UserRepository
	sessionRepo         repositories.SessionRepository
	documentRepo        repositories.DocumentRepository
	verificationLogRepo repositories.VerificationLogRepository

	// Services
	signatureService services.SignatureService
	hashService      services.HashService
	passwordService  services.PasswordService
	tokenService     services.TokenService
	keyService       services.KeyService
	pdfService       services.PDFService
	qrService        services.QRService

	// Infrastructure Services
	auditService      *infraServices.AuditService
	monitoringService *infraServices.MonitoringService
	cacheService      *infraServices.CacheService

	// Use Cases
	authUseCase         usecases.AuthUseCase
	documentUseCase     usecases.DocumentUseCase
	verificationUseCase usecases.VerificationUseCase
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	container := &Container{
		config: cfg,
		db:     db,
	}

	// Initialize repositories
	container.initRepositories()

	// Initialize services
	container.initServices()

	// Initialize use cases
	container.initUseCases()

	return container
}

// Initialize repositories
func (c *Container) initRepositories() {
	// TODO: Initialize repositories
}

// Initialize services
func (c *Container) initServices() {
	var err error
	c.signatureService, err = infraServices.NewSignatureService(c.config)
	if err != nil {
		panic("Failed to initialize signature service: " + err.Error())
	}
	c.hashService = infraServices.NewHashService()
	c.passwordService = infraServices.NewPasswordService()
	c.tokenService = infraServices.NewTokenService(c.config)
	c.keyService, _ = infraServices.NewKeyService(c.config)
	
	// Initialize PDF service with 50MB max file size
	c.pdfService = infraServices.NewPDFService(50 * 1024 * 1024) // 50MB
	
	// Initialize QR service with verification base URL from config
	verificationBaseURL := c.config.GetString("VERIFICATION_BASE_URL", "https://example.com/verify/")
	c.qrService = infraServices.NewQRService(verificationBaseURL)

	// Initialize audit service
	auditConfig := infraServices.AuditConfig{
		LogDir:        c.config.GetString("AUDIT_LOG_DIR", "./logs/audit"),
		MaxFileSize:   c.config.GetInt64("AUDIT_MAX_FILE_SIZE", 100*1024*1024), // 100MB
		MaxFiles:      c.config.GetInt("AUDIT_MAX_FILES", 30),
		RotationTime:  24 * time.Hour,
		BufferSize:    c.config.GetInt("AUDIT_BUFFER_SIZE", 1000),
		FlushInterval: c.config.GetDuration("AUDIT_FLUSH_INTERVAL", 5*time.Second),
	}
	
	c.auditService, err = infraServices.NewAuditService(auditConfig)
	if err != nil {
		// Log error but don't fail startup - audit is important but not critical
		// In production, you might want to fail startup if audit can't be initialized
		panic("Failed to initialize audit service: " + err.Error())
	}

	// Initialize monitoring service
	monitoringConfig := infraServices.MonitoringConfig{
		AuditService:           c.auditService,
		PerformanceInterval:    c.config.GetDuration("MONITORING_PERF_INTERVAL", 30*time.Second),
		AlertCheckInterval:     c.config.GetDuration("MONITORING_ALERT_INTERVAL", 10*time.Second),
		RateLimitWindow:        c.config.GetDuration("MONITORING_RATE_LIMIT_WINDOW", time.Minute),
		RateLimitThreshold:     c.config.GetInt("MONITORING_RATE_LIMIT_THRESHOLD", 100),
		RequestCountWindow:     c.config.GetDuration("MONITORING_REQUEST_WINDOW", time.Minute),
		AlertThresholds: map[string]float64{
			"cpu_usage":             c.config.GetFloat64("ALERT_CPU_THRESHOLD", 80.0),
			"memory_usage":          c.config.GetFloat64("ALERT_MEMORY_THRESHOLD", 85.0),
			"error_rate":            c.config.GetFloat64("ALERT_ERROR_RATE_THRESHOLD", 5.0),
			"response_time_p95":     c.config.GetFloat64("ALERT_RESPONSE_TIME_THRESHOLD", 2000.0),
			"failed_logins":         c.config.GetFloat64("ALERT_FAILED_LOGINS_THRESHOLD", 10.0),
			"verification_failures": c.config.GetFloat64("ALERT_VERIFICATION_FAILURES_THRESHOLD", 20.0),
		},
	}
	
	c.monitoringService = infraServices.NewMonitoringService(monitoringConfig)

	// Initialize cache service with 15 minute TTL
	c.cacheService = infraServices.NewCacheService(15 * time.Minute)
}

// Initialize use cases
func (c *Container) initUseCases() {
	// Create audit service adapter
	auditAdapter := &auditServiceAdapter{service: c.auditService}
	
	c.authUseCase = usecases.NewAuthUseCase(
		c.userRepo,
		c.sessionRepo,
		c.passwordService,
		c.tokenService,
		24*time.Hour, // Session duration
		auditAdapter,
		c.monitoringService,
	)
	
	// Initialize document use case
	c.documentUseCase = usecases.NewDocumentUseCase(
		c.documentRepo,
		c.signatureService,
		c.pdfService,
		c.qrService,
	)
	
	// Initialize verification use case
	c.verificationUseCase = usecases.NewVerificationUseCase(
		c.documentRepo,
		c.verificationLogRepo,
		c.signatureService,
		c.pdfService,
		c.qrService,
		c.auditService,
		c.monitoringService,
	)
}

// Config returns the application configuration
func (c *Container) Config() *config.Config {
	return c.config
}

// DB returns the database connection
func (c *Container) DB() *gorm.DB {
	return c.db
}

// UserRepository returns the user repository
func (c *Container) UserRepository() repositories.UserRepository {
	return c.userRepo
}

// SessionRepository returns the session repository
func (c *Container) SessionRepository() repositories.SessionRepository {
	return c.sessionRepo
}

// DocumentRepository returns the document repository
func (c *Container) DocumentRepository() repositories.DocumentRepository {
	return c.documentRepo
}

// VerificationLogRepository returns the verification log repository
func (c *Container) VerificationLogRepository() repositories.VerificationLogRepository {
	return c.verificationLogRepo
}

// SignatureService returns the signature service
func (c *Container) SignatureService() services.SignatureService {
	return c.signatureService
}

// HashService returns the hash service
func (c *Container) HashService() services.HashService {
	return c.hashService
}

// PasswordService returns the password service
func (c *Container) PasswordService() services.PasswordService {
	return c.passwordService
}

// TokenService returns the token service
func (c *Container) TokenService() services.TokenService {
	return c.tokenService
}

// KeyService returns the key service
func (c *Container) KeyService() services.KeyService {
	return c.keyService
}

// AuthUseCase returns the authentication use case
func (c *Container) AuthUseCase() usecases.AuthUseCase {
	return c.authUseCase
}
// PDFService returns the PDF service
func (c *Container) PDFService() services.PDFService {
	return c.pdfService
}

// QRService returns the QR code service
func (c *Container) QRService() services.QRService {
	return c.qrService
}

// DocumentUseCase returns the document use case
func (c *Container) DocumentUseCase() usecases.DocumentUseCase {
	return c.documentUseCase
}

// VerificationUseCase returns the verification use case
func (c *Container) VerificationUseCase() usecases.VerificationUseCase {
	return c.verificationUseCase
}

// AuditService returns the audit service
func (c *Container) AuditService() *infraServices.AuditService {
	return c.auditService
}

// MonitoringService returns the monitoring service
func (c *Container) MonitoringService() *infraServices.MonitoringService {
	return c.monitoringService
}

// CacheService returns the cache service
func (c *Container) CacheService() *infraServices.CacheService {
	return c.cacheService
}

// Close closes all services and connections
func (c *Container) Close() error {
	if c.auditService != nil {
		c.auditService.Close()
	}
	if c.monitoringService != nil {
		c.monitoringService.Close()
	}
	return nil
}