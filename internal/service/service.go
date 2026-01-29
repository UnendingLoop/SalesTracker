package service

import (
	"context"
	"errors"
	"log"

	"github.com/UnendingLoop/SalesTracker/internal/model"
	"github.com/UnendingLoop/SalesTracker/internal/repository"
)

type OperationService struct {
	repo repository.OperationsRepository
}

func NewOperationService(repo repository.OperationsRepository) *OperationService {
	return &OperationService{repo: repo}
}

func (svc *OperationService) CreateOperation(ctx context.Context, newOp *model.Operation) error {
	// валидация входящей операции
	if err := validateOperation(newOp); err != nil {
		return err
	}

	// отправляем в репо
	if err := svc.repo.Create(ctx, newOp); err != nil {
		log.Printf("Failed to create new operation in DB: %q", err.Error())
		return model.ErrCommon500
	}

	return nil
}

func (svc *OperationService) GetOperationByID(ctx context.Context, id int) (*model.Operation, error) {
	if id <= 0 {
		return nil, model.ErrInvalidID
	}

	res, err := svc.repo.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrOperationIDNotFound):
			return nil, err
		default:
			log.Printf("Failed to get operation by ID from DB: %q", err.Error())
			return nil, model.ErrCommon500
		}
	}
	return res, nil
}

func (svc *OperationService) GetAllOperations(ctx context.Context, rpo *model.RequestParamOperations) ([]model.Operation, error) {
	// валидация параметров
	if rpo == nil {
		rpo = &model.RequestParamOperations{}
	}
	if err := validateOperationReqParams(rpo); err != nil {
		return nil, err
	}

	// идем в репо
	res, err := svc.repo.List(ctx, rpo)
	if err != nil {
		log.Printf("Failed to get operations list from DB: %q", err.Error())
		return nil, model.ErrCommon500
	}

	return res, nil
}

func (svc *OperationService) UpdateOperationByID(ctx context.Context, op *model.Operation) error {
	if op.ID <= 0 {
		return model.ErrInvalidID
	}
	// идем в репо
	if err := svc.repo.Update(ctx, op); err != nil {
		switch {
		case errors.Is(err, model.ErrUnknownActorOrCategory) || errors.Is(err, model.ErrOperationIDNotFound):
			return err
		default:
			log.Printf("Failed to update operation by ID in DB: %q", err.Error())
			return model.ErrCommon500
		}
	}
	return nil
}

func (svc *OperationService) DeleteOperationByID(ctx context.Context, id int) error {
	if id <= 0 {
		return model.ErrInvalidID
	}

	// идем в репо
	if err := svc.repo.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, model.ErrOperationIDNotFound):
			return err
		default:
			log.Printf("Failed to delete operation by ID in DB: %q", err.Error())
			return model.ErrCommon500
		}
	}
	return nil
}

func (svc *OperationService) GetAnalytics(ctx context.Context, rpa *model.RequestParamAnalytics) (*model.AnalyticsSummary, error) {
	// валидируем параметры запроса
	if rpa == nil {
		rpa = &model.RequestParamAnalytics{}
	}
	if err := validateAnalyticsReqParams(rpa); err != nil {
		return nil, err
	}

	// получаем саммари
	summary, err := svc.repo.AnalyticsSummary(ctx, rpa)
	if err != nil {
		log.Printf("analytics summary query failed: %q", err.Error())
		return nil, model.ErrCommon500
	}

	if rpa.GroupBy == nil {
		return summary, nil
	}
	// если запрос с группировкой - делаем и его
	groups, err := svc.repo.AnalyticsGroup(ctx, rpa)
	if err != nil {
		log.Printf("analytics group query failed: %q", err.Error())
		return nil, model.ErrCommon500
	}

	// собираем воедино
	summary.Groups = groups
	summary.Key = *rpa.GroupBy
	return summary, nil
}

func validateOperation(op *model.Operation) error {
	if op.Amount <= 0 {
		return model.ErrInvalidAmount
	}
	if _, ok := model.ActorsMap[op.Actor]; !ok {
		return model.ErrInvalidActor
	}
	if _, ok := model.CategoriesMap[op.Category]; !ok {
		return model.ErrInvalidCategory
	}
	if _, ok := model.OpTypeMap[op.Type]; !ok {
		return model.ErrInvalidOpType
	}
	if op.OperationAt.IsZero() {
		return model.ErrInvalidOpTime
	}

	if op.Type == model.OpTypeCredit {
		op.Amount = op.Amount * -1
	}
	return nil
}

func validateOperationReqParams(rpo *model.RequestParamOperations) error {
	if rpo.OrderBy != nil {
		// валидация самого OrderBy
		if _, ok := model.OrderMap[*rpo.OrderBy]; !ok {
			return model.ErrInvalidOrderBy
		}
		// валидация asc/desc
		if rpo.ASC == rpo.DESC {
			return model.ErrInvalidAscDesc
		}
	}

	if rpo.StartTime != nil && rpo.EndTime != nil {
		if rpo.StartTime.After(*rpo.EndTime) {
			return model.ErrInvalidStartEndTime
		}
	}

	if rpo.Page != nil {
		if *rpo.Page <= 0 {
			return model.ErrInvalidPage
		}
		if *rpo.Limit <= 0 || *rpo.Limit >= 1000 {
			return model.ErrInvalidLimit
		}
	}

	return nil
}

func validateAnalyticsReqParams(rpa *model.RequestParamAnalytics) error {
	if rpa.GroupBy != nil {
		// валидация самого groupby
		if _, ok := model.GroupingMap[*rpa.GroupBy]; !ok {
			return model.ErrInvalidGroupBy
		}
	}

	if rpa.StartTime != nil && rpa.EndTime != nil {
		if rpa.StartTime.After(*rpa.EndTime) {
			return model.ErrInvalidStartEndTime
		}
	}

	if rpa.Page != nil {
		if *rpa.Page <= 0 {
			return model.ErrInvalidPage
		}
		if *rpa.Limit <= 0 || *rpa.Limit >= 1000 {
			return model.ErrInvalidLimit
		}
	}

	return nil
}
