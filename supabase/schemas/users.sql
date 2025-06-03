CREATE TABLE users (
  id bigint primary key generated always as identity,
  name text not null,
  email text unique not null,
  created_at timestamptz default now(),
  updated_at timestamptz default now(),
  deleted_at timestamptz default null
);