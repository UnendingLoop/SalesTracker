package transport

import (
	"context"
	"encoding/csv"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/UnendingLoop/SalesTracker/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/form"
	"github.com/wb-go/wbf/ginext"
)

type OperationHandler struct {
	svc OperationService
}

type OperationService interface {
	CreateOperation(ctx context.Context, newOp *model.Operation) error
	GetOperationByID(ctx context.Context, id int) (*model.Operation, error)
	GetAllOperations(ctx context.Context, rpo *model.RequestParamOperations) ([]model.Operation, error)
	UpdateOperationByID(ctx context.Context, op *model.Operation) error
	DeleteOperationByID(ctx context.Context, id int) error
	GetAnalytics(ctx context.Context, rpa *model.RequestParamAnalytics) (*model.AnalyticsSummary, error)
}

func NewOperationHandler(svc OperationService) *OperationHandler {
	return &OperationHandler{svc: svc}
}

func (h *OperationHandler) SimplePinger(ctx *ginext.Context) {
	ctx.JSON(200, gin.H{"message": "pong"})
}

func (h *OperationHandler) CreateOperation(ctx *ginext.Context) {
	var newOp model.Operation
	if err := ctx.ShouldBindJSON(&newOp); err != nil {
		log.Printf("failed to parse operation payload: %q", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation payload"})
		return
	}

	if err := h.svc.CreateOperation(ctx.Request.Context(), &newOp); err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *OperationHandler) GetOperationByID(ctx *ginext.Context) {
	// читаем id из params
	idRaw, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "operation id is missing"})
		return
	}
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to process specified operation id"})
		return
	}

	// вызываем сервис
	res, err := h.svc.GetOperationByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *OperationHandler) GetAllOperations(ctx *ginext.Context) {
	// парсим параметры запроса операций из URL
	rpo := model.RequestParamOperations{}
	decoder := form.NewDecoder()
	if err := decoder.Decode(&rpo, ctx.Request.URL.Query()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// вызывваем сервис
	res, err := h.svc.GetAllOperations(ctx.Request.Context(), &rpo)
	if err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *OperationHandler) UpdateOperationByID(ctx *ginext.Context) {
	// читаем id из params
	idRaw, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "operation id is missing"})
		return
	}
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to process specified operation id"})
		return
	}

	// читаем JSON
	var op model.Operation
	if err := ctx.ShouldBindJSON(&op); err != nil {
		log.Printf("failed to parse operation payload: %q", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation payload"})
		return
	}
	op.ID = id

	// вызываем сервис
	if err := h.svc.UpdateOperationByID(ctx.Request.Context(), &op); err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *OperationHandler) DeleteOperationByID(ctx *ginext.Context) {
	// читаем id из params
	idRaw, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "operation id is missing"})
		return
	}
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to process specified operation id"})
		return
	}

	// вызываем сервис
	if err := h.svc.DeleteOperationByID(ctx.Request.Context(), id); err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *OperationHandler) GetAnalytics(ctx *ginext.Context) {
	// парсим параметры запроса аналитики из URL
	rpa := model.RequestParamAnalytics{}
	decoder := form.NewDecoder()
	if err := decoder.Decode(&rpa, ctx.Request.URL.Query()); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// вызываем сервис
	res, err := h.svc.GetAnalytics(ctx.Request.Context(), &rpa)
	if err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	// пакуем аналитику в ответ
	ctx.JSON(http.StatusOK, res)
}

func (h *OperationHandler) ExportOperationsCSV(ctx *ginext.Context) {
	// парсим параметры запроса операций из URL
	rpo := model.RequestParamOperations{}
	if err := decodeQueryParams(ctx, &rpo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// получаем массив строк
	res, err := h.svc.GetAllOperations(ctx.Request.Context(), &rpo)
	if err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	// устанавливаем хедеры под CSV
	rows := convertOperationsToCSV(res)
	ctx.Writer.Header().Set("Cache-Control", "no-store")
	ctx.Writer.Header().Set("Pragma", "no-cache")
	ctx.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Writer.Header().Set("Content-Type", "text/csv")
	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=operations.csv")

	// пишем данные
	writer := csv.NewWriter(ctx.Writer)
	if err := writer.WriteAll(rows); err != nil {
		log.Printf("failed to Flush csv-writer: %q", err.Error())
		return
	}
}

func (h *OperationHandler) ExportAnalyticsCSV(ctx *ginext.Context) {
	// парсим параметры запроса аналитики из URL
	rpa := model.RequestParamAnalytics{}
	if err := decodeQueryParams(ctx, &rpa); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// получаем массив строк
	res, err := h.svc.GetAnalytics(ctx.Request.Context(), &rpa)
	if err != nil {
		ctx.JSON(errCodeDefiner(err), gin.H{"error": err.Error()})
		return
	}

	// устанавливаем хедеры под CSV
	ctx.Writer.Header().Set("Cache-Control", "no-store")
	ctx.Writer.Header().Set("Pragma", "no-cache")
	ctx.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Writer.Header().Set("Content-Type", "text/csv")
	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=analytics.csv")

	// пишем данные
	rows := convertAnalyticsToCSV(res)
	writer := csv.NewWriter(ctx.Writer)
	if err := writer.WriteAll(rows); err != nil {
		log.Printf("failed to Flush csv-writer: %q", err.Error())
		return
	}
}

func convertAnalyticsToCSV(input *model.AnalyticsSummary) [][]string {
	result := make([][]string, 0, len(input.Groups)+1)
	start := []string{"group_key", "total_amount", "average", "operations_in_group", "mediana", "P90"}
	result = append(result, start)

	for _, v := range input.Groups {
		row := make([]string, 0, len(start))
		row = append(row, v.Key, strconv.FormatFloat(v.Sum/100, 'f', 2, 64), strconv.FormatFloat(v.Avg/100, 'f', 2, 64), strconv.Itoa(v.Count), strconv.FormatFloat(v.Median/100, 'f', 2, 64), strconv.FormatFloat(v.P90/100, 'f', 2, 64))
		result = append(result, row)
	}

	end := []string{"TOTALS:", strconv.FormatFloat(input.Sum/100, 'f', 2, 64), strconv.FormatFloat(input.Avg/100, 'f', 2, 64), strconv.Itoa(input.Count), strconv.FormatFloat(input.Median/100, 'f', 2, 64), strconv.FormatFloat(input.P90/100, 'f', 2, 64)}
	result = append(result, end)
	return result
}

func convertOperationsToCSV(input []model.Operation) [][]string {
	result := make([][]string, 0, len(input)+1)
	start := []string{"id", "amount", "type", "category", "actor", "date", "created", "description"}
	result = append(result, start)

	for _, v := range input {
		row := make([]string, 0, len(start))
		descr := ""
		if v.Description != nil {
			descr = *v.Description
		}
		row = append(row, strconv.FormatInt(v.ID, 10), strconv.FormatFloat(float64(v.Amount)/100, 'f', 2, 64), v.Type, v.Category, v.Actor, v.OperationAt.Format("2006-01-02"), v.CreatedAt.Format("2006-01-02"), descr)
		result = append(result, row)
	}

	return result
}

func decodeQueryParams[T *model.RequestParamAnalytics | *model.RequestParamOperations](c *ginext.Context, input T) error {
	decoder := form.NewDecoder()
	if err := decoder.Decode(input, c.Request.URL.Query()); err != nil {
		return err
	}
	return nil
}

func errCodeDefiner(err error) int {
	switch {
	case errors.Is(err, model.ErrCommon500):
		return 500
	case errors.Is(err, model.ErrOperationIDNotFound):
		return 404
	default:
		return 400
	}
}
