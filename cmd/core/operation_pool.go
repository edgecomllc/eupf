package core

import (
	"github.com/rs/zerolog/log"
)

type Operation struct {
	Apply    func() error
	Rollback func() error
}

type OperationPool struct {
	operations  []Operation
	failedIndex int
}

func NewOperationPool() *OperationPool {
	return &OperationPool{
		operations:  make([]Operation, 0),
		failedIndex: -1,
	}
}

func (op *OperationPool) Add(operation Operation) {
	op.operations = append(op.operations, operation)
}

func (op *OperationPool) Commit() error {
	op.failedIndex = -1

	for i, operation := range op.operations {
		if err := operation.Apply(); err != nil {
			log.Warn().Msgf("Operation failed at index %d: %v", i, err)
			op.failedIndex = i
			op.Rollback()
			return err
		}
	}

	return nil
}

func (op *OperationPool) Rollback() {
	if op.failedIndex == -1 {
		log.Warn().Msg("Rollback called without a failed operation")
		return
	}

	for i := op.failedIndex; i >= 0; i-- {
		if err := op.operations[i].Rollback(); err != nil {
			log.Warn().Msgf("Rollback failed at index %d: %v", i, err)
		}
	}

	op.failedIndex = -1
}
