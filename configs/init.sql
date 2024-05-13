-- Проверяем существование базы данных db_commands и создаем ее при необходимости
CREATE DATABASE IF NOT EXISTS db_commands;

-- Создаем таблицу commands, если она не существует
CREATE TABLE IF NOT EXISTS commands (
    id SERIAL PRIMARY KEY,
    command TEXT NOT NULL,
    result TEXT
);