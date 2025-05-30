package middleware

import (
	"html"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SanitizeInput sanitizes input data to prevent XSS and injection attacks
func SanitizeInput() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the request body
		body := c.Body()
		if len(body) > 0 {
			// Sanitize the body content
			sanitizedBody := sanitizeString(string(body))
			c.Request().SetBody([]byte(sanitizedBody))
		}

		// Sanitize query parameters
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			sanitizedValue := sanitizeString(string(value))
			c.Request().URI().QueryArgs().Set(string(key), sanitizedValue)
		})

		return c.Next()
	}
}

// sanitizeString removes potentially dangerous characters and HTML tags
func sanitizeString(input string) string {
	// HTML escape to prevent XSS
	escaped := html.EscapeString(input)

	// Remove common SQL injection patterns (basic protection)
	dangerous := []string{
		"<script", "</script>", "javascript:", "onload=", "onerror=",
		"onclick=", "onmouseover=", "onfocus=", "onblur=", "onchange=",
		"onsubmit=", "onreset=", "onselect=", "onkeydown=", "onkeypress=",
		"onkeyup=", "onmousedown=", "onmouseup=", "onmousemove=", "onmouseout=",
	}

	result := escaped
	for _, pattern := range dangerous {
		result = strings.ReplaceAll(strings.ToLower(result), pattern, "")
	}

	return result
}

// SanitizeStruct sanitizes all string fields in a struct
func SanitizeStruct(data interface{}) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			sanitized := sanitizeString(field.String())
			field.SetString(sanitized)
		}
	}
}

// ValidateAndSanitize combines validation and sanitization
func ValidateAndSanitize(data interface{}) error {
	// First sanitize the data
	SanitizeStruct(data)

	// Then validate it
	return ValidateStruct(data)
}
