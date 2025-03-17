# TryTraGo API Documentation

## Overview

TryTraGo is a multilanguage dictionary server that provides a comprehensive REST API for managing and accessing dictionary entries, meanings, and translations.

- **Base URL:** `/api/v1`
- **Supported formats:** JSON
- **Authentication:** JWT-based authentication

## Authentication

### Authentication Endpoints

#### Register a new user

```
POST /auth/register
```

Creates a new user account.

**Request Body:**
```json
{
  "username": "user123",
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "email": "user@example.com",
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### Login

```
POST /auth/login
```

Authenticates a user and returns JWT tokens.

**Request Body:**
```json
{
  "username": "user123",
  "password": "securepassword"
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "email": "user@example.com"
  }
}
```

#### Refresh Token

```
POST /auth/refresh
```

Refreshes an access token using a refresh token.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:** `200 OK`
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600,
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "email": "user@example.com"
  }
}
```

### Using Authentication

For protected endpoints, include the JWT token in the Authorization header:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## Dictionary Entries

### Public Endpoints

#### List Entries

```
GET /entries
```

Retrieves a list of dictionary entries.

**Query Parameters:**
- `limit`: Maximum number of entries to return (default: 20, max: 100)
- `offset`: Number of entries to skip (for pagination)
- `sort_by`: Field to sort by (`word`, `created_at`, `updated_at`)
- `sort_desc`: If true, sort in descending order (default: false)
- `word_filter`: Filter entries by word (partial match)
- `type`: Filter entries by type (`WORD`, `COMPOUND_WORD`, `PHRASE`)

**Response:** `200 OK`
```json
{
  "entries": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "word": "example",
      "type": "WORD",
      "pronunciation": "ɪɡˈzæmpəl",
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z"
    },
    {
      "id": "223e4567-e89b-12d3-a456-426614174000",
      "word": "dictionary",
      "type": "WORD",
      "pronunciation": "ˈdɪkʃəˌnɛri",
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z"
    }
  ],
  "total": 2,
  "limit": 20,
  "offset": 0
}
```

#### Get Entry

```
GET /entries/{id}
```

Retrieves a specific dictionary entry by ID.

**Path Parameters:**
- `id`: UUID of the entry

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "word": "example",
  "type": "WORD",
  "pronunciation": "ɪɡˈzæmpəl",
  "meanings": [
    {
      "id": "323e4567-e89b-12d3-a456-426614174000",
      "entry_id": "123e4567-e89b-12d3-a456-426614174000",
      "part_of_speech": "noun",
      "description": "a thing characteristic of its kind",
      "examples": [
        {
          "id": "423e4567-e89b-12d3-a456-426614174000",
          "text": "this is an example of a good dictionary entry",
          "context": "educational",
          "created_at": "2023-04-10T15:30:45Z",
          "updated_at": "2023-04-10T15:30:45Z"
        }
      ],
      "translations": [...],
      "comments": [...],
      "likes_count": 5,
      "current_user_liked": false,
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z"
    }
  ],
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### List Meanings

```
GET /entries/{id}/meanings
```

Retrieves the meanings of a specific entry.

**Path Parameters:**
- `id`: UUID of the entry

**Response:** `200 OK`
```json
{
  "meanings": [
    {
      "id": "323e4567-e89b-12d3-a456-426614174000",
      "entry_id": "123e4567-e89b-12d3-a456-426614174000",
      "part_of_speech": "noun",
      "description": "a thing characteristic of its kind",
      "examples": [...],
      "translations": [...],
      "comments": [...],
      "likes_count": 5,
      "current_user_liked": false,
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z"
    }
  ],
  "total": 1
}
```

#### Get Meaning

```
GET /entries/{entryId}/meanings/{meaningId}
```

Retrieves a specific meaning of an entry.

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Response:** `200 OK`
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174000",
  "entry_id": "123e4567-e89b-12d3-a456-426614174000",
  "part_of_speech": "noun",
  "description": "a thing characteristic of its kind",
  "examples": [...],
  "translations": [...],
  "comments": [...],
  "likes_count": 5,
  "current_user_liked": false,
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### List Translations

