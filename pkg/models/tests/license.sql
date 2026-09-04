-- Test data only. The schema comes from testSchemaDDL in pkg/models/test_schema.go; do not add DDL here.
-- Serving `license`, the renamed licenses table.
INSERT INTO license (id, long_name, SPDX) VALUES (109, '3-Clause BSD License', 'BSD-3-Clause');
INSERT INTO license (id, long_name, SPDX) VALUES (552, 'Apache 2.0', 'Apache-2.0');
INSERT INTO license (id, long_name, SPDX) VALUES (850, 'Apache License 2.0', 'Apache-2.0');
INSERT INTO license (id, long_name, SPDX) VALUES (1378, 'BSD', '0BSD');
INSERT INTO license (id, long_name, SPDX) VALUES (4863, 'ISC', 'ISC');
INSERT INTO license (id, long_name, SPDX) VALUES (5236, 'LGPLv2.1+', 'LGPL-2.1-or-later');
INSERT INTO license (id, long_name, SPDX) VALUES (5614, 'MIT', 'MIT');
INSERT INTO license (id, long_name, SPDX) VALUES (9999, '', '');
