# bookmarks — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize Rails application with PostgreSQL

Create new Rails application with PostgreSQL database. Configure database.yml for development/production/tenancy. Run initial setup. Test by running rails server and accessing root path.

Commit: `feat: initialize rails application with postgresql`

## Step 2 — Add and configure Devise authentication

Install devise gem, run generator to create User model with email/password authentication. Configure devise modules (database_authenticatable, registerable, recoverable, rememberable, validatable). Add sessions and registrations routes. Test by creating a user account and logging in.

Commit: `feat: add devise authentication with email and password`

## Step 3 — Create Bookmark and Tag models with associations

Generate Bookmark model with url, title, description, user_id fields. Generate Tag model. Create join table for bookmark-tag many-to-many relationship. Add validations for URL format, presence of title. Test by creating bookmarks and tags with associations.

Commit: `feat: create bookmark and tag models with associations`

## Step 4 — Set up Pundit authorization policies

Install pundit gem. Create ApplicationPolicy as base class. Create BookmarkPolicy to ensure users can only manage their own bookmarks. Configure controller to include Pundit and use authorize calls. Test by attempting to access another user's bookmark.

Commit: `feat: set up pundit authorization policies for bookmarks`

## Step 5 — Configure Solid Queue for background jobs

Install and configure solid_queue gem. Create config/queueers.yml. Set up Solid Queue as the default Active Job backend. Test by running a background job and verifying it processes via solid_queue.

Commit: `feat: configure solid_queue as job backend`

## Step 6 — Create Bookmark CRUD controllers with Turbo

Generate BookmarksController with index, show, new, create, edit, update, destroy actions. Use Turbo for form submissions and responses. Implement pagination with will_paginate or Kaminari. Add keyboard shortcuts via Stimulus for quick navigation. Test by creating, editing, and deleting bookmarks through the UI.

Commit: `feat: create bookmark CRUD controllers with Turbo integration`

## Step 7 — Implement URL metadata fetching background job

Create FetchBookmarkMetadata job that takes a URL and extracts title and description using Nokogiri and OpenURI. Configure Active Job to process this job in background. Update Bookmark creation to enqueue this job. Test by creating a bookmark and verifying metadata is populated after job runs.

Commit: `feat: implement URL metadata fetching job`

## Step 8 — Integrate Searchkick for full-text search

Install searchkick gem and configure Elasticsearch connection. Add searchkick to Bookmark model with searchable fields: title, description, url, tags. Create search results view with Turbo Streams for live updates. Test by searching for keywords and verifying results include relevant bookmarks.

Commit: `feat: integrate searchkick for full-text search`

## Step 9 — Add bookmark import from HTML/JSON files

Create ImportBookmarksController to handle file uploads. Use Active Storage for file attachment. Parse HTML (Netscape bookmark format) and JSON exports. Create background job to process imports and create bookmarks in bulk. Test by uploading a bookmark export file and verifying bookmarks are created.

Commit: `feat: add bookmark import from HTML/JSON files`

## Step 10 — Add API endpoint for bookmark creation

Create API namespace with BookmarksController for JSON responses. Implement create action that accepts URL and optional tags. Add API authentication (token-based or session-based). Test by making POST request to /api/v1/bookmarks with curl or similar tool.

Commit: `feat: add API endpoint for programmatic bookmark creation`

## Step 11 — Implement keyboard shortcuts with Stimulus

Create Stimulus controllers for keyboard shortcuts: Ctrl+N for new bookmark, Ctrl+F for focus search, arrow keys for navigation. Add event listeners and integrate with existing views. Test by using keyboard shortcuts to navigate and create bookmarks without mouse.

Commit: `feat: implement keyboard shortcuts with Stimulus`

## Step 12 — Add responsive layout and views

Create application layout with responsive CSS using Tailwind or Bootstrap. Design bookmark list view with pagination, search form, and quick actions. Create bookmark form partials for new/edit. Test by viewing on mobile and desktop browsers, ensuring responsive behavior.

Commit: `feat: add responsive layout and views`

## Step 13 — Add tests for bookmark functionality

Write Minitest unit tests for Bookmark and Tag models with validations. Write controller tests for BookmarksController actions. Write system tests with Capybara for user flows (create bookmark, search, import). Test by running rails test and ensuring all tests pass.

Commit: `feat: add tests for all bookmark functionality`

## Step 14 — Configure Kamal for single-server deployment

Install kamal gem and generate configuration. Create Dockerfile with multi-stage build for Rails app. Configure kamal.yml for single VPS deployment. Set up database migrations and asset precompilation in deploy process. Test by running kamal deploy to a test server.

Commit: `infra: configure kamal for single-server deployment`

