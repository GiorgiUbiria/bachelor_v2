from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
import logging
import uvicorn

from api import (
    recommendations, 
    sentiment, 
    auto_tagging, 
    smart_discounts, 
    search, 
    trends,
    anomaly_detection,
    explainability
)

# Import new ML model routers
from api.sentiment import router as sentiment_router
from api.auto_tagging import router as auto_tagging_router
from api.smart_discounts import router as smart_discounts_router
from api.anomaly_detection import router as anomaly_detection_router

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Bachelor ML Service",
    description="Machine Learning service for e-commerce platform with security monitoring",
    version="2.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://localhost:5173", "http://localhost:8080"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(recommendations.router, tags=["recommendations"])
app.include_router(search.router, tags=["search"])
app.include_router(trends.router, tags=["trends"])

# Include new ML model routers
app.include_router(sentiment_router, prefix="/sentiment", tags=["sentiment"])
app.include_router(auto_tagging_router, prefix="/auto-tagging", tags=["auto-tagging"])
app.include_router(smart_discounts_router, prefix="/smart-discounts", tags=["smart-discounts"])
app.include_router(anomaly_detection_router, prefix="/anomaly-detection", tags=["anomaly-detection"])
app.include_router(explainability.router)

@app.get("/")
async def root():
    return {
        "service": "Bachelor ML Service",
        "version": "2.0.0",
        "status": "running",
        "description": "Multi-Layered Framework for E-Commerce Security, Personalization, and Monitoring",
        "endpoints": {
            "recommendations": "/recommendations",
            "search": "/search",
            "trends": "/trends",
            "sentiment": "/sentiment",
            "auto_tagging": "/auto-tagging",
            "smart_discounts": "/smart-discounts",
            "anomaly_detection": "/anomaly-detection"
        },
        "ml_applications": [
            "Request Analysis for Anomalies",
            "Sentiment Analysis", 
            "Auto-Tagging",
            "Smart Discounts",
            "Product Recommendations",
            "Enhanced Search",
            "Trends Analysis"
        ]
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
    
    try:
        from models.recommendations import recommendation_engine
        from models.search import search_engine
        from models.trends import trend_analyzer
        from models.sentiment import sentiment_analyzer
        from models.auto_tagging import auto_tagger
        from models.smart_discounts import smart_discount_engine
        from models.anomaly_detection import anomaly_detector
        
        logger.info("Initializing recommendation engine...")
        await recommendation_engine.train_models()
        
        logger.info("Initializing search engine...")
        await search_engine.initialize()
        
        logger.info("Trend analyzer ready...")
        
        logger.info("Initializing sentiment analyzer...")
        try:
            await sentiment_analyzer.initialize()
            logger.info("Sentiment analyzer initialized successfully")
        except Exception as e:
            logger.warning(f"Sentiment analyzer initialization failed: {e}")
        
        logger.info("Initializing auto-tagger...")
        try:
            await auto_tagger.initialize()
            logger.info("Auto-tagger initialized successfully")
        except Exception as e:
            logger.warning(f"Auto-tagger initialization failed: {e}")
        
        logger.info("Initializing smart discount engine...")
        try:
            await smart_discount_engine.initialize()
            logger.info("Smart discount engine initialized successfully")
        except Exception as e:
            logger.warning(f"Smart discount engine initialization failed: {e}")
        
        logger.info("Initializing anomaly detector...")
        try:
            await anomaly_detector.initialize()
            logger.info("Anomaly detector initialized successfully")
        except Exception as e:
            logger.warning(f"Anomaly detector initialization failed: {e}")
        
        logger.info("All ML models initialized successfully")
        
    except Exception as e:
        logger.error(f"Failed to initialize core ML models: {e}")

@app.on_event("shutdown")
async def shutdown_event():
    logger.info("ML Service shutting down...")

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000) 