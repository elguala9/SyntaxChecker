CREATE TABLE users (
  id   BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  age  INT DEFAULT 0
);

SELECT id, name FROM users WHERE age >= 18 ORDER BY name ASC;

INSERT INTO users (name, age) VALUES ('Alice', 30), ('Bob', 25);
