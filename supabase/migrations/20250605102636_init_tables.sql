create sequence "public"."product_images_id_seq";

create sequence "public"."product_specs_id_seq";

create sequence "public"."products_id_seq";

create table "public"."product_images" (
    "id" bigint not null default nextval('product_images_id_seq'::regclass),
    "product_id" bigint not null,
    "url" text not null,
    "alt_text" character varying(255),
    "is_primary" boolean not null default false,
    "sort_order" integer default 0,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."product_specs" (
    "id" bigint not null default nextval('product_specs_id_seq'::regclass),
    "product_id" bigint not null,
    "spec_name" character varying(100) not null,
    "spec_value" character varying(255) not null,
    "sort_order" integer default 0,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."products" (
    "id" bigint not null default nextval('products_id_seq'::regclass),
    "uuid" character varying not null default gen_random_uuid(),
    "sku" character varying(100) not null,
    "name" character varying(255) not null,
    "price" numeric(10,2) not null,
    "original_price" numeric(10,2),
    "category" character varying(100) not null,
    "in_stock" boolean not null default true,
    "stock_count" integer not null default 0,
    "specs" jsonb default '[]'::jsonb,
    "description" text,
    "full_description" text,
    "is_active" boolean not null default true,
    "sort_order" integer default 0,
    "created_at" timestamp with time zone not null default now(),
    "updated_at" timestamp with time zone not null default now()
);


create table "public"."users" (
    "id" bigint generated always as identity not null,
    "name" text not null,
    "email" text not null,
    "created_at" timestamp with time zone default now(),
    "updated_at" timestamp with time zone default now(),
    "deleted_at" timestamp with time zone
);


alter sequence "public"."product_images_id_seq" owned by "public"."product_images"."id";

alter sequence "public"."product_specs_id_seq" owned by "public"."product_specs"."id";

alter sequence "public"."products_id_seq" owned by "public"."products"."id";

CREATE INDEX idx_product_images_is_primary ON public.product_images USING btree (is_primary);

CREATE INDEX idx_product_images_product_id ON public.product_images USING btree (product_id);

CREATE INDEX idx_product_specs_name ON public.product_specs USING btree (spec_name);

CREATE INDEX idx_product_specs_product_id ON public.product_specs USING btree (product_id);

CREATE INDEX idx_products_category ON public.products USING btree (category);

CREATE INDEX idx_products_created_at ON public.products USING btree (created_at);

CREATE INDEX idx_products_in_stock ON public.products USING btree (in_stock);

CREATE INDEX idx_products_sku ON public.products USING btree (sku);

CREATE INDEX idx_products_uuid ON public.products USING btree (uuid);

CREATE UNIQUE INDEX product_images_pkey ON public.product_images USING btree (id);

CREATE UNIQUE INDEX product_specs_pkey ON public.product_specs USING btree (id);

CREATE UNIQUE INDEX products_pkey ON public.products USING btree (id);

CREATE UNIQUE INDEX products_sku_key ON public.products USING btree (sku);

CREATE UNIQUE INDEX products_uuid_key ON public.products USING btree (uuid);

CREATE UNIQUE INDEX unique_product_spec ON public.product_specs USING btree (product_id, spec_name);

CREATE UNIQUE INDEX users_email_key ON public.users USING btree (email);

CREATE UNIQUE INDEX users_pkey ON public.users USING btree (id);

alter table "public"."product_images" add constraint "product_images_pkey" PRIMARY KEY using index "product_images_pkey";

alter table "public"."product_specs" add constraint "product_specs_pkey" PRIMARY KEY using index "product_specs_pkey";

alter table "public"."products" add constraint "products_pkey" PRIMARY KEY using index "products_pkey";

alter table "public"."users" add constraint "users_pkey" PRIMARY KEY using index "users_pkey";

alter table "public"."product_images" add constraint "product_images_product_id_fkey" FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE not valid;

alter table "public"."product_images" validate constraint "product_images_product_id_fkey";

alter table "public"."product_specs" add constraint "product_specs_product_id_fkey" FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE not valid;

alter table "public"."product_specs" validate constraint "product_specs_product_id_fkey";

alter table "public"."product_specs" add constraint "unique_product_spec" UNIQUE using index "unique_product_spec";

alter table "public"."products" add constraint "products_original_price_check" CHECK ((original_price >= (0)::numeric)) not valid;

alter table "public"."products" validate constraint "products_original_price_check";

alter table "public"."products" add constraint "products_price_check" CHECK ((price >= (0)::numeric)) not valid;

alter table "public"."products" validate constraint "products_price_check";

alter table "public"."products" add constraint "products_sku_key" UNIQUE using index "products_sku_key";

alter table "public"."products" add constraint "products_stock_count_check" CHECK ((stock_count >= 0)) not valid;

alter table "public"."products" validate constraint "products_stock_count_check";

alter table "public"."products" add constraint "products_uuid_key" UNIQUE using index "products_uuid_key";

alter table "public"."users" add constraint "users_email_key" UNIQUE using index "users_email_key";

grant delete on table "public"."product_images" to "anon";

grant insert on table "public"."product_images" to "anon";

grant references on table "public"."product_images" to "anon";

grant select on table "public"."product_images" to "anon";

grant trigger on table "public"."product_images" to "anon";

grant truncate on table "public"."product_images" to "anon";

grant update on table "public"."product_images" to "anon";

grant delete on table "public"."product_images" to "authenticated";

grant insert on table "public"."product_images" to "authenticated";

grant references on table "public"."product_images" to "authenticated";

grant select on table "public"."product_images" to "authenticated";

grant trigger on table "public"."product_images" to "authenticated";

grant truncate on table "public"."product_images" to "authenticated";

grant update on table "public"."product_images" to "authenticated";

grant delete on table "public"."product_images" to "service_role";

grant insert on table "public"."product_images" to "service_role";

grant references on table "public"."product_images" to "service_role";

grant select on table "public"."product_images" to "service_role";

grant trigger on table "public"."product_images" to "service_role";

grant truncate on table "public"."product_images" to "service_role";

grant update on table "public"."product_images" to "service_role";

grant delete on table "public"."product_specs" to "anon";

grant insert on table "public"."product_specs" to "anon";

grant references on table "public"."product_specs" to "anon";

grant select on table "public"."product_specs" to "anon";

grant trigger on table "public"."product_specs" to "anon";

grant truncate on table "public"."product_specs" to "anon";

grant update on table "public"."product_specs" to "anon";

grant delete on table "public"."product_specs" to "authenticated";

grant insert on table "public"."product_specs" to "authenticated";

grant references on table "public"."product_specs" to "authenticated";

grant select on table "public"."product_specs" to "authenticated";

grant trigger on table "public"."product_specs" to "authenticated";

grant truncate on table "public"."product_specs" to "authenticated";

grant update on table "public"."product_specs" to "authenticated";

grant delete on table "public"."product_specs" to "service_role";

grant insert on table "public"."product_specs" to "service_role";

grant references on table "public"."product_specs" to "service_role";

grant select on table "public"."product_specs" to "service_role";

grant trigger on table "public"."product_specs" to "service_role";

grant truncate on table "public"."product_specs" to "service_role";

grant update on table "public"."product_specs" to "service_role";

grant delete on table "public"."products" to "anon";

grant insert on table "public"."products" to "anon";

grant references on table "public"."products" to "anon";

grant select on table "public"."products" to "anon";

grant trigger on table "public"."products" to "anon";

grant truncate on table "public"."products" to "anon";

grant update on table "public"."products" to "anon";

grant delete on table "public"."products" to "authenticated";

grant insert on table "public"."products" to "authenticated";

grant references on table "public"."products" to "authenticated";

grant select on table "public"."products" to "authenticated";

grant trigger on table "public"."products" to "authenticated";

grant truncate on table "public"."products" to "authenticated";

grant update on table "public"."products" to "authenticated";

grant delete on table "public"."products" to "service_role";

grant insert on table "public"."products" to "service_role";

grant references on table "public"."products" to "service_role";

grant select on table "public"."products" to "service_role";

grant trigger on table "public"."products" to "service_role";

grant truncate on table "public"."products" to "service_role";

grant update on table "public"."products" to "service_role";

grant delete on table "public"."users" to "anon";

grant insert on table "public"."users" to "anon";

grant references on table "public"."users" to "anon";

grant select on table "public"."users" to "anon";

grant trigger on table "public"."users" to "anon";

grant truncate on table "public"."users" to "anon";

grant update on table "public"."users" to "anon";

grant delete on table "public"."users" to "authenticated";

grant insert on table "public"."users" to "authenticated";

grant references on table "public"."users" to "authenticated";

grant select on table "public"."users" to "authenticated";

grant trigger on table "public"."users" to "authenticated";

grant truncate on table "public"."users" to "authenticated";

grant update on table "public"."users" to "authenticated";

grant delete on table "public"."users" to "service_role";

grant insert on table "public"."users" to "service_role";

grant references on table "public"."users" to "service_role";

grant select on table "public"."users" to "service_role";

grant trigger on table "public"."users" to "service_role";

grant truncate on table "public"."users" to "service_role";

grant update on table "public"."users" to "service_role";


