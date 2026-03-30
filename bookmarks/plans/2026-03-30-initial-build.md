# bookmarks — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize Phoenix project

Create Phoenix project with mix phx.new bookmarks --database postgres. Add phoenix_live_view, phoenix_html, postgrex dependencies. Configure dev.exs with PostgreSQL connection. Verify server starts with mix phx.server.

Commit: `feat: initialize Phoenix project with PostgreSQL and LiveView`

## Step 2 — Set up user authentication

Add pow and pow_extensions dependencies. Configure Pow in lib/bookmarks/accounts.ex. Generate Pow migrations for users table. Create Accounts context with registration and session management. Add Pow routes in router. Test user registration and login flow.

Commit: `feat: add Pow authentication with email/password`

## Step 3 — Create Bookmarks context

Create Bookmarks context with Bookmark schema (title, description, url, user_id, inserted_at). Create tags join table (bookmarks_tags) for many-to-many relationship. Create tags table. Add changesets with validation. Test schema associations and changesets.

Commit: `feat: create Bookmarks context with schema and migrations`

## Step 4 — Implement full-text search

Add tsvector column to bookmarks table via migration. Create GIN index on tsvector. Implement search function in Bookmarks context using Ecto fragment for PostgreSQL full-text search across title, description, url, and tags. Test search queries with various keywords.

Commit: `feat: add full-text search with PostgreSQL tsvector`

## Step 5 — Set up Oban for async metadata fetching

Add oban dependency and configure Oban in application. Create FetchMetadata job that accepts bookmark_id, fetches URL, extracts title and description using HTTPoison and Floki. Update bookmark with fetched data. Test job enqueuing and execution.

Commit: `feat: add Oban job processor for URL metadata fetching`

## Step 6 — Create bookmark form LiveView

Create NewBookmarkLiveView with form for URL input. On submit, create bookmark record, enqueue FetchMetadata job, redirect to show page. Display loading state while metadata fetches. Test bookmark creation flow.

Commit: `feat: create LiveView for bookmark creation with auto-fetch`

## Step 7 — Create bookmark list LiveView

Create IndexLiveView with search input and paginated bookmark list. Implement search debounce and query params. Display bookmark cards with title, description, tags. Add sort by date added. Test search and pagination.

Commit: `feat: create LiveView for listing and searching bookmarks`

## Step 8 — Add tag management

Add tag input component for bookmark form. Implement tag selection, creation, and removal. Update BookmarkLiveView to handle tag changes. Test adding/removing tags on bookmarks.

Commit: `feat: add tag management UI`

## Step 9 — Add bookmark import feature

Create ImportLiveView with file upload for HTML/JSON bookmark exports. Parse HTML using Floki or JSON using Jason. Create bookmarks for each entry (enqueue metadata fetch). Handle import errors. Test with sample HTML and JSON files.

Commit: `feat: implement bookmark import from HTML/JSON`

## Step 10 — Create API endpoint

Create API controller at /api/bookmarks for POST requests. Accept JSON body with url, tags. Create bookmark and enqueue metadata fetch. Add API authentication (token-based or session). Test API endpoint with curl.

Commit: `feat: add API endpoint for programmatic bookmark creation`

## Step 11 — Add keyboard shortcuts

Implement global keyboard shortcuts: 'n' for new bookmark, '/' for search focus, 'j/k' for navigation. Use LiveView JS events. Add visible shortcut hints. Test shortcuts in browser.

Commit: `feat: add keyboard shortcuts for common actions`

## Step 12 — Add responsive styling

Configure Tailwind CSS in Phoenix. Create responsive layout for bookmarks list and forms. Ensure mobile-friendly design. Test responsive breakpoints.

Commit: `feat: add responsive CSS with Tailwind`

## Step 13 — Add bookmark deletion

Add delete button to bookmark cards in IndexLiveView. Implement delete action in Bookmarks context. Broadcast deletion via PubSub. Test bookmark deletion.

Commit: `feat: add delete bookmark functionality`

## Step 14 — Add unit tests

Write tests for Accounts context (registration, login). Write tests for Bookmarks context (create, read, update, delete, search, tags). Test Oban jobs. Achieve >80% coverage.

Commit: `test: add ExUnit tests for all contexts`

## Step 15 — Configure production

Configure prod.exs with production database, Oban queue, secret_key_base. Add runtime.exs for runtime configuration. Configure Oban.Pro for production queue tuning. Test production build with mix release.

Commit: `config: configure production environment with Oban`

