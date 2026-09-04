-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `db_version`: no created_at column; the build stamp is db_release.
INSERT INTO db_version (package_name, schema_version, db_release) VALUES ('full', '1', '26.09.03.12.24');