```
GET /entries/{entryId}/meanings/{meaningId}/translations
```

Retrieves translations for a specific meaning.

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Query Parameters:**
- `language_id`: Filter translations by language (ISO 639-1 code)

**Response:** `200 OK`
```json
{
  "translations": [
    {
      "id": "523e4567-e89b-12d3-a456-426614174000",
      "meaning_id": "323e4567-e89b-12d3-a456-426614174000",
      "language_id": "fr",
      "text": "exemple",
      "comments": [...],
      "likes_count": 3,
      "current_user_liked": false,
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z",
      "created_by": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "username": "user123",
        "avatar": "https://example.com/avatar.jpg"
      }
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### Protected Endpoints

#### Create Entry

```
POST /entries
```

Creates a new dictionary entry.

**Authentication:** Required

**Request Body:**
```json
{
  "word": "example",
  "type": "WORD",
  "pronunciation": "ɪɡˈzæmpəl"
}
```

**Response:** `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "word": "example",
  "type": "WORD",
  "pronunciation": "ɪɡˈzæmpəl",
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### Update Entry

```
PUT /entries/{id}
```

Updates an existing dictionary entry.

**Authentication:** Required

**Path Parameters:**
- `id`: UUID of the entry

**Request Body:**
```json
{
  "word": "updated example",
  "type": "WORD",
  "pronunciation": "ʌpˈdeɪtɪd ɪɡˈzæmpəl"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "word": "updated example",
  "type": "WORD",
  "pronunciation": "ʌpˈdeɪtɪd ɪɡˈzæmpəl",
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T16:45:12Z"
}
```

#### Delete Entry

```
DELETE /entries/{id}
```

Deletes a dictionary entry.

**Authentication:** Required

**Path Parameters:**
- `id`: UUID of the entry

**Response:** `204 No Content`

#### Add Meaning

```
POST /entries/{entryId}/meanings
```

Adds a new meaning to an entry.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry

**Request Body:**
```json
{
  "part_of_speech_id": "723e4567-e89b-12d3-a456-426614174000",
  "description": "a thing used to illustrate a rule",
  "examples": [
    "this is an example of proper usage"
  ]
}
```

**Response:** `201 Created`
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174000",
  "entry_id": "123e4567-e89b-12d3-a456-426614174000",
  "part_of_speech": "noun",
  "description": "a thing used to illustrate a rule",
  "examples": [
    {
      "id": "423e4567-e89b-12d3-a456-426614174000",
      "text": "this is an example of proper usage",
      "context": "",
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T15:30:45Z"
    }
  ],
  "likes_count": 0,
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### Update Meaning

```
PUT /entries/{entryId}/meanings/{meaningId}
```

Updates an existing meaning.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Request Body:**
```json
{
  "part_of_speech_id": "723e4567-e89b-12d3-a456-426614174000",
  "description": "updated description",
  "examples": [
    "updated example text"
  ]
}
```

**Response:** `200 OK`
```json
{
  "id": "323e4567-e89b-12d3-a456-426614174000",
  "entry_id": "123e4567-e89b-12d3-a456-426614174000",
  "part_of_speech": "noun",
  "description": "updated description",
  "examples": [
    {
      "id": "423e4567-e89b-12d3-a456-426614174000",
      "text": "updated example text",
      "context": "",
      "created_at": "2023-04-10T15:30:45Z",
      "updated_at": "2023-04-10T16:45:12Z"
    }
  ],
  "likes_count": 0,
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T16:45:12Z"
}
```

#### Delete Meaning

```
DELETE /entries/{entryId}/meanings/{meaningId}
```

Deletes a meaning from an entry.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Response:** `204 No Content`

#### Add Translation

```
POST /entries/{entryId}/meanings/{meaningId}/translations
```

Adds a translation to a meaning.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Request Body:**
```json
{
  "language_id": "fr",
  "text": "exemple"
}
```

