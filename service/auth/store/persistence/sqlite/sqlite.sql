create table account
(
    id                  text     not null
        primary key,
    created_at          datetime not null,
    updated_at          datetime not null,
    name                text,
    first_name          text,
    last_name           text,
    communication_email text,
    avatar              text
);

create table auth_provider
(
    id               text     not null
        primary key,
    created_at       datetime not null,
    updated_at       datetime not null,
    provider_user_id text     not null,
    provider         text     not null,
    email            text,
    avatar           text,
    name             text,
    first_name       text,
    last_name        text,
    last_login       datetime not null,
    account_id       text     not null
        constraint auth_provider_account_providers
            references account
            on delete cascade
);

create unique index provider_user_id_provider_unq
    on auth_provider (provider_user_id, provider);

create table auth_user
(
    id                                    text     not null
        primary key,
    created_at                            datetime not null,
    updated_at                            datetime not null,
    email                                 text     not null,
    password                              text     not null,
    reset_password_code                   text     not null,
    reset_password_code_expired_timestamp integer  not null,
    name                                  text     not null
);

create unique index auth_user_email_key
    on auth_user (email);

create table organisation
(
    id         text     not null
        primary key,
    created_at datetime not null,
    updated_at datetime not null
);

create table account_role
(
    id              text     not null
        primary key,
    created_at      datetime not null,
    updated_at      datetime not null,
    role            text     not null,
    account_id      text     not null
        constraint account_role_account_account_roles
            references account
            on delete cascade,
    organisation_id text     not null
        constraint account_role_organisation_account_roles
            references organisation
            on delete cascade
);

create unique index org_account_unique
    on account_role (account_id, organisation_id);

create table auth_token
(
    id              text     not null
        primary key,
    created_at      datetime not null,
    updated_at      datetime not null,
    refresh_token   text     not null,
    account_role_id text     not null
        constraint auth_token_account_role_auth_tokens
            references account_role
            on delete cascade
);

create unique index auth_token_refresh_token_key
    on auth_token (refresh_token);
