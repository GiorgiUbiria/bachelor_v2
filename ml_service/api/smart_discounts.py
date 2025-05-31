from fastapi import APIRouter, HTTPException, Query
from typing import Optional
import logging

from models.smart_discounts import smart_discount_engine

logger = logging.getLogger(__name__)
router = APIRouter()

@router.get("/")
async def smart_discounts_status():
    """Get smart discounts service status"""
    return {
        "service": "smart_discounts",
        "status": "active",
        "is_trained": smart_discount_engine.is_trained
    }

@router.post("/initialize")
async def initialize_smart_discounts():
    """Initialize or retrain the smart discount engine"""
    try:
        await smart_discount_engine.initialize()
        return {
            "message": "Smart discount engine initialized successfully",
            "products_analyzed": len(smart_discount_engine.product_performance)
        }
    except Exception as e:
        logger.error(f"Failed to initialize smart discount engine: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/suggest/product/{product_id}")
async def suggest_product_discount(product_id: str):
    """Suggest optimal discount for a specific product"""
    try:
        result = await smart_discount_engine.suggest_product_discount(product_id)
        return result
    except Exception as e:
        logger.error(f"Failed to suggest product discount: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/suggest/category/{category}")
async def suggest_category_discounts(category: str):
    """Suggest discounts for an entire category"""
    try:
        result = await smart_discount_engine.suggest_category_discounts(category)
        return result
    except Exception as e:
        logger.error(f"Failed to suggest category discounts: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/insights")
async def get_discount_insights():
    """Get overall discount insights and recommendations"""
    try:
        result = await smart_discount_engine.get_discount_insights()
        return result
    except Exception as e:
        logger.error(f"Failed to get discount insights: {e}")
        raise HTTPException(status_code=500, detail=str(e)) 