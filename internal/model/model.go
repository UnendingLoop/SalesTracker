package model

import "time"

type Operation struct {
	ID          int64
	Amount      int64 // в копейках
	Actor       string
	Category    string
	Type        string    // debit/credit
	OperationAt time.Time // время самой операции
	CreatedAt   time.Time // время создания записи в БД
	Description *string
}

var ActorsMap = map[string]struct{}{FamMother: {}, FamFather: {}, FamDaughter: {}, FamSon: {}}

const (
	FamMother   = "mother"
	FamFather   = "father"
	FamDaughter = "daughter"
	FamSon      = "son"
	FamOther    = "other"
)

var OpTypeMap = map[string]struct{}{OpTypeCredit: {}, OpTypeDebit: {}}

const (
	OpTypeDebit  = "debit"  // поступление средств в бюджет
	OpTypeCredit = "credit" // трата средств из бюджета
)

var CategoriesMap = map[string]struct{}{CatSalary: {}, CatChores: {}, CatTransport: {}, CatFood: {}, CatEntert: {}, CatHealth: {}, CatEducation: {}, CatPresents: {}, CatElectronics: {}, CatComm: {}}

const (
	CatSalary      = "salary"
	CatChores      = "chores"
	CatTransport   = "transport"
	CatFood        = "food"
	CatEntert      = "entertainment"
	CatHealth      = "health"
	CatEducation   = "education"
	CatPresents    = "presents"
	CatElectronics = "electronics"
	CatComm        = "communication"
	CatOther       = "other"
)

// ANALYTICS

type AnalyticsQuantum struct { // возвращается в виде массива если в запросе указана группировка
	Key    string  `json:"key"`
	Sum    int     `json:"sum"`    // сумма в копейках
	Avg    float64 `json:"avg"`    // среднее в копейках
	Count  int     `json:"count"`  // кол-во записей, на которых основано вычисление
	Median float64 `json:"median"` // медиана в копейках
	P90    float64 `json:"p90"`    // 90й перц в копейках
}

type AnalyticsSummary struct {
	Key    string             `json:"key,omitempty"` // поле, использованное для группировки
	Sum    int                `json:"sum"`
	Avg    float64            `json:"avg"` // среднее в копейках
	Count  int                `json:"count"`
	Median float64            `json:"median"` // медиана в копейках
	P90    float64            `json:"p90"`    // 90й перц в копейках
	Groups []AnalyticsQuantum `json:"groups,omitempty"`
}

type RequestParamOperations struct {
	OrderBy   *string    `form:"order_by"`
	ASC       bool       `form:"asc"`
	DESC      bool       `form:"desc"`
	StartTime *time.Time `form:"from"`
	EndTime   *time.Time `form:"to"`
	Page      *int       `form:"page"`
	Limit     *int       `form:"limit"`
}

type RequestParamAnalytics struct {
	GroupBy   *string    `form:"group_by"`
	StartTime *time.Time `form:"from"`
	EndTime   *time.Time `form:"to"`
	Page      *int       `form:"page"`
	Limit     *int       `form:"limit"`
}

var GroupingMap = map[string]struct{}{GroupByDay: {}, GroupByWeek: {}, GroupByMonth: {}, GroupByYear: {}, GroupByActor: {}, GroupByCategory: {}, GroupByOpType: {}}

const (
	GroupByDay      = "day"
	GroupByWeek     = "week"
	GroupByMonth    = "month"
	GroupByYear     = "year"
	GroupByActor    = "actor"
	GroupByCategory = "category"
	GroupByOpType   = "type"
)

var OrderMap = map[string]struct{}{OrderByOpID: {}, OrderByAmount: {}, OrderByActor: {}, OrderByCategory: {}, OrderByType: {}, OrderByOpDate: {}}

const (
	OrderByOpID     = "id"
	OrderByAmount   = "amount"
	OrderByActor    = "actor"
	OrderByCategory = "category"
	OrderByType     = "type"
	OrderByOpDate   = "operation_at"
)
