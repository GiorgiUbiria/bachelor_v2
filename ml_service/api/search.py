from fastapi import APIRouter, Query, HTTPException, Depends
from typing import Optional, List
import logging

from models.search import search_engine

logger = logging.getLogger(__name__)

router = APIRouter()

@router.get("/search")
async def enhanced_search(
    q: str = Query(..., description="Search query"),
    user_id: Optional[str] = Query(None, description="User ID for personalization"),
    category: Optional[str] = Query(None, description="Filter by category"),
    min_price: Optional[float] = Query(None, description="Minimum price filter"),
    max_price: Optional[float] = Query(None, description="Maximum price filter"),
    sort_by: str = Query("relevance", description="Sort by: relevance, price_low, price_high, newest, name"),
    limit: int = Query(20, description="Number of results to return", le=100)
):
    """
    Enhanced search endpoint with ML-powered features:
    - Semantic search using TF-IDF
    - Query expansion and preprocessing
    - Personalized results based on user behavior
    - Advanced filtering and sorting
    """
    try:
        results = await search_engine.search(
            query=q,
            user_id=user_id,
            category=category,
            min_price=min_price,
            max_price=max_price,
            sort_by=sort_by,
            limit=limit
        )
        
        return {
            "success": True,
            "data": results
        }
        
    except Exception as e:
        logger.error(f"Search failed: {e}")
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")

@router.get("/search/suggestions")
async def get_search_suggestions(
    q: str = Query(..., description="Partial search query"),
    limit: int = Query(5, description="Number of suggestions to return", le=10)
):
    """
    Get search suggestions based on query
    """
    try:
        suggestions = await search_engine._generate_suggestions(q)
        
        return {
            "success": True,
            "data": {
                "query": q,
                "suggestions": suggestions[:limit]
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get suggestions: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get suggestions: {str(e)}")

@router.get("/search/analytics")
async def get_search_analytics(
    days: int = Query(30, description="Number of days to analyze", le=365)
):
    """
    Get search analytics and insights
    """
    try:
        analytics = await search_engine.get_search_analytics(days)
        
        return {
            "success": True,
            "data": analytics
        }
        
    except Exception as e:
        logger.error(f"Failed to get search analytics: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get analytics: {str(e)}")

@router.post("/search/reindex")
async def reindex_search():
    """
    Reindex search engine with latest product data
    """
    try:
        await search_engine.initialize()
        
        return {
            "success": True,
            "message": "Search index rebuilt successfully"
        }
        
    except Exception as e:
        logger.error(f"Failed to reindex: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to reindex: {str(e)}")

@router.get("/search/status")
async def get_search_status():
    """
    Get search engine status
    """
    try:
        return {
            "success": True,
            "data": {
                "is_trained": search_engine.is_trained,
                "product_count": len(search_engine.product_data) if search_engine.product_data is not None else 0,
                "search_history_count": len(search_engine.search_history),
                "has_tfidf_vectorizer": search_engine.tfidf_vectorizer is not None,
                "has_product_features": search_engine.product_features is not None
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get search status: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get status: {str(e)}")

@router.get("/search/categories")
async def get_search_categories():
    """
    Get available product categories for search filtering
    """
    try:
        if search_engine.product_data is None:
            await search_engine.initialize()
            
        categories = search_engine.product_data['category'].unique().tolist()
        
        return {
            "success": True,
            "data": {
                "categories": sorted(categories)
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get categories: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get categories: {str(e)}")

@router.get("/search/popular")
async def get_popular_searches(
    limit: int = Query(10, description="Number of popular searches to return", le=50)
):
    """
    Get popular search queries
    """
    try:
        analytics = await search_engine.get_search_analytics(30)
        popular = analytics.get('top_queries', [])[:limit]
        
        return {
            "success": True,
            "data": {
                "popular_searches": popular
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to get popular searches: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get popular searches: {str(e)}") 