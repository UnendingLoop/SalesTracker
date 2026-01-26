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

const (
	FamMother   = "mother"
	FamFather   = "father"
	FamDaughter = "daughter"
	FamSon      = "son"
	FamUnknown  = "unknown"
)

const (
	OpTypeDebit  = "debit"  // поступление средств в бюджет
	OpTypeCredit = "credit" // трата средств из бюджета
)

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

type AnalyticsQuantum struct { // возвращается при запросе без группировок
	Key    string  `json:"key,omitempty"`
	Sum    float64 `json:"sum"`    // сумма в копейках
	Avg    float64 `json:"avg"`    // среднее в копейках
	Count  int     `json:"count"`  // кол-во записей, на которых основано вычисление
	Median float64 `json:"median"` // медиана в копейках
	P90    float64 `json:"p90"`    // 90й перц в копейках
}

type AnalyticsSummary struct { // возвращается если в запросе указана группировка
	Sum    int                `json:"sum"`
	Avg    float64            `json:"avg"` // среднее в копейках
	Count  int                `json:"count"`
	Median float64            `json:"median"` // медиана в копейках
	P90    float64            `json:"p90"`    // 90й перц в копейках
	Groups []AnalyticsQuantum `json:"groups,omitempty"`
}

type RequestParamOperations struct {
	OrderBy   *string
	ASC       bool
	DESC      bool
	StartTime *time.Time
	EndTime   *time.Time
	Page      *int
	Limit     *int
}

type RequestParamAnalytics struct {
	GroupBy   *string
	StartTime *time.Time
	EndTime   *time.Time
	Page      *int
	Limit     *int
}

const (
	GroupByDay      = "day"
	GroupByWeek     = "week"
	GroupByMonth    = "month"
	GroupByYear     = "year"
	GroupByActor    = "actor"
	GroupByCategory = "category"
	GroupByOpType   = "type"
)

const (
	OrderByOpID     = "id"
	OrderByAmount   = "amount"
	OrderByActor    = "actor"
	OrderByCategory = "category"
	OrderByType     = "type"
	OrderByOpDate   = "operation_at"
)
