# Product Overview

cont-gen (Content Engine) is a multi-tenant AI-powered content generation and CRM platform. It automates the full content lifecycle — from competitor intelligence and trend detection, through script writing and video production, to posting and analytics.

The system is branded "Scrpy / Elara" on the AI side. Elara is the main AI supervisor that routes user intent across specialized agents (lead enrichment, system insights, user management, communications, campaigns, media generation, ads content).

## Core Domains

- **Lead Enrichment**: Find, enrich, and manage company leads from external sources (Apollo/web search) with quota/balance checks.
- **Content Intelligence**: Monitor competitor accounts for outlier reels, detect trend signals, maintain hook and angle libraries.
- **Content Pipeline**: AI agents produce content briefs → script generation (with 5-dimension scoring) → SSML conversion → video production (HeyGen/Kling AI) → posting packages.
- **Analytics**: Track video performance at 24h/7d/30d windows against baselines.
- **User & RBAC**: Multi-tenant user management with sub-users, roles, and permissions.
- **Communications**: Email, SMS, WhatsApp, AI voice calls, proposals, and appointment booking.
- **Campaigns**: Google/Meta Ads campaign creation and management.

## Multi-Tenant Architecture

All data is scoped by `tenant_id` and `user_id`. The AI orchestration layer receives these via `x-user-id` header from the Go backend (which extracts them from JWT claims). External traffic never reaches the AI service directly.
