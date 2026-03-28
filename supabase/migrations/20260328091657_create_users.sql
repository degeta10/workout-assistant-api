-- Enable required extension for UUID generation
create extension if not exists pgcrypto;

-- users table used by internal/auth/repository.go
create table if not exists public.users (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    email text not null unique,
    password_hash text not null,
    is_free_user boolean not null default true,
    created_at timestamptz not null default now()
);

-- Explicit index for common login lookup path
create index if not exists idx_users_email on public.users(email);