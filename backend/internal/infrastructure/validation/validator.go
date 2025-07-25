package validation

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Validator provides comprehensive input validation
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	validate := validator.New()
	
	// Register custom validators
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("filename", validateFilename)
	validate.RegisterValidation("issuer", validateIssuer)
	validate.RegisterValidation("uuid", validateUUID)
	validate.RegisterValidation("safe_string", validateSafeString)
	validate.RegisterValidation("no_sql_injection", validateNoSQLInjection)
	validate.RegisterValidation("no_xss", validateNoXSS)
	
	return &Validator{
		validate: validate,
	}
}

// ValidateStruct validates a struct using validation tags
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateFile validates uploaded files
func (v *Validator) ValidateFile(file *multipart.FileHeader, allowedTypes []string, maxSize int64) error {
	// Check file size
	if file.Size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", file.Size, maxSize)
	}
	
	// Check file type by extension
	if !v.isAllowedFileType(file.Filename, allowedTypes) {
		return fmt.Errorf("file type not allowed, supported types: %v", allowedTypes)
	}
	
	// Check MIME type
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()
	
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}
	
	mimeType := http.DetectContentType(buffer)
	if !v.isAllowedMimeType(mimeType, allowedTypes) {
		return fmt.Errorf("invalid file content type: %s", mimeType)
	}
	
	return nil
}

// ValidatePDFFile specifically validates PDF files
func (v *Validator) ValidatePDFFile(file *multipart.FileHeader, maxSize int64) error {
	return v.ValidateFile(file, []string{"pdf"}, maxSize)
}

// SanitizeString removes potentially dangerous characters from strings
func (v *Validator) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except newline, carriage return, and tab
	var result strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			continue
		}
		result.WriteRune(r)
	}
	
	// Remove potentially dangerous patterns
	dangerous := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"eval(", "expression(", "url(", "import(",
	}
	
	sanitized := result.String()
	for _, pattern := range dangerous {
		sanitized = strings.ReplaceAll(strings.ToLower(sanitized), pattern, "")
	}
	
	return strings.TrimSpace(sanitized)
}

// SanitizeFilename sanitizes filenames to prevent path traversal
func (v *Validator) SanitizeFilename(filename string) string {
	// Remove path separators and dangerous characters
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	sanitized := filename
	
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}
	
	// Limit length
	if len(sanitized) > 255 {
		sanitized = sanitized[:255]
	}
	
	return strings.TrimSpace(sanitized)
}

// ValidateAndSanitizeInput validates and sanitizes common input types
func (v *Validator) ValidateAndSanitizeInput(input string, inputType string) (string, error) {
	// First sanitize
	sanitized := v.SanitizeString(input)
	
	// Then validate based on type
	switch inputType {
	case "username":
		if err := validateUsername(validator.FieldLevel(nil)); err != nil {
			return "", fmt.Errorf("invalid username format")
		}
	case "email":
		if !isValidEmail(sanitized) {
			return "", fmt.Errorf("invalid email format")
		}
	case "filename":
		sanitized = v.SanitizeFilename(sanitized)
		if len(sanitized) == 0 {
			return "", fmt.Errorf("filename cannot be empty after sanitization")
		}
	case "issuer":
		if len(sanitized) < 2 || len(sanitized) > 100 {
			return "", fmt.Errorf("issuer must be between 2 and 100 characters")
		}
	}
	
	return sanitized, nil
}

// Helper functions for validation

func (v *Validator) isAllowedFileType(filename string, allowedTypes []string) bool {
	ext := strings.ToLower(strings.TrimPrefix(getFileExtension(filename), "."))
	for _, allowedType := range allowedTypes {
		if ext == strings.ToLower(allowedType) {
			return true
		}
	}
	return false
}

