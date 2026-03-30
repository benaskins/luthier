# bookmarks

Create new Rails application with PostgreSQL database. Configure database.yml for development/production/tenancy. Run initial setup. Test by running rails server and accessing root path.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/bookmarks
just install   # copies to ~/.local/bin/bookmarks
```

## Module Selections

- **rails**: Full-stack web framework required for MVC architecture, routing, ORM, and asset pipeline. Core platform for this bookmark management web application. (deterministic)
- **activerecord**: ORM needed for PostgreSQL database interactions, migrations, validations, and associations. Required for Bookmark, User, Tag models. (deterministic)
- **actionpack**: Controllers and routing required for handling HTTP requests, bookmark CRUD operations, and API endpoints. (deterministic)
- **activejob**: Background job framework needed for fetching URL metadata (title/description) asynchronously and processing bulk imports. (deterministic)
- **turbo**: SPA-like page updates over WebSocket for interactive bookmark management without custom JavaScript. Used for form submissions and search results. (deterministic)
- **stimulus**: Modest JavaScript framework for keyboard shortcuts (new bookmark, search, navigate) and client-side interactions on server-rendered HTML. (deterministic)
- **solid_queue**: Database-backed job queue (Rails 8 default) for background jobs. No Redis required, matches single-server deployment constraint. (deterministic)
- **devise**: Flexible authentication with email/password required for user login. Provides secure hashed passwords, CSRF protection, and session management. (deterministic)
- **pundit**: Authorization via plain Ruby policy objects for controlling access to bookmark CRUD operations (users can only manage their own bookmarks). (deterministic)
- **searchkick**: Full-text search across titles, descriptions, URLs, and tags required. PostgreSQL full-text search is insufficient for the search quality requirements (top 5 results must contain target link). Searchkick with Elasticsearch/OpenSearch provides intelligent search beyond simple SQL LIKE. (deterministic)
- **activestorage**: File uploads needed for importing bookmark exports (HTML/JSON files). Active Storage handles file processing and storage. (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| actionpack | devise | det |
| actionpack | pundit | det |
| actionpack | activerecord | det |
| activerecord | searchkick | det |
| activejob | activerecord | det |
| actionpack | turbo | det |
| actionpack | stimulus | det |
| actionpack | activestorage | det |

