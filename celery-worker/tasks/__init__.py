from celery import Celery
import os

app = Celery("tasks")
app.config_from_object("celeryconfig")

# Import tasks here
from .your_tasks import process_job
