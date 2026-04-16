import os

broker_url = os.environ.get("CELERY_BROKER_URL", "redis://redis-service:6379/0")
result_backend = os.environ.get("CELERY_RESULT_BACKEND", "redis://redis-service:6379/0")
task_serializer = "json"
result_serializer = "json"
accept_content = ["json"]
timezone = "UTC"