func (v *Validator) isAllowedMimeType(mimeType string, allowedTypes []string) bool {
	mimeMap := map[string]string{
		"pdf": "application/pdf",
		"jpg": "image/jpeg",
		"jpeg": "image/jpeg",
		"png": "image/png",
		"gif": "image/gif",
	}
	
	for _, allowedType := range allowedTypes {
		if expectedMime, exists := mimeMap[strings.ToLower(allowedType)]; exists {
			if strings.HasPrefix(mimeType, expectedMime) {
				return true
			}
		}
	}
	return false
}

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// Custom validation functions

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	
	// Only allow alphanumeric characters, underscores, and hyphens
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	return matched
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 || len(password) > 100 {
		return false
	}
	
	// Check for at least one uppercase, one lowercase, one digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	
	return hasUpper && hasLower && hasDigit
}

func validateFilename(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}
	
	// Check for dangerous characters
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range dangerous {
		if strings.Contains(filename, char) {
			return false
		}
	}
	
	return true
}

func validateIssuer(fl validator.FieldLevel) bool {
	issuer := fl.Field().String()
	return len(issuer) >= 2 && len(issuer) <= 100
}

func validateUUID(fl validator.FieldLevel) bool {
	uuid := fl.Field().String()
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}

func validateSafeString(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	
	// Check for dangerous patterns
	dangerous := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"eval(", "expression(", "url(", "import(",
	}
	
	lowerStr := strings.ToLower(str)
	for _, pattern := range dangerous {
		if strings.Contains(lowerStr, pattern) {
			return false
		}
	}
	
	return true
}

func validateNoSQLInjection(fl validator.FieldLevel) bool {
	str := strings.ToLower(fl.Field().String())
	
	// Common SQL injection patterns
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete", "drop",
		"create", "alter", "exec", "execute", "declare",
	}
	
	for _, pattern := range sqlPatterns {
		if strings.Contains(str, pattern) {
			return false
		}
	}
	
	return true
}

func validateNoXSS(fl validator.FieldLevel) bool {
	str := strings.ToLower(fl.Field().String())
	
	// Common XSS patterns
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"onload", "onerror", "onclick", "onmouseover", "onsubmit",
		"eval(", "expression(", "alert(", "confirm(", "prompt(",
	}
	
	for _, pattern := range xssPatterns {
		if strings.Contains(str, pattern) {
			return false
		}
	}
	
	return true
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateIPAddress validates IP addresses
func (v *Validator) ValidateIPAddress(ip string) bool {
	// Basic IP validation - could be enhanced with more sophisticated checks
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	return ipRegex.MatchString(ip)
}

// ValidateUserAgent validates user agent strings
func (v *Validator) ValidateUserAgent(userAgent string) bool {
	// Check for suspicious patterns in user agent
	suspiciousPatterns := []string{
		"<script", "javascript:", "data:", "vbscript:",
		"eval(", "expression(", "import(",
		"sqlmap", "nmap", "nikto", "burp",
	}
	
	lowerUA := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerUA, pattern) {
			return false
		}
	}
	
	// Check length
	if len(userAgent) > 512 {
		return false
	}
	
	return true
}

// ValidateReferer validates referer headers
func (v *Validator) ValidateReferer(referer string) bool {
	if referer == "" {
		return true // Empty referer is allowed
	}
	
	// Basic URL validation
	if len(referer) > 2048 {
		return false
	}
	
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"javascript:", "data:", "vbscript:",
		"<script", "eval(", "expression(",
	}
	
	lowerRef := strings.ToLower(referer)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerRef, pattern) {
			return false
		}
	}
	
	return true
}

