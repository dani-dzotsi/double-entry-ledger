-- +goose Up
CREATE TABLE currencies (
    code CHAR(3) PRIMARY KEY,
    name TEXT NOT NULL,
    exponent SMALLINT NOT NULL CHECK (exponent BETWEEN 0 AND 6),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT chk_currencies_code_upper CHECK (code = upper(code))
);

INSERT INTO currencies (code, name, exponent) VALUES
    ('USD', 'US Dollar', 2),
    ('GHS', 'Ghanaian Cedi', 2),
    ('EUR', 'Euro', 2),
    ('JPY', 'Japanese Yen', 0);

-- +goose Down
DROP TABLE currencies;
