package service

import (
	"context"

	"github.com/UnendingLoop/SalesTracker/internal/model"
	"github.com/UnendingLoop/SalesTracker/internal/repository"
)

type OperationService struct {
	repo repository.OperationsRepository
}

func NewOperationService(repo repository.OperationsRepository) *OperationService {
	return &OperationService{repo: repo}
}

func (os *OperationService) CreateOperation(ctx context.Context, newOp *model.Operation) error{
	//валидация входящих данных

}
func (os *OperationService) GetOperationByID(ctx context.Context, id int) (*model.Operation, error)
func (os *OperationService) GetAllOperations(ctx context.Context, rpo *model.RequestParamOperations) ([]*model.Operation, error)
func (os *OperationService) UpdateOperationByID(ctx context.Context, op *model.Operation) error
func (os *OperationService) DeleteOperationByID(ctx context.Context, id int) error
func (os *OperationService) GetAnalytics(ctx context.Context, rpa *model.RequestParamAnalytics) (*model.AnalyticsSummary, error){

}

func validateOperation(op *model.Operation) error{
	if op.Amount<=0{}
	if op.Actor==""{}
	if op.Category==""{}
	 if op.Type 
	}
}