CREATE TABLE IF NOT EXISTS public.products
(
    product_id SERIAL PRIMARY KEY,
    product_name character varying COLLATE pg_catalog."default" NOT NULL,
    product_description character varying COLLATE pg_catalog."default" NOT NULL,
    product_images character varying[] COLLATE pg_catalog."default" NOT NULL,
    product_price integer NOT NULL,
    compressed_product_images character varying[] COLLATE pg_catalog."default",
    created_at time with time zone,
    updated_at time with time zone,
    user_id integer NOT NULL,
	FOREIGN KEY (user_id) REFERENCES users (id)
)