// DetectSQLInjection detects potential SQL injection attempts
func (v *Validator) DetectSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)
	
	// Common SQL injection patterns
	sqlPatterns := []string{
		"'", "\"", ";", "--", "/*", "*/",
		" or ", " and ", " union ", " select ", " insert ", " update ", 
		" delete ", " drop ", " create ", " alter ", " exec ", " execute ",
		" declare ", " cast(", " convert(", " substring(", " ascii(",
		" char(", " nchar(", " varchar(", " nvarchar(", " waitfor ",
		" delay ", " benchmark(", " sleep(", " pg_sleep(",
		"xp_", "sp_", "@@", "information_schema", "sysobjects",
		// Advanced SQL injection patterns
		" having ", " group by ", " order by ", " limit ", " offset ",
		" into ", " from ", " where ", " join ", " inner ", " outer ",
		" left ", " right ", " full ", " cross ", " on ", " using ",
		" case ", " when ", " then ", " else ", " end ", " if(",
		" ifnull(", " isnull(", " coalesce(", " nullif(",
		" count(", " sum(", " avg(", " min(", " max(",
		" concat(", " substring(", " length(", " upper(", " lower(",
		" trim(", " ltrim(", " rtrim(", " replace(",
		" database(", " version(", " user(", " current_user",
		" session_user", " system_user", " schema(", " table_name",
		" column_name", " table_schema", " information_schema",
		" mysql.", " sys.", " performance_schema", " pg_catalog",
		" pg_user", " pg_shadow", " pg_group", " pg_database",
		" master.", " msdb.", " tempdb.", " model.",
		" sysdatabases", " syscolumns", " systables", " sysusers",
		" load_file(", " into outfile", " into dumpfile",
		" union all ", " union distinct ", " union select ",
		"0x", "char(", "chr(", "ascii(", "hex(",
		"md5(", "sha1(", "sha2(", "password(",
		"encode(", "decode(", "compress(", "uncompress(",
		"benchmark(", "sleep(", "pg_sleep(", "waitfor delay",
		"dbms_pipe.receive_message", "dbms_lock.sleep",
		"utl_inaddr.get_host_name", "utl_http.request",
		"extractvalue(", "updatexml(", "exp(", "floor(",
		"rand(", "count(*)", "group_concat(", "concat_ws(",
		"make_set(", "export_set(", "load_file(",
		"@@version", "@@datadir", "@@hostname", "@@basedir",
		"current_database()", "current_schema()", "current_user()",
		"session_user()", "system_user()", "user()",
	}
	
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	// Check for SQL injection using regex patterns
	sqlRegexPatterns := []string{
		`\b(union|select|insert|update|delete|drop|create|alter|exec|execute|declare)\b`,
		`\b(or|and)\s+\d+\s*=\s*\d+`,
		`\b(or|and)\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?`,
		`\b(or|and)\s+\d+\s*(>|<|>=|<=|<>|!=)\s*\d+`,
		`['"][^'"]*['"](\s*(or|and)\s*['"][^'"]*['"])+`,
		`\b(waitfor|delay|benchmark|sleep|pg_sleep)\s*\(`,
		`\b(load_file|into\s+(outfile|dumpfile))\b`,
		`\b(information_schema|mysql\.|sys\.|pg_catalog)\b`,
		`\b(@@\w+|current_user|session_user|system_user|user\(\))\b`,
		`\b(concat|substring|ascii|char|hex|md5|sha1)\s*\(`,
		`\b(extractvalue|updatexml|exp|floor|rand)\s*\(`,
		`\b(union\s+(all\s+)?select)\b`,
		`\b(having\s+\d+\s*=\s*\d+)\b`,
		`\b(group\s+by\s+\d+)\b`,
		`\b(order\s+by\s+\d+)\b`,
		`\b(limit\s+\d+(\s*,\s*\d+)?)\b`,
		`\b(offset\s+\d+)\b`,
	}
	
	for _, pattern := range sqlRegexPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return true
		}
	}
	
	return false
}

