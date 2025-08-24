-- Initialize database tables
-- This file will be executed when the PostgreSQL container starts for the first time

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop existing tables if they exist to handle schema changes
DROP TABLE IF EXISTS repository_todos CASCADE;
DROP TABLE IF EXISTS repository_users CASCADE;
DROP TABLE IF EXISTS oauth_tokens CASCADE;
