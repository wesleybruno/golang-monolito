ALTER TABLE IF EXISTS 
  users
ADD
  COLUMN role_id INT REFERENCES roles(id) DEFAULT 1;

UPDATE
    users
SET 
    role_id = (SELECT id from roles where name = 'user');

ALTER TABLE 
  users
ALTER COLUMN
    role_id DROP DEFAULT;


ALTER TABLE 
  users
ALTER COLUMN
    role_id 
SET 
    NOT NULL;