// DetectXSS detects potential XSS attempts
func (v *Validator) DetectXSS(input string) bool {
	lowerInput := strings.ToLower(input)
	
	// Common XSS patterns
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=", "onsubmit=",
		"onfocus=", "onblur=", "onchange=", "onkeyup=", "onkeydown=",
		"onmousedown=", "onmouseup=", "onmousemove=", "onmouseenter=", "onmouseleave=",
		"onkeypress=", "oncontextmenu=", "ondblclick=", "ondrag=", "ondrop=",
		"onscroll=", "onresize=", "onselect=", "ontoggle=", "onwheel=",
		"onanimationend=", "onanimationstart=", "ontransitionend=",
		"eval(", "expression(", "alert(", "confirm(", "prompt(",
		"document.cookie", "document.write", "window.location",
		"<iframe", "<object", "<embed", "<applet", "<meta",
		"<link", "<style", "<img", "<svg", "<form",
		"<input", "<textarea", "<select", "<option", "<button",
		"<audio", "<video", "<source", "<track", "<canvas",
		"<details", "<summary", "<marquee", "<bgsound",
		"<base", "<isindex", "<keygen", "<menuitem",
		"<math", "<foreignobject", "<desc", "<title",
		"<animate", "<animatetransform", "<animatemotion",
		"<set", "<use", "<image", "<text", "<tspan",
		"<textpath", "<altglyph", "<altglyphdef", "<altglyphitem",
		"<glyph", "<glyphref", "<marker", "<symbol",
		"<defs", "<g", "<switch", "<foreignobject",
		// Event handlers
		"onabort=", "onafterprint=", "onbeforeprint=", "onbeforeunload=",
		"oncanplay=", "oncanplaythrough=", "onchange=", "onclick=",
		"oncontextmenu=", "oncopy=", "oncuechange=", "oncut=",
		"ondblclick=", "ondrag=", "ondragend=", "ondragenter=",
		"ondragleave=", "ondragover=", "ondragstart=", "ondrop=",
		"ondurationchange=", "onemptied=", "onended=", "onerror=",
		"onfocus=", "onhashchange=", "oninput=", "oninvalid=",
		"onkeydown=", "onkeypress=", "onkeyup=", "onload=",
		"onloadeddata=", "onloadedmetadata=", "onloadstart=",
		"onmousedown=", "onmousemove=", "onmouseout=", "onmouseover=",
		"onmouseup=", "onmousewheel=", "onoffline=", "ononline=",
		"onpagehide=", "onpageshow=", "onpaste=", "onpause=",
		"onplay=", "onplaying=", "onpopstate=", "onprogress=",
		"onratechange=", "onreset=", "onresize=", "onscroll=",
		"onsearch=", "onseeked=", "onseeking=", "onselect=",
		"onstalled=", "onstorage=", "onsubmit=", "onsuspend=",
		"ontimeupdate=", "ontoggle=", "onunload=", "onvolumechange=",
		"onwaiting=", "onwheel=",
		// JavaScript protocols
		"javascript:", "data:text/html", "data:text/javascript",
		"data:application/javascript", "vbscript:", "livescript:",
		"mocha:", "feed:", "view-source:", "jar:", "wyciwyg:",
		// Dangerous functions
		"settimeout(", "setinterval(", "function(", "constructor(",
		"import(", "require(", "process.exit(", "process.kill(",
		"child_process", "fs.readfile", "fs.writefile",
		// HTML entities that could be used for bypassing
		"&lt;script", "&lt;/script", "&lt;iframe", "&lt;object",
		"&#60;script", "&#60;/script", "&#x3c;script", "&#x3c;/script",
		"&amp;lt;script", "&amp;#60;script", "&amp;#x3c;script",
		// Unicode bypasses
		"\u003cscript", "\u003c/script", "\u0022", "\u0027",
		"\u003e", "\u003c", "\u0026", "\u0023",
		// CSS expressions
		"expression(", "behavior:", "binding:", "-moz-binding:",
		"@import", "url(", "\\0000", "\\0001", "\\0002",
		// SVG XSS
		"<svg", "</svg>", "<g", "</g>", "<path", "<circle",
		"<rect", "<line", "<polygon", "<polyline", "<ellipse",
		"<text", "<tspan", "<use", "<image", "<foreignobject",
		// Data URLs
		"data:image/svg+xml", "data:text/html", "data:text/xml",
		// Base64 encoded
		"base64,", "atob(", "btoa(",
	}
	
	for _, pattern := range xssPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	// Check for XSS using regex patterns
	xssRegexPatterns := []string{
		`<[^>]*on\w+\s*=`,                                    // Event handlers
		`<[^>]*javascript:`,                                  // JavaScript protocol
		`<[^>]*data:\s*text/html`,                           // Data URL HTML
		`<[^>]*vbscript:`,                                   // VBScript protocol
		`<[^>]*expression\s*\(`,                             // CSS expression
		`<[^>]*behavior\s*:`,                                // CSS behavior
		`<[^>]*@import`,                                     // CSS import
		`<[^>]*url\s*\(`,                                    // CSS URL
		`<script[^>]*>.*?</script>`,                         // Script tags
		`<iframe[^>]*>.*?</iframe>`,                         // Iframe tags
		`<object[^>]*>.*?</object>`,                         // Object tags
		`<embed[^>]*>.*?</embed>`,                           // Embed tags
		`<applet[^>]*>.*?</applet>`,                         // Applet tags
		`<form[^>]*>.*?</form>`,                             // Form tags
		`<input[^>]*>`,                                      // Input tags
		`<textarea[^>]*>.*?</textarea>`,                     // Textarea tags
		`<select[^>]*>.*?</select>`,                         // Select tags
		`<button[^>]*>.*?</button>`,                         // Button tags
		`<link[^>]*>`,                                       // Link tags
		`<style[^>]*>.*?</style>`,                           // Style tags
		`<meta[^>]*>`,                                       // Meta tags
		`<base[^>]*>`,                                       // Base tags
		`<svg[^>]*>.*?</svg>`,                               // SVG tags
		`<math[^>]*>.*?</math>`,                             // Math tags
		`<details[^>]*>.*?</details>`,                       // Details tags
		`<summary[^>]*>.*?</summary>`,                       // Summary tags
		`<marquee[^>]*>.*?</marquee>`,                       // Marquee tags
		`<audio[^>]*>.*?</audio>`,                           // Audio tags
		`<video[^>]*>.*?</video>`,                           // Video tags
		`<canvas[^>]*>.*?</canvas>`,                         // Canvas tags
		`&#x?[0-9a-f]+;?`,                                   // HTML entities
		`\\u[0-9a-f]{4}`,                                    // Unicode escapes
		`\\x[0-9a-f]{2}`,                                    // Hex escapes
		`\\[0-7]{1,3}`,                                      // Octal escapes
		`eval\s*\(`,                                         // Eval function
		`function\s*\(`,                                     // Function constructor
		`constructor\s*\(`,                                  // Constructor
		`settimeout\s*\(`,                                   // setTimeout
		`setinterval\s*\(`,                                  // setInterval
		`document\.(write|writeln|cookie|domain|location)`,  // Document methods
		`window\.(location|open|eval|execScript)`,           // Window methods
		`location\.(href|replace|assign)`,                   // Location methods
		`history\.(pushState|replaceState)`,                 // History methods
		`alert\s*\(`,                                        // Alert function
		`confirm\s*\(`,                                      // Confirm function
		`prompt\s*\(`,                                       // Prompt function
		`console\.(log|error|warn|info)`,                    // Console methods
		`fetch\s*\(`,                                        // Fetch API
		`xmlhttprequest`,                                    // XMLHttpRequest
		`websocket`,                                         // WebSocket
		`postmessage`,                                       // PostMessage
		`localstorage`,                                      // LocalStorage
		`sessionstorage`,                                    // SessionStorage
		`indexeddb`,                                         // IndexedDB
		`webworker`,                                         // Web Worker
		`serviceworker`,                                     // Service Worker
		`import\s*\(`,                                       // Dynamic import
		`require\s*\(`,                                      // CommonJS require
	}
	
	for _, pattern := range xssRegexPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return true
		}
	}
	
	return false
}

// DetectPathTraversal detects path traversal attempts
func (v *Validator) DetectPathTraversal(input string) bool {
	// Common path traversal patterns
	pathPatterns := []string{
		"../", "..\\", "..", "/etc/", "/proc/", "/sys/",
		"\\windows\\", "\\system32\\", "/var/", "/usr/",
		"%2e%2e", "%2f", "%5c", "..%2f", "..%5c",
		"....//", "....\\\\",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range pathPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	
	return false
}

// ValidateContentLength validates content length headers
func (v *Validator) ValidateContentLength(contentLength int64, maxSize int64) bool {
	return contentLength >= 0 && contentLength <= maxSize
}

// SanitizeHeader sanitizes HTTP headers
func (v *Validator) SanitizeHeader(header string) string {
	// Remove control characters and dangerous patterns
	sanitized := v.SanitizeString(header)
	
	// Additional header-specific sanitization
	if len(sanitized) > 8192 { // Max header size
		sanitized = sanitized[:8192]
	}
	
	return sanitized
}