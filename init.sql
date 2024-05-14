-- Проверяем существование базы данных db_commands и создаем ее при необходимости
SELECT 'CREATE DATABASE db_commands' WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'db_commands');

-- Создаем таблицу commands, если она не существует
CREATE TABLE IF NOT EXISTS commands (
    id SERIAL PRIMARY KEY,
    command TEXT NOT NULL,
    result TEXT
);