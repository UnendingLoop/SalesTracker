CREATE TABLE IF NOT EXISTS category (
    id SERIAL PRIMARY KEY,
    cat_name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS family_members (
    id SERIAL PRIMARY KEY,
    fam_member TEXT UNIQUE
);

CREATE TYPE operation_type AS ENUM ('debit', 'credit');

CREATE TABLE IF NOT EXISTS operations (
    id SERIAL PRIMARY KEY,
    amount BIGINT NOT NULL CHECK (amount >= 0), --хранение в копейках
    actor_id INT,
    category_id INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    type operation_type NOT NULL, --доход или расход
    operation_at TIMESTAMPTZ NOT NULL,
    description TEXT,
    CONSTRAINT fk_operations_category FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE SET NULL,
    CONSTRAINT fk_operations_family_members FOREIGN KEY (actor_id) REFERENCES family_members (id) ON DELETE SET NULL
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_category_cat_name ON category (cat_name);

CREATE INDEX idx_operations_operation_at ON operations (operation_at);

CREATE INDEX idx_operations_category_id ON operations (category_id);

CREATE INDEX idx_operations_type ON operations (type);

CREATE INDEX idx_operations_actor_id ON operations (actor_id);

CREATE INDEX idx_operations_operation_at_type ON operations (operation_at, type);

INSERT INTO
    category (cat_name)
VALUES ('salary'),
    ('chores'),
    ('transport'),
    ('food'),
    ('entertainment'),
    ('health'),
    ('education'),
    ('presents'),
    ('electronics'),
    ('other'),
    ('communication')
ON CONFLICT (cat_name) DO NOTHING;

INSERT INTO
    family_members (fam_member)
VALUES ('mother'),
    ('father'),
    ('son'),
    ('daughter'),
    ('other')
ON CONFLICT (fam_member) DO NOTHING;