SELECT 'CREATE DATABASE test_workshop' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'test_workshop')\gexec
