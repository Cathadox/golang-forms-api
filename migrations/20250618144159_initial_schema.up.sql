CREATE SCHEMA IF NOT EXISTS public;
CREATE SCHEMA IF NOT EXISTS authz;

CREATE TABLE IF NOT EXISTS public.form
(
    id                     UUID    NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name                   TEXT    NOT NULL,
    open_tracking_enabled  BOOLEAN NOT NULL DEFAULT FALSE,
    click_tracking_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS public.form_steps
(
    id      UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name    TEXT NOT NULL,
    content TEXT NOT NULL,
    step_order INTEGER NOT NULL,
    form_id UUID NOT NULL,
    CONSTRAINT fk_form
        FOREIGN KEY(form_id) REFERENCES form(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS authz.credentials
(
    id       UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    username     TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

-- Index for performance optimization
CREATE INDEX IF NOT EXISTS idx_form_steps_form_id ON public.form_steps(form_id);
