# Workshop Service

## Local run

### Requirements

- Docker
  - Windows: https://www.docker.com/
- Make CLI
  - Windows: https://gnuwin32.sourceforge.net/downlinks/make.php

### Configuration

1. Make a copy of `.env.dist` file and name it `.env`.
2. Configure `.env` if you need.

### Run

1. Run `make rebuild` cmd.
2. To restart container run `make restart`.

> Keep in mind that `make rebuild` will also prune unused images and volumes.

- Workshop service is available at `http://localhost:8080/` by default.
- Swagger documentation is available at `http://localhost:14000/` by default.
- DB connection data is inside `.env` file.
- With `AUTH_DEBUG` mode you are available to impersonate any user / roles. 
  - Use `X-Debug-User-Roles` and `X-Debug-User-ID` headers to modify it. 
  - More info in swagger.

### Tests

To run tests run `make test_all` cmd.

## Project structure

- `cmd` / `internal` / `pkg` are code related directories.
- `docker` - contains all required resources and `Dockerfiles` for building container.
  - `docker/postgres/init` - contains initialization scripts for postgres database.  
- `docs` - contains auto-generated documentation.
- `migrations` - contains all required migrations.

## Docker Compose structure

- `workshop` is a core service.
- `postgres` is a main DB.
- `migrate` is a migrator image, that applies migrations when container starts.
- `redis` is a main cache. Used for rate-limiting purposes.

## Features

### Posts

- Get One
- Get List

> Only users with `POST_CREATOR` role can access following:

- Create
- Update
- Delete

### Interactions

> Only users with `USER` role can access this.

- Favorite post
- Unfavorite post
- Rate post

### Comments

- Get List

> Only users with `USER` role can access following:

- Add
- Update
- Delete

### Moderation

> Only users with `POST_MODERATOR` role can access this.

- Moderate Post
  - Approve / Decline
  - Note
  - Persistent moderation history
- Get Moderation Actions