-- Migration 002: Seed user schema
-- Inserts prerequisite applications and roles required for artist registration.

INSERT INTO "user".application (code, name, created_by, updated_by)
VALUES
    ('ARTIST_PORTAL', 'Artist Portal', 'system', 'system'),
    ('CLIENT_PORTAL', 'Client Portal', 'system', 'system'),
    ('BACKOFFICE',    'Backoffice',    'system', 'system');

INSERT INTO "user".role (code, name, application_id, created_by, updated_by)
VALUES
    ('ARTIST',     'Artist',      (SELECT id FROM "user".application WHERE code = 'ARTIST_PORTAL'), 'system', 'system'),
    ('CLIENT',     'Client',      (SELECT id FROM "user".application WHERE code = 'CLIENT_PORTAL'), 'system', 'system'),
    ('ADMIN',      'Admin',       (SELECT id FROM "user".application WHERE code = 'BACKOFFICE'),    'system', 'system'),
    ('SUPERADMIN', 'Super Admin', (SELECT id FROM "user".application WHERE code = 'BACKOFFICE'),    'system', 'system');
