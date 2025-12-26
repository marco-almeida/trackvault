# Spotify Backup Tool — Project Specification

## 1. Problem Statement

Loss of Spotify account access (e.g., ban) results in permanent loss of liked tracks and playlists, including private playlists.

---

## 2. Objectives

- Automatically back up Spotify user data (liked tracks and playlists)
- Preserve private playlists by authenticating with the user’s account
- Allow restoration of backed-up data into a new Spotify account
- Minimize manual user intervention

---

## 3. Functional Requirements

- Authenticate users via Spotify OAuth
- Fetch user liked tracks and playlists (including private playlists)
- Export backup data to a platform-agnostic JSON format
- Import data from JSON files to recreate likes and playlists
- Support repeated backup runs over time

---

## 4. Architecture & Design Decisions

- The system will be implemented as a **reusable core library** with a **CLI frontend**
- The CLI is responsible only for orchestration and user interaction
- All Spotify interaction, backup logic, and restore logic reside in the core library
- Persistent state will be stored **exclusively using files**, without any database
- Backups will be represented as **versioned JSON files**
- Each playlist is backed up into a separate file

---

## 5. Backup Model

A **backup run** represents a snapshot of all accessible user data at a specific point in time.

A backup run:
- Is associated with a single authenticated user
- Produces a dedicated directory on disk
- Contains:
  - One backup file for liked tracks
  - One backup file per playlist
- Is snapshot-based (no incremental backups)
- May complete fully or partially

Multiple backup runs may coexist on disk, but a single command execution produces only one run.

Restore operations may accept:
- A directory representing a single backup run
- A list of individual backup files

---

## 6. File Layout

Each backup run is stored in its own directory, typically named using a datetime-based identifier.

Example structure:

```
backups/
 user-id/
  2025-01-15T22-40-00/
   metadata.json
   likes.json
   playlist_<id>.json
   playlist_<id>.json
```


### Metadata File (`metadata.json`)

Each backup run must include a metadata file containing:
- Backup format version
- Timestamp
- Opaque user identifier
- List of expected backup files
- Backup status: `complete`, `partial`, or `failed`

### File Writing Rules

- Backup files must be written atomically (write to temp file, then rename)
- Corruption is detected at runtime when reading files
- Corruption or parse errors must be logged

---

## 7. Non-Functional Requirements

- Backups should run automatically (e.g., via scheduled CLI execution)
- Partial failures must not invalidate existing valid backup files
- Backup format must be stable and versioned
- Credentials and tokens must be stored securely
- Restoration must avoid duplicate likes and playlists

---

## 8. Core Library API (Conceptual)

The core library exposes high-level services:

### Authentication
- Authenticate with a streaming service provider
- Return an authenticated session object

### Backup
- Run a backup using an authenticated session
- Write backup files to a destination directory
- Return a backup result indicating success, partial success, or failure

### Restore
- Restore data into a target authenticated account
- Accept one or more backup sources (files or directories)
- Skip invalid backup files while logging errors

The core library is provider-agnostic by design, although only Spotify is supported initially.

---

## 9. Failure & Recovery Rules

### Backup
- Failure to back up a single playlist:
  - Does not abort the entire backup run
  - Marks the run as `partial`
- Authentication failure (e.g., token refresh failure):
  - Aborts the backup run
  - Marks the run as `failed`
- File system or IO failure:
  - Aborts the backup run
  - Marks the run as `failed`

### Restore
- Invalid or corrupted backup files:
  - Are skipped
  - Errors are logged
- Duplicate tracks or playlists:
  - Are ignored
  - Skipped items are logged

---

## 10. Data Format

- Backup data is stored as platform-agnostic JSON
- JSON schemas should avoid Spotify-specific assumptions where possible
- The format must support future migration to other streaming platforms

---

## 11. Future Enhancements

- Migration to other streaming platforms
- Differential or incremental backups
- Cloud storage backends
- UI for backup history and restore previews

---

## 12. Open Questions

- Final naming conventions for backup files
- Backup scheduling strategy (cron, systemd, etc.)
- Choice of programming language
