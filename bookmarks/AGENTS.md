# bookmarks

Create new Rails application with PostgreSQL database, set up basic structure. Run rails new bookmarks --database=postgresql. Verify database connection and basic Rails setup.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/bookmarks
just install   # copies to ~/.local/bin/bookmarks
```

## Module Selections

- **rails**: Full-stack web framework needed for MVC architecture, routing, views, and ORM integration. Core platform for the bookmark web application. (deterministic)
- **activerecord**: ORM with migrations, validations, and query interface needed for Bookmark, Tag, and User models with PostgreSQL storage. (deterministic)
- **actionpack**: Controllers and routing needed for web interface and API endpoints. Handles request/response cycle for bookmark CRUD operations. (deterministic)
- **activejob**: Background job framework needed for async URL fetching (title/description extraction) to meet 3-second save time requirement. (deterministic)
- **solid_queue**: Database-backed job queue (Rails 8 default) for single-server deployment without Redis. Required for async URL fetching jobs. (deterministic)
- **devise**: User authentication with email/password needed. Provides secure password hashing, CSRF protection, session management, and password reset flows. (deterministic)
- **pundit**: Authorization framework needed to ensure users can only access their own bookmarks. One policy class per model for resource-based access control. (deterministic)
- **turbo**: SPA-like page updates over WebSocket for fast, interactive UI without custom JavaScript. Used for bookmark list updates and search results. (deterministic)
- **stimulus**: Modest JavaScript framework for keyboard shortcuts and client-side interactions on server-rendered HTML. Required for keyboard navigation. (deterministic)
- **actiontext**: Rich text content support for bookmark descriptions. Stored as HTML with Active Storage attachments for any embedded content. (deterministic)
- **activestorage**: File upload support for bookmark attachments. Used by ActionText for storing images/files in descriptions. (deterministic)
- **actionmailer**: Email sending for Devise password reset and confirmation flows. Required for secure authentication user experience. (deterministic)
- **kamal**: Zero-downtime deployment to single VPS/bare metal via Docker. Required for single-server deployment constraint. (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| actionpack | devise | det |
| actionpack | pundit | det |
| actionpack | activerecord | det |
| activejob | activerecord | det |
| solid_queue | activejob | det |
| turbo | actionpack | det |
| stimulus | turbo | det |
| devise | actionmailer | det |
| actiontext | activestorage | det |
| kamal | rails | det |

