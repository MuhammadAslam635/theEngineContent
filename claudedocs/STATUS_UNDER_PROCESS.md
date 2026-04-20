# Status — Under Process

## Active Work (Right Now)

- Align ai-orchestration + celery-worker with the “DB ownership” rule: no direct Postgres connections from these services.
- Introduce the shared `http_client.py` pattern (ai-orchestration + celery-worker) so all reads/writes go through backend-go.
- Implement backend-go settings endpoints so orchestration can fetch:
  - integration API keys (OpenAI/Gemini/SociaVault/HeyGen/ElevenLabs/etc.)
  - agent config (agent, supervisor, provider, model, prompt)

## Known Gaps Detected While Reading Architecture Doc

- Current ai-orchestration contains direct DB code for audit logging, which violates the architecture rule.
- “integration_settings” table name in the doc differs from the currently created “global_settings” table name.
