from fastapi import FastAPI, Request, HTTPException
import os
import uvicorn

app = FastAPI()

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.get("/ready")
async def ready():
    # Check DB connection here if needed
    return {"status": "ready"}

@app.post("/generate")
async def generate(request: Request):
    user_id = request.headers.get("x-user-id")
    if not user_id:
        raise HTTPException(status_code=401, detail="Missing x-user-id")
    
    # Logic for AI generation
    return {"message": f"Generation started for user {user_id}"}

if __name__ == "__main__":
    port = int(os.getenv("APP_PORT", 8086))
    uvicorn.run(app, host="0.0.0.0", port=port)
