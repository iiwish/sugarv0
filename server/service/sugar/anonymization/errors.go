package anonymization

import "fmt"

// ValidationError 表示验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("验证错误: %s", e.Message)
}

// NewValidationError 创建新的验证错误
func NewValidationError(message string) error {
	return &ValidationError{Message: message}
}

// ProcessingError 表示数据处理错误
type ProcessingError struct {
	Message string
	Cause   error
}

func (e *ProcessingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("处理错误: %s, 原因: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("处理错误: %s", e.Message)
}

// NewProcessingError 创建新的处理错误
func NewProcessingError(message string, cause error) error {
	return &ProcessingError{Message: message, Cause: cause}
}

// AnonymizationError 表示匿名化错误
type AnonymizationError struct {
	Message string
	Cause   error
}

func (e *AnonymizationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("匿名化错误: %s, 原因: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("匿名化错误: %s", e.Message)
}

// NewAnonymizationError 创建新的匿名化错误
func NewAnonymizationError(message string, cause error) error {
	return &AnonymizationError{Message: message, Cause: cause}
}
