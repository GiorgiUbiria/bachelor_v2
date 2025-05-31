from fastapi import APIRouter, HTTPException, Query
from typing import Optional
import logging

from models.sentiment import sentiment_analyzer

logger = logging.getLogger(__name__)
router = APIRouter()

@router.get("/")
async def sentiment_status():
    """Get sentiment analysis service status"""
    return {
        "service": "sentiment_analysis",
        "status": "active",
        "is_trained": sentiment_analyzer.is_trained,
        "last_updated": getattr(sentiment_analyzer, 'sentiment_data', {}).get('last_updated') if sentiment_analyzer.is_trained else None
    }

@router.post("/initialize")
async def initialize_sentiment_analyzer():
    """Initialize or retrain the sentiment analyzer"""
    try:
        await sentiment_analyzer.initialize()
        return {
            "message": "Sentiment analyzer initialized successfully",
            "total_analyzed": sentiment_analyzer.sentiment_data.get('total_analyzed', 0)
        }
    except Exception as e:
        logger.error(f"Failed to initialize sentiment analyzer: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/product/{product_id}")
async def analyze_product_sentiment(product_id: str):
    """Analyze sentiment for a specific product"""
    try:
        result = await sentiment_analyzer.analyze_product_sentiment(product_id)
        return result
    except Exception as e:
        logger.error(f"Failed to analyze product sentiment: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/category/{category}")
async def analyze_category_sentiment(category: str):
    """Analyze sentiment for a product category"""
    try:
        result = await sentiment_analyzer.get_category_sentiment(category)
        return result
    except Exception as e:
        logger.error(f"Failed to analyze category sentiment: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/trends")
async def get_sentiment_trends(days: int = Query(30, ge=1, le=365)):
    """Get sentiment trends over time"""
    try:
        result = await sentiment_analyzer.get_sentiment_trends(days)
        return result
    except Exception as e:
        logger.error(f"Failed to get sentiment trends: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/insights")
async def get_sentiment_insights():
    """Get actionable sentiment insights"""
    try:
        result = await sentiment_analyzer.get_sentiment_insights()
        return result
    except Exception as e:
        logger.error(f"Failed to get sentiment insights: {e}")
        raise HTTPException(status_code=500, detail=str(e)) 