SELECT 'CREATE DATABASE workshop' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'workshop')\gexec
