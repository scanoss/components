-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `SPDX_license_data`: membership here is what makes a licence
-- SPDX-approved, replacing the mining is_spdx flag.
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('BSD-3-Clause', '3-Clause BSD License');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('Apache-2.0', 'Apache 2.0');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('Apache-2.0', 'Apache License 2.0');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('0BSD', 'BSD');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('ISC', 'ISC');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('LGPL-2.1-or-later', 'LGPLv2.1+');
INSERT INTO SPDX_license_data (SPDX_id, name) VALUES ('MIT', 'MIT');
