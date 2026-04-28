package middleware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriterWrapper wraps gin.ResponseWriter to capture response body
type ResponseWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while writing to the original writer
func (w *ResponseWriterWrapper) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

// Body returns the captured response body
func (w *ResponseWriterWrapper) Body() []byte {
	return w.body.Bytes()
}

// ETagData represents the components used for ETag generation
type ETagData struct {
	Path      string
	Method    string
	Content   []byte
	Timestamp time.Time
	UserID    string
	Params    string
}

// ETagGenerator handles sophisticated ETag generation
type ETagGenerator struct {
	secretKey    []byte
	cacheStorage sync.Map // For storing computed ETags
}

// NewETagGenerator creates a new ETag generator instance
func NewETagGenerator(secretKey string) *ETagGenerator {
	return &ETagGenerator{
		secretKey: []byte(secretKey),
	}
}

// GenerateETag creates a sophisticated ETag based on multiple factors
func (g *ETagGenerator) GenerateETag(data ETagData) string {
	// Create a composite hash from multiple factors
	h := sha256.New()

	// Include path and method
	h.Write([]byte(data.Path))
	h.Write([]byte(data.Method))

	// Include content if available
	if len(data.Content) > 0 {
		h.Write(data.Content)
	}

	// Include timestamp (rounded to second for consistency)
	timestampBytes := fmt.Appendf(nil, "%d", data.Timestamp.Unix())
	h.Write(timestampBytes)

	// Include user context for personalized responses
	if data.UserID != "" {
		h.Write([]byte(data.UserID))
	}

	// Include query parameters for cache differentiation
	if data.Params != "" {
		h.Write([]byte(data.Params))
	}

	// Add secret key for security
	h.Write(g.secretKey)

	// Generate final ETag
	hash := h.Sum(nil)
	return fmt.Sprintf("\"%x\"", hash)
}

// GenerateWeakETag creates a weak ETag for dynamic content
func (g *ETagGenerator) GenerateWeakETag(data ETagData) string {
	etag := g.GenerateETag(data)
	return "W/" + etag
}

// GetCachedETag retrieves a cached ETag if available
func (g *ETagGenerator) GetCachedETag(key string) (string, bool) {
	if cached, ok := g.cacheStorage.Load(key); ok {
		if etag, ok := cached.(string); ok {
			return etag, true
		}
	}
	return "", false
}

// CacheETag stores an ETag in the cache
func (g *ETagGenerator) CacheETag(key, etag string) {
	g.cacheStorage.Store(key, etag)
}

// InvalidateCache removes ETag entries matching a pattern
func (g *ETagGenerator) InvalidateCache(pattern string) {
	g.cacheStorage.Range(func(key, value any) bool {
		if keyStr, ok := key.(string); ok {
			if strings.Contains(keyStr, pattern) {
				g.cacheStorage.Delete(key)
			}
		}
		return true
	})
}

// Global ETag generator instance
var etagGenerator = NewETagGenerator("telecom-platform-etag-secret")

// generateETag generates a sophisticated ETag for caching
func generateETag(c *gin.Context) string {
	// Extract relevant data for ETag generation
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	// Try to get cached ETag first
	cacheKey := fmt.Sprintf("%s:%s:%s", data.Method, data.Path, data.Params)
	if cached, exists := etagGenerator.GetCachedETag(cacheKey); exists {
		return cached
	}

	// Generate new ETag
	etag := etagGenerator.GenerateETag(data)

	// Cache the ETag for future use
	etagGenerator.CacheETag(cacheKey, etag)

	return etag
}

// generateContentETag creates an ETag based on response content
func generateContentETag(c *gin.Context, content []byte) string {
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Content:   content,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	return etagGenerator.GenerateETag(data)
}

// generateWeakETag creates a weak ETag for dynamic content
func generateWeakETag(c *gin.Context) string {
	data := ETagData{
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
		Timestamp: time.Now(),
		Params:    c.Request.URL.RawQuery,
	}

	// Add user context if available
	if userID, exists := c.Get("user_id"); exists {
		if idStr, ok := userID.(string); ok {
			data.UserID = idStr
		}
	}

	return etagGenerator.GenerateWeakETag(data)
}

// ETagValidationMiddleware handles ETag validation for conditional requests
func ETagValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only handle GET and HEAD requests for ETag validation
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.Next()
			return
		}

		// Check If-None-Match header (for GET requests)
		if noneMatch := c.GetHeader("If-None-Match"); noneMatch != "" {
			// Generate current ETag
			currentETag := generateETag(c)

			// Check if any of the ETags match
			if etagMatches(noneMatch, currentETag) {
				c.Header("ETag", currentETag)
				c.Status(http.StatusNotModified)
				c.Abort()
				return
			}
		}

		// Check If-Match header (for PUT/PATCH/DELETE requests)
		if match := c.GetHeader("If-Match"); match != "" {
			currentETag := generateETag(c)

			if !etagMatches(match, currentETag) {
				c.Status(http.StatusPreconditionFailed)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ContentBasedETagMiddleware generates ETags based on response content
func ContentBasedETagMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Wrap the response writer to capture the body
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = wrapper

		c.Next()

		// Only generate ETags for successful GET requests
		if c.Request.Method != "GET" || c.Writer.Status() != http.StatusOK {
			return
		}

		// Generate ETag based on captured response content
		responseBody := wrapper.Body()
		if len(responseBody) > 0 {
			etag := generateContentETag(c, responseBody)
			c.Header("ETag", etag)
		} else {
			// Fallback to weak ETag for empty responses
			etag := generateWeakETag(c)
			c.Header("ETag", etag)
		}
	}
}

// CacheInvalidationMiddleware handles cache invalidation for modified resources
func CacheInvalidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original request for post-processing
		c.Next()

		// Invalidate cache for POST, PUT, PATCH, DELETE requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" ||
			c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {

			// Invalidate ETags for this resource pattern
			pattern := c.Request.URL.Path
			etagGenerator.InvalidateCache(pattern)

			// Also invalidate related collection resources
			if strings.Contains(pattern, "/") {
				parts := strings.Split(strings.Trim(pattern, "/"), "/")
				if len(parts) > 0 {
					collectionPattern := "/" + parts[0]
					etagGenerator.InvalidateCache(collectionPattern)
				}
			}
		}
	}
}

// etagMatches checks if any of the provided ETags match the current ETag
func etagMatches(headerValue, currentETag string) bool {
	// Handle wildcard
	if headerValue == "*" {
		return true
	}

	// Parse the header value (can contain multiple ETags)
	etags := strings.SplitSeq(headerValue, ",")
	for etag := range etags {
		etag = strings.TrimSpace(etag)
		// Remove quotes for comparison
		etag = strings.Trim(etag, "\"")
		current := strings.Trim(currentETag, "\"W/")

		if etag == current || etag == "W/"+current {
			return true
		}
	}

	return false
}
