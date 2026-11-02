-- Mock data for local development. Usage:
--   sqlite3 monsterdatabase.db < scripts/seed.sql
-- Re-running replaces all existing rows in the monsters table.
CREATE TABLE IF NOT EXISTS monsters ("id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,"gross_price" INTEGER,"gross_price_normalised" TEXT, "created_at" DATETIME DEFAULT CURRENT_TIMESTAMP);

DELETE FROM monsters;
DELETE FROM sqlite_sequence WHERE name = 'monsters';

-- One price per day for the last 180 days: a slow upward drift, a seasonal wave and some jitter
WITH RECURSIVE days(n) AS (
    SELECT 0 UNION ALL SELECT n + 1 FROM days WHERE n < 179
),
prices(n, price) AS (
    SELECT n, ROUND(29.9 + n * 0.03 + 1.5 * sin(n / 11.0) + ((n * 7919) % 11) / 10.0 - 0.5, 1) FROM days
)
INSERT INTO monsters (gross_price, gross_price_normalised, created_at)
SELECT printf('%.2f', price),
       CAST(ROUND(price * 100) AS INTEGER),
       datetime('now', '-' || (179 - n) || ' days', 'start of day', '+12 hours')
FROM prices;
