# Social Profile API — Implementation Log

## Date: 2026-04-20

## Summary

Added SociaVault integration to fetch and persist social media channel profiles. Starting with YouTube; Instagram and Facebook will follow the same pattern.

## Migration

**File:** `migrations/20260420_03_create_youtube_channels_table.sql`

**Table:** `youtube_channels`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Auto-increment |
| user_id | INTEGER FK | Owner user |
| competitor_account_id | INTEGER FK | Optional link to competitor_accounts |
| channel_id | VARCHAR(255) UNIQUE | YouTube channel ID |
| channel_url | TEXT | Full channel URL |
| handle | VARCHAR(255) | @handle |
| name | VARCHAR(255) | Display name |
| avatar_url | TEXT | Profile image URL |
| description | TEXT | Channel bio |
| country | VARCHAR(100) | Registered country |
| joined_date | VARCHAR(100) | Join date text |
| subscriber_count | BIGINT | Subscriber count |
| subscriber_count_text | VARCHAR(50) | Human-readable (e.g. "2.97M") |
| video_count | INTEGER | Total videos |
| view_count | BIGINT | Total views |
| view_count_text | VARCHAR(100) | Human-readable views |
| tags | TEXT | Comma-separated tags |
| email | VARCHAR(255) | Public contact email |
| links | JSONB | External links |
| last_fetched_at | TIMESTAMPTZ | Last SociaVault fetch time |
| is_active | BOOLEAN | Active flag |
| created_at / updated_at | TIMESTAMPTZ | Timestamps |

## Go Backend Files

| File | Purpose |
|------|---------|
| `backend-go/config/config.go` | Added SociaVaultAPIKey, SociaVaultBaseURL |
| `backend-go/internal/models/youtube_channel.go` | GORM model |
| `backend-go/internal/dto/social_profile_dto.go` | Request DTO + SociaVault response mapping |
| `backend-go/internal/repositories/youtube_channel_repository.go` | Data access |
| `backend-go/internal/services/social_profile_service.go` | SociaVault API call + upsert logic |
| `backend-go/internal/handlers/social_profile_handler.go` | HTTP handler |

## API Endpoint

### POST `/social/fetch-profile`

Fetches a channel profile from SociaVault and saves/updates it in the database.

**Request:**
```json
{
  "platform": "youtube",
  "channel_name": "@ThePatMcAfeeShow"
}
```

`platform` must be one of: `youtube`, `instagram`, `facebook`
`channel_name` can be a @handle or channel URL/ID.

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "channel_id": "UCxcTeAKWJca6XyJ37_ZoKIQ",
    "channel_url": "http://www.youtube.com/@ThePatMcAfeeShow",
    "handle": "@ThePatMcAfeeShow",
    "name": "The Pat McAfee Show",
    "avatar_url": "https://yt3.googleusercontent.com/...",
    "description": "",
    "country": "United States",
    "joined_date": "Joined Aug 23, 2017",
    "subscriber_count": 2970000,
    "subscriber_count_text": "2.97M subscribers",
    "video_count": 10538,
    "view_count": 2462865520,
    "view_count_text": "2,462,865,520 views",
    "tags": "pat mcafee, football, ...",
    "email": null,
    "links": {"0": "https://store.patmcafeeshow.com", ...},
    "last_fetched_at": "2026-04-20T18:55:00Z",
    "is_active": true,
    "created_at": "2026-04-20T18:55:00Z",
    "updated_at": "2026-04-20T18:55:00Z"
  }
}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| SOCIAVAULT_API | API key for SociaVault |
| SOCIAVAULT_BASE_URL | Base URL (default: https://api.sociavault.com/v1/scrape/) |

## Design Decisions

- Upsert logic: if channel_id already exists, update stats instead of creating duplicate
- Separate tables per platform (youtube_channels, instagram_channels, facebook_pages) for schema flexibility
- Links stored as JSONB for variable structure across platforms
- `competitor_account_id` FK allows linking to existing competitor monitoring system
