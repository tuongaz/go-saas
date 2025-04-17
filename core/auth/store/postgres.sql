CREATE TABLE IF NOT EXISTS login_credentials_user
(
    id                                    VARCHAR PRIMARY KEY,
    email                                 VARCHAR                  NOT NULL,
    name                                  VARCHAR                  NOT NULL,
    password                              VARCHAR                  NOT NULL,
    reset_password_code                   VARCHAR                  NOT NULL,
    reset_password_code_expired_timestamp TIMESTAMP WITH TIME ZONE,
    created_at                            TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                            TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS login_credentials_user_email_unq
    ON login_credentials_user (email);



CREATE TABLE IF NOT EXISTS login_credentials_user_reset_password
(
    id         VARCHAR PRIMARY KEY,
    code       VARCHAR                  NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    receipt    VARCHAR,
    user_id    VARCHAR                  NOT NULL
        CONSTRAINT login_credentials_user_reset_password_user_id_fk
            REFERENCES login_credentials_user
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS login_credentials_user_reset_password_code_unq
    ON login_credentials_user_reset_password (code);


CREATE TABLE IF NOT EXISTS account
(
    id                  VARCHAR PRIMARY KEY,
    name                VARCHAR,
    first_name          VARCHAR,
    last_name           VARCHAR,
    communication_email VARCHAR,
    avatar              VARCHAR,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS organisation
(
    id         VARCHAR PRIMARY KEY,
    name       TEXT,
    description TEXT,
    avatar     TEXT,
    metadata   JSONB,
    owner_id   TEXT NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS organisation_account_role
(
    id              VARCHAR PRIMARY KEY,
    role            TEXT NOT NULL,
    account_id      VARCHAR NOT NULL REFERENCES account(id) ON DELETE CASCADE,
    organisation_id VARCHAR NOT NULL REFERENCES organisation(id) ON DELETE CASCADE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (organisation_id, account_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_owner_per_organisation
    ON organisation_account_role (organisation_id)
    WHERE role = 'OWNER';

CREATE TABLE IF NOT EXISTS login_provider
(
    id               VARCHAR PRIMARY KEY,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    provider_user_id VARCHAR                  NOT NULL,
    provider         VARCHAR                  NOT NULL,
    email            VARCHAR,
    avatar           VARCHAR,
    name             VARCHAR,
    first_name       VARCHAR,
    last_name        VARCHAR,
    last_login       TIMESTAMP WITH TIME ZONE NOT NULL,
    account_id       VARCHAR                  NOT NULL
        CONSTRAINT auth_provider_account_providers
            REFERENCES account
            ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS provider_user_id_provider_unq
    ON login_provider (provider_user_id, provider);

CREATE TABLE IF NOT EXISTS access_token
(
    id               VARCHAR PRIMARY KEY,
    refresh_token    VARCHAR                  NOT NULL,
    provider_user_id VARCHAR                  NOT NULL,
    device           VARCHAR                  NOT NULL,
    account_role_id  VARCHAR                  NOT NULL
        CONSTRAINT auth_token_account_role_auth_tokens
            REFERENCES organisation_account_role
            ON DELETE CASCADE,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_token_refresh_token_key
    ON access_token (refresh_token);
