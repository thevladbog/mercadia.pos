package app

import (
	"context"
	"errors"

	"mercadia.dev/pos/services/store-edge/internal/domain"
)

var ErrInvalidMarkingCommand = errors.New("invalid marking command")

type MarkingService struct {
	receipts ReceiptRepository
}

func NewMarkingService(receipts ReceiptRepository) *MarkingService {
	return &MarkingService{receipts: receipts}
}

type ValidateMarkingCommand struct {
	ReceiptID string
	Code      string
}

type MarkingValidationResult struct {
	Validation domain.MarkingValidationResult
}

func (s *MarkingService) ValidateMarking(ctx context.Context, command ValidateMarkingCommand) (MarkingValidationResult, error) {
	if command.ReceiptID == "" || command.Code == "" {
		return MarkingValidationResult{}, ErrInvalidMarkingCommand
	}
	if _, err := s.receipts.FindReceipt(ctx, command.ReceiptID); err != nil {
		return MarkingValidationResult{}, err
	}

	validation, err := domain.ValidateDataMatrixCode(command.Code)
	if err != nil {
		return MarkingValidationResult{}, ErrInvalidMarkingCommand
	}
	return MarkingValidationResult{Validation: validation}, nil
}
