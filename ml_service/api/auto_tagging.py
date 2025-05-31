from fastapi import APIRouter, HTTPException, Query
from typing import Optional
import logging

from models.auto_tagging import auto_tagger

logger = logging.getLogger(__name__)
router = APIRouter()

@router.get("/")
async def auto_tagging_status():
    """Get auto-tagging service status"""
    return {
        "service": "auto_tagging",
        "status": "active",
        "is_trained": auto_tagger.is_trained
    }

@router.post("/initialize")
async def initialize_auto_tagger():
    """Initialize or retrain the auto-tagger"""
    try:
        await auto_tagger.initialize()
        return {
            "message": "Auto-tagger initialized successfully",
            "products_loaded": len(auto_tagger.product_data) if hasattr(auto_tagger, 'product_data') else 0
        }
    except Exception as e:
        logger.error(f"Failed to initialize auto-tagger: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/suggest/{product_id}")
async def suggest_tags_for_product(product_id: str):
    """Suggest tags for a specific product"""
    try:
        result = await auto_tagger.suggest_tags_for_product(product_id)
        return result
    except Exception as e:
        logger.error(f"Failed to suggest tags for product: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/auto-tag")
async def auto_tag_products(limit: int = Query(50, ge=1, le=200)):
    """Automatically suggest tags for products that need them"""
    try:
        result = await auto_tagger.auto_tag_products(limit)
        return result
    except Exception as e:
        logger.error(f"Failed to auto-tag products: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/insights")
async def get_tagging_insights():
    """Get insights about current tagging status"""
    try:
        result = await auto_tagger.get_tagging_insights()
        return result
    except Exception as e:
        logger.error(f"Failed to get tagging insights: {e}")
        raise HTTPException(status_code=500, detail=str(e)) 