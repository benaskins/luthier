# Bookmarks — Product Requirements Document

## Problem

Developers collect links from Twitter, Hacker News, GitHub, and Slack throughout the day but have no good way to save, tag, and search them later. Browser bookmarks are unorganised. Read-later apps are designed for articles, not technical links. Most saved links are never found again.

## Solution

A web application for saving, tagging, and searching bookmarks. Paste a URL, the app fetches the title and description automatically, you add tags, and it's saved. Full-text search across titles, descriptions, and tags. A clean, fast interface that stays out of the way.

## Target User

A developer who collects 5-20 links per day and wants to find them again weeks later. They value speed and keyboard shortcuts over visual design.

## User Stories

**As a user**, I want to paste a URL and have the title and description auto-populated so I don't have to type metadata manually.

**As a user**, I want to add tags to bookmarks so I can group related links.

**As a user**, I want to search across all my bookmarks by keyword so I can find links I saved weeks ago.

**As a user**, I want to log in with my email so my bookmarks are private and persistent.

**As a user**, I want to import bookmarks from a JSON or HTML export so I can migrate from other tools.

## Requirements

### Must Have

1. User authentication (email/password)
2. Create bookmark: accept URL, auto-fetch title and description from the page
3. Add, edit, and remove tags on bookmarks
4. Full-text search across title, description, URL, and tags
5. List bookmarks with pagination, sorted by date added
6. Delete bookmarks
7. Responsive web interface

### Should Have

8. Bulk import from browser bookmark export (HTML) or JSON
9. Keyboard shortcuts for common actions (new bookmark, search, navigate)
10. API endpoint for programmatic bookmark creation (for browser extensions)

### Won't Have (for now)

11. Browser extension
12. Social/sharing features
13. AI-powered tagging suggestions
14. Mobile native app

## Success Metrics

- Time from paste to saved bookmark: under 3 seconds
- Search returns relevant results: top 5 results contain the target link
- Page load time: under 500ms

## Constraints

- Must use PostgreSQL for storage
- Must work as a single-server deployment (no distributed infrastructure required)
- Authentication must be secure (hashed passwords, CSRF protection)
