# bookmarks — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize Rails application with PostgreSQL

Create new Rails application with PostgreSQL database, configure database.yml for local and production environments. Run initial setup to verify database connection. Test: rails db:create and rails db:migrate should succeed.

Commit: `feat: initialize rails app with postgresql database`

## Step 2 — Configure Devise authentication

Install and configure Devise gem with default modules (database_authenticatable, registerable, recoverable, rememberable, validatable). Generate Devise models and migrations. Add authentication helpers to ApplicationController. Test: User registration, login, logout, and password reset flows work correctly.

Commit: `feat: configure devise authentication`

## Step 3 — Create Bookmark model with validations

Generate Bookmark model with fields: url (string, indexed), title (string), description (text), user_id (foreign key). Add validations for url format and presence. Create belongs_to :user association. Test: Bookmark creation with valid/invalid data.

Commit: `feat: create bookmark model with validations`

## Step 4 — Implement Tag model and association

Create Tag model with name field. Implement has_many :through association between Bookmark and Tag via BookmarkTag join table. Add tag name uniqueness validation. Test: Creating bookmarks with multiple tags, querying by tags.

Commit: `feat: implement tag model and bookmark-tag association`

## Step 5 — Setup Pundit authorization

Install Pundit gem, generate ApplicationPolicy base class. Create BookmarkPolicy to ensure users can only manage their own bookmarks. Add before_action :authenticate_user! and authorization checks to controllers. Test: Users cannot access/modify other users' bookmarks.

Commit: `feat: setup pundit authorization for bookmarks`

## Step 6 — Configure Solid Queue for background jobs

Install and configure solid_queue gem. Generate queue tables and run migrations. Configure Active Job to use solid_queue adapter. Create initial job processor configuration. Test: Background jobs are queued and processed correctly.

Commit: `feat: configure solid_queue background jobs`

## Step 7 — Create URL fetcher job for metadata extraction

Create FetchBookmarkMetadata job that accepts a bookmark_id, fetches the URL content, extracts title and description using HTTP requests and HTML parsing. Update bookmark with fetched data. Test: Job successfully fetches and saves metadata from valid URLs.

Commit: `feat: create url metadata fetcher job`

## Step 8 — Setup Searchkick for full-text search

Install Searchkick gem, configure Elasticsearch connection. Add searchkick to Bookmark model with searchable fields: title, description, url, tags. Implement reindex logic after bookmark updates. Test: Search returns correct results across all indexed fields.

Commit: `feat: setup searchkick for full-text search`

## Step 9 — Create bookmark CRUD controllers

Generate BookmarksController with index, show, new, create, edit, update, destroy actions. Implement pagination with Kaminari or will_paginate. Add authorization checks to each action. Test: All CRUD operations work with proper authorization.

Commit: `feat: create bookmark CRUD controllers`

## Step 10 — Build bookmark form with auto-fetch

Create new/edit bookmark form with URL input that triggers metadata fetch via Turbo Frame or Stimulus controller. Display fetched title and description in real-time. Add tag input with autocomplete. Test: URL paste triggers fetch, form submits with all data.

Commit: `feat: build bookmark form with auto-fetch`

## Step 11 — Implement search interface

Create search page with query input and results display. Show bookmark cards with title, description, URL, and tags. Implement search filtering and pagination. Test: Search returns relevant results, displays correctly.

Commit: `feat: implement search interface`

## Step 12 — Add keyboard shortcuts with Stimulus

Implement Stimulus controller for keyboard shortcuts: 'n' for new bookmark, '/' for search focus, 'j/k' for navigation. Add visual hints for shortcuts. Test: Keyboard shortcuts trigger correct actions without mouse.

Commit: `feat: add keyboard shortcuts with stimulus`

## Step 13 — Create bookmark import functionality

Create ImportBookmarksController to handle HTML (Netscape format) and JSON bookmark imports. Parse uploaded files, create bookmarks and tags for user. Queue import job for large files. Test: Successfully imports bookmarks from HTML and JSON formats.

Commit: `feat: create bookmark import functionality`

## Step 14 — Implement API endpoint for programmatic bookmark creation

Create Api::BookmarksController with create action accepting URL, title, description, and tags. Implement API authentication (API key or token). Return JSON response. Test: API endpoint accepts valid requests and returns correct JSON.

Commit: `feat: implement api endpoint for bookmark creation`

## Step 15 — Configure Kamal deployment

Generate Kamal configuration files for single-server deployment. Create Dockerfile with multi-stage build. Configure database, cache, and queue services. Set up production environment variables. Test: Kamal deploy command works on target server.

Commit: `feat: configure kamal deployment`

## Step 16 — Setup Action Mailer for transactional emails

Configure Action Mailer with letter_opener for development and production SMTP settings. Generate password reset and account confirmation mailers. Test: Emails are sent and received correctly in both environments.

Commit: `feat: setup action mailer for transactional emails`

## Step 17 — Configure solid_cache for caching

Install and configure solid_cache gem. Set up cache store in development and production. Implement caching for search results and frequently accessed bookmarks. Test: Cache hits improve response times.

Commit: `feat: configure solid_cache for caching`

## Step 18 — Write comprehensive test suite

Add system tests with Capybara for critical user flows: registration, bookmark creation, search, import. Add model tests for validations and associations. Add controller tests for authorization. Test: All tests pass with >90% coverage.

Commit: `feat: write comprehensive test suite`

## Step 19 — Create responsive layout with Turbo

Build main application layout with navigation, search bar, and responsive design. Implement Turbo Frames for partial page updates. Add loading states and error handling. Test: Interface is responsive and updates work without full page reload.

Commit: `feat: create responsive layout with turbo`

