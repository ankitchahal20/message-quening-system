CREATE TABLE IF NOT EXISTS public.users
(
    id serial PRIMARY KEY,
    name character varying COLLATE pg_catalog."default" NOT NULL,
    mobile character varying(10) COLLATE pg_catalog."default" NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    created_at time with time zone,
    updated_at time with time zone
)