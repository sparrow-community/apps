-- +goose Up
-- +goose StatementBegin
create table if not exists "users"
(
    id          varchar(36)            not null constraint user_pk primary key,
    avatar      varchar(255) default ''::character varying,
    nickname    varchar(255) default ''::character varying not null,
    disabled_at bigint       default 0 not null,
    deleted_at  bigint       default 0 not null,
    created_at  bigint       default 0 not null
);

create table if not exists user_credentials
(
    id              serial                not null constraint user_credential_pk primary key,
    user_id         varchar(36)           not null constraint user_credential_user_fk references "users",
    "type"          integer               not null,
    identifier      varchar(255)          not null constraint user_credential_identifier_uk unique,
    secret_data     varchar(255),
    salt            bytea,
    verified        boolean default false not null,
    verification_at bigint,
    created_at      bigint       default 0 not null,
    constraint user_credential_user_type_uk unique (user_id, type)
);

-- +goose StatementEnd

-- +goose Down
drop table user_credentials;
drop table "users";