**Response:** `201 Created`
```json
{
  "id": "523e4567-e89b-12d3-a456-426614174000",
  "meaning_id": "323e4567-e89b-12d3-a456-426614174000",
  "language_id": "fr",
  "text": "exemple",
  "likes_count": 0,
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z",
  "created_by": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "avatar": "https://example.com/avatar.jpg"
  }
}
```

#### Update Translation

```
PUT /entries/{entryId}/meanings/{meaningId}/translations/{translationId}
```

Updates a translation.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning
- `translationId`: UUID of the translation

**Request Body:**
```json
{
  "text": "updated exemple"
}
```

**Response:** `200 OK`
```json
{
  "id": "523e4567-e89b-12d3-a456-426614174000",
  "meaning_id": "323e4567-e89b-12d3-a456-426614174000",
  "language_id": "fr",
  "text": "updated exemple",
  "likes_count": 0,
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T16:45:12Z",
  "created_by": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "avatar": "https://example.com/avatar.jpg"
  }
}
```

#### Delete Translation

```
DELETE /entries/{entryId}/meanings/{meaningId}/translations/{translationId}
```

Deletes a translation.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning
- `translationId`: UUID of the translation

**Response:** `204 No Content`

### Social Interaction Endpoints

#### Add Comment to Meaning

```
POST /entries/{entryId}/meanings/{meaningId}/comments
```

Adds a comment to a meaning.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Request Body:**
```json
{
  "content": "This meaning is very helpful!"
}
```

**Response:** `201 Created`
```json
{
  "id": "623e4567-e89b-12d3-a456-426614174000",
  "content": "This meaning is very helpful!",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "avatar": "https://example.com/avatar.jpg"
  },
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### Toggle Like on Meaning

```
POST /entries/{entryId}/meanings/{meaningId}/likes
```

Toggles a like on a meaning (adds if not present, removes if present).

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning

**Response:** `204 No Content`

#### Add Comment to Translation

```
POST /entries/{entryId}/meanings/{meaningId}/translations/{translationId}/comments
```

Adds a comment to a translation.

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning
- `translationId`: UUID of the translation

**Request Body:**
```json
{
  "content": "This translation is accurate."
}
```

**Response:** `201 Created`
```json
{
  "id": "723e4567-e89b-12d3-a456-426614174000",
  "content": "This translation is accurate.",
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "user123",
    "avatar": "https://example.com/avatar.jpg"
  },
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

#### Toggle Like on Translation

```
POST /entries/{entryId}/meanings/{meaningId}/translations/{translationId}/likes
```

Toggles a like on a translation (adds if not present, removes if present).

**Authentication:** Required

**Path Parameters:**
- `entryId`: UUID of the entry
- `meaningId`: UUID of the meaning
- `translationId`: UUID of the translation

**Response:** `204 No Content`

## User Endpoints

### Get Current User

```
GET /users/me
```

Retrieves the currently authenticated user's profile.

