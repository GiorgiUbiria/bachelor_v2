from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import logging

from api import recommendations, search, trends

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Bachelor ML Service",
    description="Machine Learning service for e-commerce platform",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:5173", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(recommendations.router, tags=["recommendations"])
app.include_router(search.router, tags=["search"])
app.include_router(trends.router, tags=["trends"])

@app.get("/")
async def root():
    return {
        "service": "Bachelor ML Service",
        "version": "1.0.0",
        "status": "running",
        "endpoints": {
            "recommendations": "/recommendations",
            "search": "/search",
            "trends": "/trends"
        }
    }

@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "ml_service"
    }

@app.on_event("startup")
async def startup_event():
    logger.info("ML Service starting up...")
    
    # Initialize ML models on startup
    try:
        from models.recommendations import recommendation_engine
        from models.search import search_engine
        from models.trends import trend_analyzer
        
        # Train recommendation engine
        logger.info("Initializing recommendation engine...")
        await recommendation_engine.train_models()
        
        # Initialize search engine
        logger.info("Initializing search engine...")
        await search_engine.initialize()
        
        # No initialization needed for trend analyzer as it works on-demand
        logger.info("Trend analyzer ready...")
        
        logger.info("All ML models initialized successfully")
        
    except Exception as e:
        logger.error(f"Failed to initialize ML models: {e}")
        # Don't fail startup, models can be initialized on first request

@app.on_event("shutdown")
async def shutdown_event():
    logger.info("ML Service shutting down...")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000) 