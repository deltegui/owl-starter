create table users (
    id            bigserial primary key,
    name          text        not null,
    email         text        not null unique,
    password_hash text        not null,
    role          int         not null,
    created_at    timestamptz not null default now()
);

-- The Manager keeps only the hash of the session id (key).
create table sessions (
    key        text        primary key,
    user_id    bigint      not null references users (id) on delete cascade,
    value      jsonb       not null,
    created_at timestamptz not null,
    expires_at timestamptz not null
);

create index sessions_user_id on sessions (user_id);
create index sessions_expires_at on sessions (expires_at);
