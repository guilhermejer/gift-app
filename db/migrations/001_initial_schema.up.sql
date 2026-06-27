-- Enable pgvector extension (requires pgvector installed on the server)
CREATE EXTENSION IF NOT EXISTS vector;

CREATE SCHEMA IF NOT EXISTS giftowner;

-- ─── users ───────────────────────────────────────────────────────────────────
CREATE TABLE giftowner.users (
    user_id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    active     BOOLEAN     NOT NULL DEFAULT true,
    plan_id    UUID,
    birth_date DATE,
    city       TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── friends ─────────────────────────────────────────────────────────────────
CREATE TYPE giftowner.gender AS ENUM ('male', 'female', 'other');

CREATE TABLE giftowner.friends (
    friend_id     UUID                NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID                NOT NULL REFERENCES giftowner.users(user_id) ON DELETE CASCADE,
    user_relation TEXT                NOT NULL,
    name          TEXT                NOT NULL,
    gender        giftowner.gender,
    birth_date    DATE,
    city          TEXT,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ         NOT NULL DEFAULT now()
);

CREATE INDEX ON giftowner.friends (user_id);

-- ─── profiles ────────────────────────────────────────────────────────────────
-- likes / dislikes são arrays de texto livres; embedding armazena o vetor RAG
CREATE TABLE giftowner.profiles (
    friend_id  UUID        PRIMARY KEY REFERENCES giftowner.friends(friend_id) ON DELETE CASCADE,
    likes      TEXT[]      NOT NULL DEFAULT '{}',
    dislikes   TEXT[]      NOT NULL DEFAULT '{}',
    embedding  vector(1536),          -- dimensão padrão OpenAI text-embedding-3-small
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ─── gifts ───────────────────────────────────────────────────────────────────
CREATE TABLE giftowner.gifts (
    gift_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    friend_id   UUID        NOT NULL REFERENCES giftowner.friends(friend_id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT,
    price_range TEXT,
    tags        TEXT[]      NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON giftowner.gifts (friend_id);

-- ─── reminders ───────────────────────────────────────────────────────────────
CREATE TYPE giftowner.reminder_type AS ENUM ('birthday', 'custom');

CREATE TABLE giftowner.reminders (
    reminder_id UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID                    NOT NULL REFERENCES giftowner.users(user_id) ON DELETE CASCADE,
    friend_id   UUID                    NOT NULL REFERENCES giftowner.friends(friend_id) ON DELETE CASCADE,
    type        giftowner.reminder_type NOT NULL,
    trigger_at  TIMESTAMPTZ             NOT NULL,
    message     TEXT,
    created_at  TIMESTAMPTZ             NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ             NOT NULL DEFAULT now()
);

CREATE INDEX ON giftowner.reminders (user_id);
CREATE INDEX ON giftowner.reminders (friend_id);
CREATE INDEX ON giftowner.reminders (trigger_at);
