from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel
from typing import List, Optional
import logging

from models.trends import trend_analyzer

logger = logging.getLogger(__name__)

router = APIRouter()

class TrendRequest(BaseModel):
    period: Optional[str] = "30d"  # 7d, 30d, 90d, 1y
    category: Optional[str] = None

@router.get("/sales")
async def get_sales_trends(
    period: str = Query("30d", description="Time period: 7d, 30d, 90d, 1y"),
    category: Optional[str] = Query(None, description="Product category filter")
):
    """Get sales trends analysis"""
    return {
        "message": "Sales trends endpoint - coming soon",
        "period": period,
        "category": category
    }

@router.get("/forecast")
async def get_sales_forecast(
    days: int = Query(30, description="Number of days to forecast"),
    category: Optional[str] = Query(None, description="Product category filter")
):
    """Get sales forecast"""
    return {
        "message": "Sales forecast endpoint - coming soon",
        "days": days,
        "category": category
    }

@router.get("/popular")
async def get_trending_products(
    period: str = Query("7d", description="Time period for trending calculation"),
    limit: int = Query(10, description="Number of trending products")
):
    """Get trending products"""
    return {
        "message": "Trending products endpoint - coming soon",
        "period": period,
        "limit": limit
    }

@router.get("/trends/products")
async def get_product_trends(
    limit: int = Query(20, description="Number of trending products to return", le=100)
):
    """
    Get trending products based on sales and interaction data
    """
    try:
        trends = await trend_analyzer.get_product_trends(limit)
        
        return {
            "success": True,
            "data": trends
        }
        
    except Exception as e:
        logger.error(f"Failed to get product trends: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get product trends: {str(e)}")

@router.get("/trends/categories")
async def get_category_trends():
    """
    Get category-level trends and performance
    """
    try:
        trends = await trend_analyzer.get_category_trends()
        
        return {
            "success": True,
            "data": trends
        }
        
    except Exception as e:
        logger.error(f"Failed to get category trends: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get category trends: {str(e)}")

@router.get("/trends/search")
async def get_search_trends():
    """
    Get search query trends and patterns
    """
    try:
        trends = await trend_analyzer.get_search_trends()
        
        return {
            "success": True,
            "data": trends
        }
        
    except Exception as e:
        logger.error(f"Failed to get search trends: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get search trends: {str(e)}")

@router.get("/trends/seasonal")
async def get_seasonal_patterns():
    """
    Get seasonal patterns in sales data
    """
    try:
        patterns = await trend_analyzer.get_seasonal_patterns()
        
        return {
            "success": True,
            "data": patterns
        }
        
    except Exception as e:
        logger.error(f"Failed to get seasonal patterns: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get seasonal patterns: {str(e)}")

@router.get("/forecast/demand/{product_id}")
async def forecast_product_demand(
    product_id: str,
    days_ahead: int = Query(30, description="Number of days to forecast ahead", le=365)
):
    """
    Forecast demand for a specific product
    """
    try:
        forecast = await trend_analyzer.forecast_demand(product_id, days_ahead)
        
        return {
            "success": True,
            "data": forecast
        }
        
    except Exception as e:
        logger.error(f"Failed to forecast demand for product {product_id}: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to forecast demand: {str(e)}")

@router.get("/trends/dashboard")
async def get_trends_dashboard():
    """
    Get comprehensive trends dashboard data
    """
    try:
        # Get all trend data
        product_trends = await trend_analyzer.get_product_trends(10)
        category_trends = await trend_analyzer.get_category_trends()
        search_trends = await trend_analyzer.get_search_trends()
        seasonal_patterns = await trend_analyzer.get_seasonal_patterns()
        
        dashboard_data = {
            "product_trends": product_trends,
            "category_trends": category_trends,
            "search_trends": search_trends,
            "seasonal_patterns": seasonal_patterns,
            "summary": {
                "total_products_analyzed": product_trends.get("total_products_analyzed", 0),
                "total_categories": len(category_trends),
                "is_trained": trend_analyzer.is_trained
            }
        }
        
        return {
            "success": True,
            "data": dashboard_data
        }
        
    except Exception as e:
        logger.error(f"Failed to get trends dashboard: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get trends dashboard: {str(e)}")

