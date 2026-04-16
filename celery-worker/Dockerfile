FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Celery health check uses "celery inspect ping"
# which is why we don't need a separate API here.

CMD ["celery", "-A", "tasks", "worker", "--loglevel=info"]
