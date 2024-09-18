-- +goose Up
-- +goose StatementBegin

DO
$$
BEGIN
  IF NOT EXISTS (SELECT * FROM pg_roles WHERE rolname = 'alloydbsuperuser') THEN
     CREATE ROLE alloydbsuperuser;
  END IF;
  IF NOT EXISTS (SELECT * FROM pg_roles WHERE rolname = 'readaccess') THEN
     CREATE ROLE readaccess;
  END IF;
END
$$
;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'alloydbsuperuser') THEN
     DROP ROLE alloydbsuperuser;
  END IF;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'readaccess') THEN
     DROP ROLE readaccess;
  END IF;
END
$$;

-- +goose StatementEnd
