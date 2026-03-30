# bookmarks

Create new Rails application with PostgreSQL database, configure database.yml for local and production environments. Run initial setup to verify database connection. Test: rails db:create and rails db:migrate should succeed.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/bookmarks
just install   # copies to ~/.local/bin/bookmarks
```

## Module Selections

- **rails**: Core full-stack web framework for the bookmarks application with MVC, routing, ORM, and asset pipeline (deterministic)
- **devise**: User authentication with email/password as required by the PRD (deterministic)
- **pundit**: Authorization to ensure users can only manage their own bookmarks (deterministic)
- **solid_queue**: Background job processing for URL metadata fetching and bookmark imports without Redis dependency (deterministic)
- **solid_cache**: Database-backed caching for single-server deployment without Redis (deterministic)
- **solid_cable**: Database-backed Action Cable adapter for potential real-time features, Rails 8 default (deterministic)
- **turbo**: Interactive UI with SPA-like page updates without custom JavaScript framework (deterministic)
- **stimulus**: Client-side behavior for keyboard shortcuts and form interactions on server-rendered HTML (deterministic)
- **searchkick**: Full-text search across bookmarks titles, descriptions, URLs, and tags with intelligent ranking (non-deterministic)
- **kamal**: Zero-downtime deployment to single VPS/bare metal server as specified in constraints (deterministic)
- **actionmailer**: Transactional emails for password reset and account notifications (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| rails | devise | det |
| rails | pundit | det |
| rails | solid_queue | det |
| rails | searchkick | non-det |
| rails | turbo | det |
| rails | stimulus | det |
| rails | kamal | det |
| rails | actionmailer | det |
| devise | pundit | det |
| solid_queue | rails | det |
| searchkick | rails | non-det |
| turbo | rails | det |
| stimulus | rails | det |

