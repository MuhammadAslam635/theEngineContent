# API Reference — Backend Go

**Base URL:** `http://localhost:9001` (local) or via Kong at `http://localhost:8000/api`

---

## Health & Status

### GET `/health`
Simple liveness check.

**Response:** `200 OK` — plain text `OK`

---

### GET `/ready`
Readiness check (verifies DB connection).

**Response:**
- `200 OK` — plain text `READY`
- `503 Service Unavailable` — plain text `NOT READY`

---

### GET `/ai-health`
Checks AI Orchestration service health.

**Response (200):**
```json
{ "status": "AI Orchestration is healthy" }
```

**Response (503):**
```json
{ "status": "AI Orchestration unreachable", "error": "..." }
```

---

## Authentication

### POST `/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",       // required, valid email
  "password": "mypassword123"        // required
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "username": "john_doe",
    "email": "user@example.com",
    "utype": "user"
  }
}
```

**Response (401):**
```json
{ "error": "invalid email or password" }
```

---

### POST `/auth/forget-password`

**Request Body:**
```json
{
  "email": "user@example.com"        // required, valid email
}
```

**Response (200):**
```json
{ "message": "Verification code sent to your email" }
```

**Response (500):**
```json
{ "error": "user not found" }
```

---

### POST `/auth/reset-password`

**Request Body:**
```json
{
  "otp": "123456",                   // required
  "email": "user@example.com",      // required, valid email
  "new_password": "newpass123",      // required, min 8 chars
  "confirm_password": "newpass123"   // required, must match new_password
}
```

**Response (200):**
```json
{ "message": "Password reset successfully" }
```

**Response (400):**
```json
{ "error": "invalid or expired OTP" }
```

---

## Agent Settings

### GET `/agents/all`

Returns all agent settings ordered by creation date (newest first).

**Response (200):**
```json
{
  "data": [
    {
      "id": 1,
      "agent_name": "CompanyResearchAgent",
      "supervisor": "LeadEnrichmentSupervisor",
      "provider": "openai",
      "model_name": "gpt-4o",
      "prompt": "Research {{company_name}} in {{industry}}...",
      "variables": [
        { "key": "company_name", "description": "Target company", "required": true },
        { "key": "industry", "description": "Industry vertical", "required": false }
      ],
      "output_type": "json",
      "output_schema": {
        "type": "object",
        "properties": {
          "summary": { "type": "string" },
          "competitors": { "type": "array", "items": { "type": "string" } }
        },
        "required": ["summary"]
      },
      "temperature": 0.5,
      "max_tokens": 4096,
      "is_active": true,
      "created_at": "2026-04-20T10:00:00Z",
      "updated_at": "2026-04-20T10:00:00Z"
    }
  ]
}
```

---

### GET `/agents/:id`

Returns a single agent setting by ID.

**Path Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| id | int | Agent setting ID |

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "agent_name": "CompanyResearchAgent",
    "supervisor": "LeadEnrichmentSupervisor",
    "provider": "openai",
    "model_name": "gpt-4o",
    "prompt": "Research {{company_name}} in {{industry}}...",
    "variables": [...],
    "output_type": "json",
    "output_schema": {...},
    "temperature": 0.5,
    "max_tokens": 4096,
    "is_active": true,
    "created_at": "2026-04-20T10:00:00Z",
    "updated_at": "2026-04-20T10:00:00Z"
  }
}
```

**Response (404):**
```json
{ "error": "agent setting not found" }
```

---

### GET `/agents/by-name/:name`

Returns a single agent setting by its agent_name.

**Path Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| name | string | Agent name (e.g. CompanyResearchAgent) |

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "agent_name": "CompanyResearchAgent",
    ...full agent object...
  }
}
```

**Response (404):**
```json
{ "error": "agent setting not found" }
```

---

### POST `/agents/create`

Creates a new agent setting.

**Request Body:**
```json
{
  "agent_name": "CompanyResearchAgent",       // required
  "supervisor": "LeadEnrichmentSupervisor",   // optional
  "provider": "openai",                       // required
  "model_name": "gpt-4o",                    // required
  "prompt": "Research {{company_name}}...",    // optional
  "variables": [                              // optional, defaults to []
    {
      "key": "company_name",
      "description": "Target company name",
      "required": true
    }
  ],
  "output_type": "json",                      // required: "text" or "json"
  "output_schema": {                          // optional, used when output_type="json"
    "type": "object",
    "properties": {
      "summary": { "type": "string" }
    },
    "required": ["summary"]
  },
  "temperature": 0.5,                         // optional, defaults to 0.7
  "max_tokens": 4096,                         // optional, defaults to 2048
  "is_active": true                           // optional, defaults to true
}
```

**Field Validation:**
| Field | Rules |
|-------|-------|
| agent_name | required |
| provider | required |
| model_name | required |
| output_type | required, must be `text` or `json` |
| temperature | 0.0 – 2.0 recommended |
| max_tokens | positive integer |

**Response (201):**
```json
{
  "data": {
    "id": 2,
    "agent_name": "CompanyResearchAgent",
    ...full agent object...
  }
}
```

**Response (400):**
```json
{ "error": "Key: 'CreateAgentSettingRequest.AgentName' Error:Field validation for 'AgentName' failed on the 'required' tag" }
```

---

### PUT `/agents/update/:id`

Partially updates an existing agent setting. Only send fields you want to change.

**Path Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| id | int | Agent setting ID |

**Request Body (all fields optional):**
```json
{
  "agent_name": "UpdatedAgentName",
  "prompt": "New prompt with {{new_var}}...",
  "variables": [
    { "key": "new_var", "description": "A new variable", "required": true }
  ],
  "output_type": "text",
  "temperature": 0.9,
  "is_active": false
}
```

**Response (200):**
```json
{
  "data": {
    "id": 1,
    ...updated agent object...
  }
}
```

**Response (404):**
```json
{ "error": "agent setting not found" }
```

---

### DELETE `/agents/delete/:id`

Deletes an agent setting permanently.

**Path Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| id | int | Agent setting ID |

**Response (200):**
```json
{ "message": "agent setting deleted" }
```

**Response (404):**
```json
{ "error": "agent setting not found" }
```

---

## Common Error Response Format

All error responses follow this shape:
```json
{ "error": "description of what went wrong" }
```

## Notes

- All timestamps are in ISO 8601 / RFC 3339 format (UTC)
- `variables` uses `{{key}}` placeholder syntax in prompts
- `output_schema` follows JSON Schema draft-07 format
- Agent settings with `is_active: false` are stored but not used by the orchestration layer
