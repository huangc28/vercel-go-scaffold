-- name: GetUser :one
SELECT id, name, email, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users
  (name, email)
VALUES
  ($1, $2)
RETURNING id, name, email, created_at, updated_at, deleted_at;

-- name: ListUsers :many
SELECT id, name, email, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- Product queries
-- name: CreateProduct :one
INSERT INTO products
  (sku, name, price, category, stock_count, description)
VALUES
  ($1, $2, $3, $4, $5, $6)
RETURNING id, uuid, sku, name, price, original_price, category, in_stock, stock_count, specs, description, full_description, is_active, sort_order, created_at, updated_at;

-- name: GetProductBySKU :one
SELECT id, uuid, sku, name, price, original_price, category, in_stock, stock_count, specs, description, full_description, is_active, sort_order, created_at, updated_at
FROM products
WHERE sku = $1 AND is_active = true
LIMIT 1;

-- name: CreateProductSpec :one
INSERT INTO product_specs
  (product_id, spec_name, spec_value, sort_order)
VALUES
  ($1, $2, $3, $4)
RETURNING id, product_id, spec_name, spec_value, sort_order, created_at, updated_at;

-- name: CreateProductImage :one
INSERT INTO product_images
  (product_id, url, alt_text, is_primary, sort_order)
VALUES
  ($1, $2, $3, $4, $5)
RETURNING id, product_id, url, alt_text, is_primary, sort_order, created_at, updated_at;

-- name: GetProductSpecs :many
SELECT id, product_id, spec_name, spec_value, sort_order, created_at, updated_at
FROM product_specs
WHERE product_id = $1
ORDER BY sort_order;

-- name: GetProductImages :many
SELECT id, product_id, url, alt_text, is_primary, sort_order, created_at, updated_at
FROM product_images
WHERE product_id = $1
ORDER BY sort_order;