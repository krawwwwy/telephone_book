-- Удаление данных в обратном порядке
SET search_path TO grafit;
DELETE FROM workers;
DELETE FROM sections;
DELETE FROM departments;

SET search_path TO giredmet;
DELETE FROM workers;
DELETE FROM sections;
DELETE FROM departments;
