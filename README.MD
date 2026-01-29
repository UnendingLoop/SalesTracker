## Finance Tracker (Golang)

Проект финансового трекера для учета семейного бюджета и аналитики операций. 
Проект построен по слоистой архитектуре (repository → service → transport).

---

# Возможности

* CRUD операции финансовых транзакций (debit/credit)
* Фильтрация по времени (from / to)
* Сортировка по полям (id, amount, actor, category, type, operation_at)
* Пагинация (page / limit)
* Аналитика:

  * суммарная статистика (sum, avg, median, p90, count)
  * группировка по:

    * day / week / month / year
    * actor / category / type

* Минималистичный Web UI (HTML + JS)
* Экспорт операций и аналитики (JSON / CSV)

---

# Архитектура

```
cmd/            # точка входа в приложение - main.go
internal/
  model/        # доменные модели
  repository/   # слой взаимодействия(SQL) с БД 
  service/      # бизнес‑логика
  transport/    # HTTP handlers (Gin)
  web/          # UI (HTML)
  migrations/   # миграции
```

# Модель операции

```go
type Operation struct {
    ID          int64     `json:"id,omitempty"`
    Amount      int64     `json:"anount"` // копейки
    Actor       string    `json:"actor"`
    Category    string    `json:"category"`
    Type        string    `json:"type"` // debit / credit
    OperationAt time.Time `json:"operation_at"`
    CreatedAt   time.Time
    Description *string   `json:"description,omitempty"`
}
```

---

# Модель аналитики
```go
type AnalyticsQuantum struct { // возвращается в виде массива в поле Groups в AnalyticsSummary если в запросе указана группировка
	Key    string  `json:"key"`
	Sum    float64 `json:"sum"`    // сумма в копейках
	Avg    float64 `json:"avg"`    // среднее в копейках
	Count  int     `json:"count"`  // кол-во записей, на которых основано вычисление
	Median float64 `json:"median"` // медиана в копейках
	P90    float64 `json:"p90"`    // 90й перц в копейках
}

type AnalyticsSummary struct {
	Key    string             `json:"key,omitempty"` // поле, использованное для группировки
	Sum    float64            `json:"sum"`
	Avg    float64            `json:"avg"` // среднее в копейках
	Count  int                `json:"count"`
	Median float64            `json:"median"` // медиана в копейках
	P90    float64            `json:"p90"`    // 90й перц в копейках
	Groups []AnalyticsQuantum `json:"groups,omitempty"`
}
```

---

# HTTP API

## Создание операции

```
POST /operations
Content-Type: application/json
```

Пример:

```json
{
  "amount": 12345,
  "actor": "father",
  "category": "food",
  "type": "credit",
  "operation_at": "2026-01-01T12:00:00Z",
  "description": "Lunch"
}
```

---

## Получение списка операций

```
GET /operations
```

Query параметры:

```
from=2026-01-07T15:02:00.000Z
&to=2026-01-08T15:02:00.000Z
&order_by=id|amount|actor|category|type|operation_at
&asc=true | desc=true
&page=1
&limit=50
```

Ответ:
```json
[
    {
        "id": 1,
        "amount": -1000000,
        "actor": "mother",
        "category": "chores",
        "type": "credit",
        "operation_at": "2026-01-03T12:48:00Z",
        "CreatedAt": "2026-01-29T12:52:01.649416Z",
        "description": ""
    },
...
    {
        "id": n,
        "amount": 800000,
        "actor": "son",
        "category": "salary",
        "type": "debit",
        "operation_at": "2026-01-10T12:52:00Z",
        "CreatedAt": "2026-01-29T12:52:36.21694Z",
        "description": ""
    }
]
```

---

## Аналитика

```
GET /analytics
```

Query параметры:

```json
from=2026-01-07T15:02:00.000Z
&to=2026-01-08T15:02:00.000Z
&group_by=day|week|month|year|actor|category|type
&page=1
&limit=50
```

Ответ:

```json
{
  "sum": 80000,
  "avg": 5333.33,
  "median": 5000,
  "p90": 10000,
  "count": 15,
  "groups": [ ... ]
}
```

---

# Web UI

UI расположен по адресу:

```
internal/web/index.html
```

Функции:

* создание операций
* таблица операций с сортировкой и фильтрами
* аналитическая таблица с итоговой строкой TOTAL
* кнопки экспорта данных в CSV

---

# Экспорт данных

UI поддерживает экспорт:

* операций
* аналитики

Через параметры UI передаются фильтры и сортировка.

Backend возвращает файл с заголовками:

```
Content-Disposition: attachment
```

---

## Запуск проекта

1. Склонировать репозиторий
2. Создать файл окружения из примера(скопировать и переименовать), внести в него необходимые правки:

```bash
cp .env.example .env
```

3. Запустить приложение:

```bash
docker-compose up
```

4. Открыть в браузере:

```
http://localhost:8080/web
```

---
# Design Notes

* Деньги хранятся в int64 (копейки)
* float используется только для аналитики и UI
* Аналитика реализует OLAP‑паттерн Summary + Groups
* Query‑параметры сериализуются через form tags
