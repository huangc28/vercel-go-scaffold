drop index if exists "public"."idx_products_in_stock";

alter table "public"."products" drop column "description";

alter table "public"."products" drop column "full_description";

alter table "public"."products" drop column "in_stock";

alter table "public"."products" drop column "is_active";

alter table "public"."products" drop column "sort_order";

alter table "public"."products" add column "full_desc" text;

alter table "public"."products" add column "reserved_count" integer not null default 0;

alter table "public"."products" add column "short_desc" text;

alter table "public"."products" add constraint "products_reserved_count_check" CHECK ((reserved_count >= 0)) not valid;

alter table "public"."products" validate constraint "products_reserved_count_check";


