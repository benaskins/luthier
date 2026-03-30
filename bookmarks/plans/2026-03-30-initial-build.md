# bookmarks — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize Rails application with PostgreSQL

Create new Rails application with PostgreSQL database, set up basic structure. Run rails new bookmarks --database=postgresql. Verify database connection and basic Rails setup.

Commit: `feat: initialize rails application with postgresql`

## Step 2 — Add and configure Devise authentication

Install Devise gem, run devise:install generator, create User model with devise modules (database_authenticatable, registerable, recoverable, rememberable, validatable). Configure email delivery with letter_opener in development. Test user registration and login flow.

Commit: `feat: add devise authentication with email/password`

## Step 3 — Create Bookmark and Tag models with associations

Generate Bookmark model with url, title, description, user_id fields. Generate Tag model with name field. Create join table bookmark_tags for many-to-many relationship. Add validations for URL format, required fields. Create scopes for user-specific queries and date sorting. Run migrations and verify associations.

Commit: `feat: create bookmark and tag models with associations`

## Step 4 — Implement Pundit authorization policies

Install Pundit gem, create ApplicationPolicy base class. Create BookmarkPolicy and TagPolicy with user-scoped permissions (only owner can edit/delete). Configure controllers to use authorize method. Test that users cannot access other users' bookmarks.

Commit: `feat: add pundit authorization policies for bookmarks`

## Step 5 — Create BookmarkController with CRUD actions

Generate BookmarkController with index, show, new, create, edit, update, destroy actions. Implement pagination using will_paginate or kaminari. Add pundit authorization to each action. Create views for listing, creating, editing bookmarks. Test CRUD operations with authorization.

Commit: `feat: implement bookmark CRUD controller with views`

## Step 6 — Add URL metadata fetching background job

Create FetchBookmarkMetadata job that accepts bookmark_id, fetches page content using HTTP client (faraday or httparty), extracts title and description using Nokogiri, updates bookmark record. Configure Solid Queue as default queue adapter. Test job execution and metadata extraction.

Commit: `feat: add background job for fetching bookmark metadata`

## Step 7 — Implement async bookmark creation with job dispatch

Modify BookmarkController#create to enqueue FetchBookmarkMetadata job after bookmark creation. Show user immediate feedback that bookmark is being processed. Add job status display if needed. Test that bookmark is created and job is queued.

Commit: `feat: implement async bookmark creation with metadata fetching`

## Step 8 — Add full-text search functionality

Implement full-text search across title, description, URL, and tags using PostgreSQL tsvector/tsquery. Add search_scope to Bookmark model. Create search action in BookmarksController that accepts query parameter. Display search results with highlighting. Test search returns relevant results.

Commit: `feat: add full-text search across bookmark fields`

## Step 9 — Add tag management interface

Create TagsController for tag CRUD. Add tag selection interface in bookmark form (autocomplete or multi-select). Implement tag filtering in bookmark index. Test adding, removing, and filtering by tags.

Commit: `feat: add tag management interface for bookmarks`

## Step 10 — Implement bookmark import from HTML/JSON

Create import action in BookmarksController that accepts HTML (Netscape bookmark format) or JSON file uploads. Parse bookmark data, create bookmarks for current user, skip duplicates. Add progress indicator for bulk imports. Test import from sample HTML and JSON files.

Commit: `feat: implement bookmark import from HTML and JSON`

## Step 11 — Add API endpoint for programmatic bookmark creation

Create API namespace with API key authentication (or Devise token auth). Implement POST /api/v1/bookmarks endpoint that accepts URL and optional tags. Return JSON response. Add API documentation. Test API endpoint with curl/Postman.

Commit: `feat: add API endpoint for programmatic bookmark creation`

## Step 12 — Add keyboard shortcuts with Stimulus

Create Stimulus controller for keyboard shortcuts (n for new bookmark, / for search focus, arrow keys for navigation). Add keyboard event listeners. Test keyboard navigation works without mouse.

Commit: `feat: add keyboard shortcuts for common actions`

## Step 13 — Add responsive CSS layout with Turbo

Create responsive layout with Bootstrap or Tailwind CSS. Implement Turbo Streams for live updates after bookmark creation/search. Ensure interface works on mobile devices. Test responsive breakpoints and Turbo Stream updates.

Commit: `feat: add responsive layout with Turbo Stream updates`

## Step 14 — Configure Kamal deployment

Create Kamal configuration for single-server deployment. Write Dockerfile with multi-stage build. Configure nginx, postgres, and app container. Set up environment variables for production. Test local Docker build and deployment.

Commit: `infra: configure kamal deployment for single server`

## Step 15 — Add comprehensive test suite

Write model tests for Bookmark, Tag, User with associations and validations. Write controller tests for all CRUD actions with authorization. Write system tests for user flows (create bookmark, search, import). Configure fixtures. Run full test suite.

Commit: `test: add comprehensive test suite for all features`

