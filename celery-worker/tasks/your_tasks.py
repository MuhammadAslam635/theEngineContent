from . import app
import time

@app.task(bind=True, max_retries=3)
def process_job(self, job_id: str, payload: dict):
    try:
        print(f"Processing job {job_id} with payload {payload}")
        time.sleep(2) # Simulate work
        return {"status": "completed", "job_id": job_id}
    except Exception as exc:
        raise self.retry(exc=exc, countdown=60)
