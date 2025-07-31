-- Удаляем поле photo_filename

SET search_path TO grafit;
ALTER TABLE workers DROP COLUMN IF EXISTS photo_filename;

SET search_path TO giredmet;
ALTER TABLE workers DROP COLUMN IF EXISTS photo_filename; 