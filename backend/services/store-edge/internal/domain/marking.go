package domain

import (
	"errors"
	"strings"
)

var ErrInvalidMarkingInput = errors.New("invalid marking input")

type MarkingValidationResult struct {
	Valid     bool
	Code      string
	ProductID string
	Message   string
}

func ValidateDataMatrixCode(code string) (MarkingValidationResult, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return MarkingValidationResult{}, ErrInvalidMarkingInput
	}

	// Mock provider: GS1 DataMatrix codes for marked goods start with application identifier 01.
	if strings.HasPrefix(code, "01") && len(code) >= 16 {
		return MarkingValidationResult{
			Valid:     true,
			Code:      code,
			ProductID: "marked-" + code[2:16],
			Message:   "marking code validated",
		}, nil
	}

	return MarkingValidationResult{
		Valid:   false,
		Code:    code,
		Message: "marking code rejected by mock provider",
	}, nil
}
