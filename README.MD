## Finance Tracker (Golang)

Проект финансового трекера для учета семейного бюджета и аналитики операций. 
Проект построен по слоистой архитектуре (repository → service → transport) и демонстрирует практики backend‑разработки на Go.

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

# 🧱 Архитектура

```
cmd/
internal/
  model/        # доменные модели
  repository/   # SQL слой
  service/       # бизнес‑логика
  transport/     # HTTP handlers (Gin)
  web/            # UI (HTML)
  db/             # миграции
```

Принципы:

* Clean Architecture (без жёсткой привязки к DB и HTTP)
* Разделение ответственности по слоям
* Денежные значения хранятся в копейках (int64)

---

# 📦 Модель операции

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

# 🌐 HTTP API

## ➕ Создание операции

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

## 📄 Получение операций

```
GET /operations
```

Query параметры:

```
from=RFC3339
&to=RFC3339
&order_by=id|amount|actor|category|type|operation_at
&asc=true | desc=true
&page=1
&limit=50
```

---

## 📊 Аналитика

```
GET /analytics
```

Query параметры:

```
from=RFC3339
&to=RFC3339
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

# 🖥️ Web UI

UI расположен в:

```
internal/web/index.html
```

Функции:

* создание операций
* таблица операций с сортировкой и фильтрами
* аналитическая таблица с итоговой строкой TOTAL
* кнопки экспорта данных

Запуск UI:

```
http://localhost:8080
```

---

# 🐳 Запуск проекта

## 1️⃣ Конфигурация

Пример `.env`:

```
DB_DSN=postgres://user:pass@localhost:5432/finance?sslmode=disable
HTTP_PORT=8080
```

---

## 2️⃣ Миграции

```
make migrate-up
```

---

## 3️⃣ Запуск сервера

```
go run ./cmd/server
```

---

# 📤 Экспорт данных

UI поддерживает экспорт:

* операций
* аналитики

Через параметры UI передаются фильтры и сортировка.

Backend возвращает файл с заголовками:

```
Content-Disposition: attachment
```

---

# 🧪 Тестирование

Планируется:

* unit tests для service слоя
* integration tests с PostgreSQL
* HTTP contract tests

---

# 📈 Roadmap

* [ ] DELETE /operations/{id}
* [ ] UPDATE /operations/{id}
* [ ] Pagination UI
* [ ] Графики (Chart.js)
* [ ] Decimal money type (shopspring/decimal)
* [ ] Redis caching аналитики
* [ ] Docker Compose (Postgres + API + Nginx)
* [ ] Auth (JWT)

---

# 🧠 Design Notes

* Деньги хранятся в int64 (копейки)
* float используется только для аналитики и UI
* Аналитика реализует OLAP‑паттерн Summary + Groups
* Query‑параметры сериализуются через form tags

---

# 🧑‍💻 Автор

Фаниль — учебный проект по курсу "Горутиновый Golang".
