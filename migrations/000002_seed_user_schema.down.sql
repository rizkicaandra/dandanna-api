-- Remove seed data in reverse insertion order (roles before applications).

DELETE FROM "user".role        WHERE code IN ('ARTIST', 'CLIENT', 'ADMIN', 'SUPERADMIN');
DELETE FROM "user".application WHERE code IN ('ARTIST_PORTAL', 'CLIENT_PORTAL', 'BACKOFFICE');
