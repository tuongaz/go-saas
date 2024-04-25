create table if not exists auth_user
(
    id                                    varchar                  not null
        primary key,
    created_at                            timestamp with time zone not null,
    updated_at                            timestamp with time zone not null,
    email                                 varchar                  not null,
    password                              varchar                  not null,
    reset_password_code                   varchar                  not null,
    reset_password_code_expired_timestamp bigint                   not null,
    name                                  varchar                  not null
);

create unique index if not exists auth_user_email_key
    on auth_user (email);

create table if not exists account
(
    id                  varchar                  not null
        primary key,
    created_at          timestamp with time zone not null,
    updated_at          timestamp with time zone not null,
    name                varchar,
    first_name          varchar,
    last_name           varchar,
    communication_email varchar,
    avatar              varchar
);

create table if not exists organisation
(
    id         varchar                  not null
        primary key,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null
);

create table if not exists account_role
(
    id              varchar                  not null
        primary key,
    created_at      timestamp with time zone not null,
    updated_at      timestamp with time zone not null,
    role            varchar                  not null,
    account_id      varchar                  not null
        constraint account_role_account_account_roles
            references account
            on delete cascade,
    organisation_id varchar                  not null
        constraint account_role_organisation_account_roles
            references organisation
            on delete cascade
);

create unique index if not exists org_account_unique
    on account_role (account_id, organisation_id);

create table if not exists auth_provider
(
    id               varchar                  not null
        primary key,
    created_at       timestamp with time zone not null,
    updated_at       timestamp with time zone not null,
    provider_user_id varchar                  not null,
    provider         varchar                  not null,
    email            varchar,
    avatar           varchar,
    name             varchar,
    first_name       varchar,
    last_name        varchar,
    last_login       timestamp with time zone not null,
    account_id       varchar                  not null
        constraint auth_provider_account_providers
            references account
            on delete cascade
);

create unique index if not exists provider_user_id_provider_unq
    on auth_provider (provider_user_id, provider);

create table if not exists auth_token
(
    id              varchar                  not null
        primary key,
    created_at      timestamp with time zone not null,
    updated_at      timestamp with time zone not null,
    refresh_token   varchar                  not null,
    account_role_id varchar                  not null
        constraint auth_token_account_role_auth_tokens
            references account_role
            on delete cascade
);

create unique index if not exists auth_token_refresh_token_key
    on auth_token (refresh_token);
