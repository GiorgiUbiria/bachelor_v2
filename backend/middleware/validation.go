package middleware

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents validation error response
type ErrorResponse struct {
	Error       bool        `json:"error"`
	FailedField string      `json:"failed_field"`
	Tag         string      `json:"tag"`
	Value       interface{} `json:"value"`
}

// XValidator wraps the validator instance
type XValidator struct {
	validator *validator.Validate
}

// GlobalErrorHandlerResp represents global error response
type GlobalErrorHandlerResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// This is the validator instance
var validate = validator.New()

// NewValidator creates a new validator instance
func NewValidator() *XValidator {
	return &XValidator{
		validator: validate,
	}
}

// Validate validates the given data structure
func (v XValidator) Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	errs := validate.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem ErrorResponse

			elem.FailedField = err.Field()
			elem.Tag = err.Tag()
			elem.Value = err.Value()
			elem.Error = true

			validationErrors = append(validationErrors, elem)
		}
	}

	return validationErrors
}

// ValidateStruct validates a struct and returns formatted error
func ValidateStruct(data interface{}) error {
	validator := NewValidator()

	if errs := validator.Validate(data); len(errs) > 0 && errs[0].Error {
		errMsgs := make([]string, 0)

		for _, err := range errs {
			errMsgs = append(errMsgs, fmt.Sprintf(
				"[%s]: '%v' | Needs to implement '%s'",
				err.FailedField,
				err.Value,
				err.Tag,
			))
		}

		return &fiber.Error{
			Code:    fiber.ErrBadRequest.Code,
			Message: strings.Join(errMsgs, " and "),
		}
	}

	return nil
}

// ValidationMiddleware creates a validation middleware for specific struct types
func ValidationMiddleware(structType interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := c.BodyParser(structType); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Success: false,
				Message: "Invalid request body: " + err.Error(),
			})
		}

		if err := ValidateStruct(structType); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(GlobalErrorHandlerResp{
				Success: false,
				Message: err.Error(),
			})
		}

		// Store validated data in context
		c.Locals("validated_data", structType)
		return c.Next()
	}
}

// GetValidatedData retrieves validated data from context
func GetValidatedData(c *fiber.Ctx) interface{} {
	return c.Locals("validated_data")
}
