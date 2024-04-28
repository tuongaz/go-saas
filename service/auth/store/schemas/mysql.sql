create table if not exists account
(
    id                  varchar(255) not null
        primary key,
    created_at          timestamp    not null,
    updated_at          timestamp    not null,
    name                varchar(255) null,
    first_name          varchar(255) null,
    last_name           varchar(255) null,
    communication_email varchar(255) null,
    avatar              varchar(255) null
) collate = utf8mb4_bin;

create table if not exists auth_provider
(
    id               varchar(255) not null
        primary key,
    created_at       timestamp    not null,
    updated_at       timestamp    not null,
    provider_user_id varchar(255) not null,
    provider         varchar(255) not null,
    email            varchar(255) null,
    avatar           varchar(255) null,
    name             varchar(255) null,
    first_name       varchar(255) null,
    last_name        varchar(255) null,
    last_login       timestamp    not null,
    account_id       varchar(255) not null,
    constraint provider_user_id_provider_unq
        unique (provider_user_id, provider),
    constraint auth_provider_account_providers
        foreign key (account_id) references account (id)
            on delete cascade
) collate = utf8mb4_bin;

create table if not exists auth_user
(
    id                                    varchar(255) not null
        primary key,
    created_at                            timestamp    not null,
    updated_at                            timestamp    not null,
    email                                 varchar(255) not null,
    password                              varchar(255) not null,
    reset_password_code                   varchar(255) not null,
    reset_password_code_expired_timestamp bigint       not null,
    name                                  varchar(255) not null,
    constraint email
        unique (email)
) collate = utf8mb4_bin;

create table if not exists organisation
(
    id         varchar(255) not null
        primary key,
    created_at timestamp    not null,
    updated_at timestamp    not null
) collate = utf8mb4_bin;

create table if not exists account_role
(
    id              varchar(255) not null
        primary key,
    created_at      timestamp    not null,
    updated_at      timestamp    not null,
    role            varchar(255) not null,
    account_id      varchar(255) not null,
    organisation_id varchar(255) not null,
    constraint org_account_unique
        unique (account_id, organisation_id),
    constraint account_role_account_account_roles
        foreign key (account_id) references account (id)
            on delete cascade,
    constraint account_role_organisation_account_roles
        foreign key (organisation_id) references organisation (id)
            on delete cascade
) collate = utf8mb4_bin;

create table if not exists auth_token
(
    id              varchar(255) not null
        primary key,
    created_at      timestamp    not null,
    updated_at      timestamp    not null,
    refresh_token   varchar(255) not null,
    account_role_id varchar(255) not null,
    constraint refresh_token
        unique (refresh_token),
    constraint auth_token_account_role_auth_tokens
        foreign key (account_role_id) references account_role (id)
            on delete cascade
) collate = utf8mb4_bin;