**Authentication:** Required

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "user123",
  "email": "user@example.com",
  "avatar": "https://example.com/avatar.jpg",
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T15:30:45Z"
}
```

### Update Current User

```
PUT /users/me
```

Updates the currently authenticated user's profile.

**Authentication:** Required

**Request Body:**
```json
{
  "username": "newusername",
  "email": "new@example.com",
  "password": "newpassword",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

**Response:** `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "newusername",
  "email": "new@example.com",
  "avatar": "https://example.com/new-avatar.jpg",
  "created_at": "2023-04-10T15:30:45Z",
  "updated_at": "2023-04-10T16:45:12Z"
}
```

### Delete Current User

```
DELETE /users/me
```

Deletes the currently authenticated user's account.

**Authentication:** Required

**Response:** `204 No Content`

### List User Entries

```
GET /users/me/entries
```

Retrieves entries created by the current user.

**Authentication:** Required

**Query Parameters:**
- `limit`: Maximum number of entries to return (default: 20, max: 100)
- `offset`: Number of entries to skip (for pagination)
- `sort_by`: Field to sort by (`word`, `created_at`, `updated_at`)
- `sort_desc`: If true, sort in descending order (default: false)
- `word_filter`: Filter entries by word (partial match)
- `type`: Filter entries by type (`WORD`, `COMPOUND_WORD`, `PHRASE`)

**Response:** `200 OK`
```json
{
  "entries": [...],
  "total": 10,
  "limit": 20,
  "offset": 0
}
```

### List User Translations

```
GET /users/me/translations
```

Retrieves translations created by the current user.

**Authentication:** Required

**Query Parameters:**
- `limit`: Maximum number of translations to return (default: 20, max: 100)
- `offset`: Number of translations to skip (for pagination)
- `sort_by`: Field to sort by (`created_at`, `updated_at`, `language_id`)
- `sort_desc`: If true, sort in descending order (default: false)
- `language_id`: Filter translations by language (ISO 639-1 code)
- `text_search`: Filter translations by text (partial match)

**Response:** `200 OK`
```json
{
  "translations": [...],
  "total": 15,
  "limit": 20,
  "offset": 0
}
```

### List User Comments

```
GET /users/me/comments
```

Retrieves comments created by the current user.

**Authentication:** Required

**Query Parameters:**
- `limit`: Maximum number of comments to return (default: 20, max: 100)
- `offset`: Number of comments to skip (for pagination)
- `sort_by`: Field to sort by (`created_at`, `target_type`)
- `sort_desc`: If true, sort in descending order (default: false)
- `target_type`: Filter comments by target type (`meaning`, `translation`)
- `from_date`: Filter comments created after this date
- `to_date`: Filter comments created before this date

**Response:** `200 OK`
```json
{
  "comments": [...],
  "total": 8,
  "limit": 20,
  "offset": 0
}
```

### List User Likes

```
GET /users/me/likes
```

Retrieves likes created by the current user.

**Authentication:** Required

**Query Parameters:**
- `limit`: Maximum number of likes to return (default: 20, max: 100)
- `offset`: Number of likes to skip (for pagination)
- `sort_by`: Field to sort by (`created_at`, `target_type`)
- `sort_desc`: If true, sort in descending order (default: false)
- `target_type`: Filter likes by target type (`meaning`, `translation`)
- `from_date`: Filter likes created after this date
- `to_date`: Filter likes created before this date

**Response:** `200 OK`
```json
{
  "likes": [...],
  "total": 12,
  "limit": 20,
  "offset": 0
}
```

## Admin Endpoints

### Get System Status

```
GET /admin/status
```

Retrieves system status information.

**Authentication:** Required (Admin role)

**Response:** `200 OK`
```json
{
  "status": "ok",
  "message": "Admin stats endpoint"
}
```

## Error Responses

The API returns standard HTTP status codes along with error messages in JSON format:

### Bad Request

```
400 Bad Request
```

```json
{
  "error": "Invalid request format"
}
```

### Unauthorized

```
401 Unauthorized
```

```json
{
  "error": "Unauthorized"
}
```

### Forbidden

```
403 Forbidden
```

```json
{
  "error": "Forbidden"
}
```

### Not Found

```
404 Not Found
```

```json
{
  "error": "Resource not found"
}
```

### Conflict

```
409 Conflict
```

```json
{
  "error": "Resource already exists"
}
```

### Rate Limit Exceeded

```
429 Too Many Requests
```

```json
{
  "error": "Rate limit exceeded. Please try again later."
}
```

### Internal Server Error

```
500 Internal Server Error
```

```json
{
  "error": "An internal server error occurred"
}
```

## API Versioning

The API uses URL versioning (e.g., `/api/v1`). Future versions may use a different URL prefix (e.g., `/api/v2`).

## Rate Limiting

The API implements rate limiting to prevent abuse. Clients are limited to 10 requests per second with a burst capacity of 20 requests. When rate limits are exceeded, the API returns a 429 Too Many Requests response.

## Health Check

A simple health check endpoint is available at `/health` to verify the API server is running:

```
GET /health
```

**Response:** `200 OK`
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```
