CREATE TABLE IF NOT EXISTS public.users
(
    id integer PRIMARY KEY,
    name character varying COLLATE pg_catalog."default",
    mobile character varying(10) COLLATE pg_catalog."default",
    latitude double precision,
    longitude double precision,
    created_at time with time zone,
    updated_at time with time zone
)