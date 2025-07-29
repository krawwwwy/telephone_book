-- Добавляем поле для хранения имени файла фотографии

SET search_path TO grafit;
ALTER TABLE workers ADD COLUMN IF NOT EXISTS photo_filename TEXT;

SET search_path TO giredmet;
ALTER TABLE workers ADD COLUMN IF NOT EXISTS photo_filename TEXT; 