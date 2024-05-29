CREATE TABLE IF NOT EXISTS payment
(
    id                TEXT PRIMARY KEY,
    invoice_id        VARCHAR                  NOT NULL,
    payment_method_id VARCHAR                  NOT NULL,
    amount_in_cents   INT                      NOT NULL,
    currency          VARCHAR(3)               NOT NULL,
    status            VARCHAR(20)              NOT NULL,
    charge_data       TEXT,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS payment_method
(
    id                   TEXT PRIMARY KEY,
    account_id           VARCHAR     NOT NULL,
    provider             VARCHAR     NOT NULL,
    provider_customer_id VARCHAR     NOT NULL,
    is_default           BOOLEAN     NOT NULL DEFAULT FALSE,
    data                 TEXT        NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL,
    updated_at           TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS invoice
(
    id              VARCHAR PRIMARY KEY,
    account_id      VARCHAR                  NOT NULL,
    reference_id    VARCHAR                  NOT NULL,
    amount_in_cents INT                      NOT NULL,
    currency        VARCHAR(3)               NOT NULL,
    status          VARCHAR(20)              NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS stripe_customer
(
    account_id  VARCHAR PRIMARY KEY,
    customer_id VARCHAR NOT NULL
);



