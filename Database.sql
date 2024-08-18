CREATE DATABASE tiktok_db;

-- Select the database to use
USE tiktok_db;

-- Create the messages table if it doesn't already exist
CREATE TABLE IF NOT EXISTS messages (
    id INT AUTO_INCREMENT PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the favorites table
CREATE TABLE IF NOT EXISTS favorites (
    id INT AUTO_INCREMENT PRIMARY KEY,  -- Primary key for the favorites table
    message_id INT UNIQUE,              -- Foreign key column referencing messages
    FOREIGN KEY (message_id) REFERENCES messages(id)   -- Foreign key constraint
);

ALTER TABLE favorites
DROP FOREIGN KEY favorites_ibfk_1,
ADD CONSTRAINT favorites_ibfk_1 FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE;


