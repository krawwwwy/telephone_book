-- Добавляем поле для хранения имени файла фотографии

-- Проверяем и создаем схему grafit если не существует
CREATE SCHEMA IF NOT EXISTS grafit;
SET search_path TO grafit;
-- Добавляем колонку только если таблица существует
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'grafit' AND table_name = 'workers') THEN
        ALTER TABLE workers ADD COLUMN IF NOT EXISTS photo_filename TEXT;
    END IF;
END $$;

-- Проверяем и создаем схему giredmet если не существует
CREATE SCHEMA IF NOT EXISTS giredmet;
SET search_path TO giredmet;
-- Добавляем колонку только если таблица существует
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'giredmet' AND table_name = 'workers') THEN
        ALTER TABLE workers ADD COLUMN IF NOT EXISTS photo_filename TEXT;
    END IF;
END $$; 