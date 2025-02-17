
-- SQL миграция для создания таблиц
CREATE TABLE posts (
  id VARCHAR(255) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  authorID VARCHAR(255) NOT NULL,
  createdAt VARCHAR(255) DEFAULT CURRENT_TIMESTAMP::VARCHAR,
  commentsDisabled BOOLEAN DEFAULT FALSE
);

CREATE TABLE comments (
  id VARCHAR(255) PRIMARY KEY,
  postID VARCHAR(255) NOT NULL,
  parentID VARCHAR(255),
  authorID VARCHAR(255) NOT NULL,
  content TEXT NOT NULL,
  createdAt VARCHAR(255) DEFAULT CURRENT_TIMESTAMP::VARCHAR,
  FOREIGN KEY (postID) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY (parentID) REFERENCES comments(id) ON DELETE CASCADE
);


-- Удаление таблиц при откате миграции
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;