@router.post("/trends/retrain")
async def retrain_trend_analyzer():
    """
    Retrain the trend analyzer with latest data
    """
    try:
        await trend_analyzer.initialize()
        
        return {
            "success": True,
            "message": "Trend analyzer retrained successfully"
        }
        
    except Exception as e:
        logger.error(f"Failed to retrain trend analyzer: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to retrain: {str(e)}")

@router.get("/trends/status")
async def get_trends_status():
    """
    Get trend analyzer status and metrics
    """
    try:
        status = {
            "is_trained": trend_analyzer.is_trained,
            "has_sales_data": hasattr(trend_analyzer, 'sales_data') and not trend_analyzer.sales_data.empty,
            "has_interaction_data": hasattr(trend_analyzer, 'interaction_data') and not trend_analyzer.interaction_data.empty,
            "has_search_data": hasattr(trend_analyzer, 'search_data') and not trend_analyzer.search_data.empty,
            "trend_data_keys": list(trend_analyzer.trend_data.keys()) if trend_analyzer.trend_data else []
        }
        
        if hasattr(trend_analyzer, 'sales_data') and not trend_analyzer.sales_data.empty:
            status["sales_data_count"] = len(trend_analyzer.sales_data)
            status["date_range"] = {
                "start": trend_analyzer.sales_data['date'].min().isoformat() if 'date' in trend_analyzer.sales_data.columns else None,
                "end": trend_analyzer.sales_data['date'].max().isoformat() if 'date' in trend_analyzer.sales_data.columns else None
            }
        
        return {
            "success": True,
            "data": status
        }
        
    except Exception as e:
        logger.error(f"Failed to get trends status: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get status: {str(e)}")

@router.get("/trends/insights")
async def get_trend_insights():
    """
    Get actionable insights from trend analysis
    """
    try:
        insights = []
        
        # Get product trends for insights
        product_trends = await trend_analyzer.get_product_trends(50)
        trending_products = product_trends.get('trending_products', [])
        
        # Identify top growing products
        growing_products = [
            p for p in trending_products 
            if p.get('trend_metrics', {}).get('units_sold_trend') == 'growing'
        ][:5]
        
        if growing_products:
            insights.append({
                "type": "opportunity",
                "title": "Fast Growing Products",
                "description": f"Found {len(growing_products)} products with strong growth trends",
                "products": [p['product_name'] for p in growing_products],
                "action": "Consider increasing inventory and marketing focus"
            })
        
        # Identify declining products
        declining_products = [
            p for p in trending_products 
            if p.get('trend_metrics', {}).get('units_sold_trend') == 'declining'
        ][:5]
        
        if declining_products:
            insights.append({
                "type": "warning",
                "title": "Declining Products",
                "description": f"Found {len(declining_products)} products with declining sales",
                "products": [p['product_name'] for p in declining_products],
                "action": "Review pricing, promotions, or consider discontinuation"
            })
        
        # Category insights
        category_trends = await trend_analyzer.get_category_trends()
        if category_trends:
            top_category = max(
                category_trends.items(),
                key=lambda x: x[1].get('total_revenue', 0)
            )
            
            insights.append({
                "type": "info",
                "title": "Top Performing Category",
                "description": f"'{top_category[0]}' is the highest revenue category",
                "revenue": top_category[1].get('total_revenue', 0),
                "action": "Focus marketing efforts on this category"
            })
        
        # Search insights
        search_trends = await trend_analyzer.get_search_trends()
        if search_trends and search_trends.get('top_queries'):
            top_query = search_trends['top_queries'][0]
            insights.append({
                "type": "info",
                "title": "Most Popular Search",
                "description": f"'{top_query['query']}' is the most searched term",
                "search_count": top_query['search_count'],
                "action": "Ensure adequate inventory for related products"
            })
        
        return {
            "success": True,
            "data": {
                "insights": insights,
                "generated_at": trend_analyzer.trend_data.get('last_updated', 'unknown')
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get trend insights: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get insights: {str(e)}") 