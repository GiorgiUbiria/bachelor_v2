from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os
from dotenv import load_dotenv

from api.recommendations import router as recommendations_router
from api.search import router as search_router
from api.trends import router as trends_router
from api.chatbot import router as chatbot_router
from database.connection import init_db

# Load environment variables
load_dotenv()

# Create FastAPI app
app = FastAPI(
    title="Bachelor ML Service",
    description="Machine Learning service for e-commerce platform",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize database connection
@app.on_event("startup")
async def startup_event():
    await init_db()

# Health check endpoint
@app.get("/health")
async def health_check():
    return {
        "status": "ok",
        "service": "bachelor_ml_service",
        "version": "1.0.0"
    }

# Include routers
app.include_router(recommendations_router, prefix="/api/v1/recommendations", tags=["recommendations"])
app.include_router(search_router, prefix="/api/v1/search", tags=["search"])
app.include_router(trends_router, prefix="/api/v1/trends", tags=["trends"])
app.include_router(chatbot_router, prefix="/api/v1/chatbot", tags=["chatbot"])

# Root endpoint
@app.get("/")
async def root():
    return {
        "message": "Bachelor ML Service API",
        "docs": "/docs",
        "health": "/health"
    }

if __name__ == "__main__":
    port = int(os.getenv("ML_SERVICE_PORT", 8000))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=True if os.getenv("ENVIRONMENT") == "development" else False
    ) 