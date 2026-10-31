CREATE TABLE users (
 id uuid PRIMARY KEY, name text NOT NULL, email text NOT NULL, role text NOT NULL,
 version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
 created_by_id uuid NOT NULL, created_by_name text NOT NULL, created_at timestamptz NOT NULL,
 updated_by_id uuid, updated_by_name text, updated_at timestamptz,
 CHECK ((updated_by_id IS NULL AND updated_by_name IS NULL AND updated_at IS NULL) OR (updated_by_id IS NOT NULL AND updated_by_name IS NOT NULL AND updated_at IS NOT NULL))
);
CREATE TABLE books (
 id uuid PRIMARY KEY, title text NOT NULL, isbn text, description text, owner_id uuid NOT NULL, owner_name text NOT NULL,
 version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
 created_by_id uuid NOT NULL, created_by_name text NOT NULL, created_at timestamptz NOT NULL,
 updated_by_id uuid, updated_by_name text, updated_at timestamptz,
 CHECK ((updated_by_id IS NULL AND updated_by_name IS NULL AND updated_at IS NULL) OR (updated_by_id IS NOT NULL AND updated_by_name IS NOT NULL AND updated_at IS NOT NULL))
);
CREATE TABLE book_authors (
 book_id uuid NOT NULL REFERENCES books(id) ON DELETE CASCADE, name text NOT NULL, order_index integer NOT NULL,
 PRIMARY KEY (book_id, order_index), UNIQUE (book_id, name)
);
CREATE TABLE book_checkouts (
 id uuid PRIMARY KEY, book_id uuid NOT NULL REFERENCES books(id) ON DELETE CASCADE, user_id uuid NOT NULL, user_name text NOT NULL,
 checked_out_at timestamptz NOT NULL, returned_at timestamptz, position integer NOT NULL, UNIQUE (book_id, position)
);
