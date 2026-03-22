-- Migration: Add is_active column to user table
-- Description: Adds status column to support user activation/deactivation
-- Date: 2024-01-15

ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;