from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel
from typing import List, Optional
import uuid

from models.recommendations import recommendation_engine

router = APIRouter()

class RecommendationRequest(BaseModel):
    user_id: str
    algorithm: Optional[str] = "hybrid"  # collaborative, content_based, hybrid, popular
    limit: Optional[int] = 10

class RecommendationResponse(BaseModel):
    product_id: str
    score: float
    algorithm: str

class RecommendationsListResponse(BaseModel):
    recommendations: List[RecommendationResponse]
    user_id: str
    algorithm: str
    total: int

@router.post("/generate", response_model=RecommendationsListResponse)
async def generate_recommendations(request: RecommendationRequest):
    """Generate recommendations for a user"""
    try:
        # Validate user_id format
        try:
            uuid.UUID(request.user_id)
        except ValueError:
            raise HTTPException(status_code=400, detail="Invalid user_id format")
        
        # Ensure models are trained
        if not recommendation_engine.is_trained:
            await recommendation_engine.train_models()
        
        # Generate recommendations based on algorithm
        if request.algorithm == "collaborative":
            recommendations = await recommendation_engine.get_collaborative_recommendations(
                request.user_id, request.limit
            )
        elif request.algorithm == "content_based":
            recommendations = await recommendation_engine.get_content_based_recommendations(
                request.user_id, request.limit
            )
        elif request.algorithm == "hybrid":
            recommendations = await recommendation_engine.get_hybrid_recommendations(
                request.user_id, request.limit
            )
        elif request.algorithm == "popular":
            recommendations = await recommendation_engine._get_popular_items(request.limit)
        else:
            raise HTTPException(status_code=400, detail="Invalid algorithm type")
        
        # Save recommendations to database
        if recommendations:
            await recommendation_engine.save_recommendations_to_db(request.user_id, recommendations)
        
        # Convert to response format
        recommendation_responses = [
            RecommendationResponse(
                product_id=rec['product_id'],
                score=rec['score'],
                algorithm=rec['algorithm']
            )
            for rec in recommendations
        ]
        
        return RecommendationsListResponse(
            recommendations=recommendation_responses,
            user_id=request.user_id,
            algorithm=request.algorithm,
            total=len(recommendation_responses)
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error generating recommendations: {str(e)}")

@router.get("/user/{user_id}", response_model=RecommendationsListResponse)
async def get_user_recommendations(
    user_id: str,
    algorithm: str = Query("hybrid", description="Algorithm type: collaborative, content_based, hybrid, popular"),
    limit: int = Query(10, description="Number of recommendations to return")
):
    """Get recommendations for a specific user"""
    request = RecommendationRequest(
        user_id=user_id,
        algorithm=algorithm,
        limit=limit
    )
    return await generate_recommendations(request)

@router.post("/train")
async def train_recommendation_models():
    """Train/retrain the recommendation models"""
    try:
        await recommendation_engine.train_models()
        return {
            "message": "Recommendation models trained successfully",
            "status": "success"
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error training models: {str(e)}")

@router.get("/status")
async def get_recommendation_status():
    """Get the status of recommendation models"""
    return {
        "is_trained": recommendation_engine.is_trained,
        "has_user_item_matrix": recommendation_engine.user_item_matrix is not None,
        "has_content_similarity": recommendation_engine.content_similarity_matrix is not None,
        "has_svd_model": recommendation_engine.svd_model is not None
    }

@router.post("/feedback")
async def record_recommendation_feedback(
    user_id: str,
    product_id: str,
    feedback_type: str  # 'clicked', 'purchased', 'dismissed'
):
    """Record user feedback on recommendations"""
    try:
        # Validate UUIDs
        try:
            uuid.UUID(user_id)
            uuid.UUID(product_id)
        except ValueError:
            raise HTTPException(status_code=400, detail="Invalid user_id or product_id format")
        
        if feedback_type not in ['clicked', 'purchased', 'dismissed']:
            raise HTTPException(status_code=400, detail="Invalid feedback_type")
        
        # Save feedback to database
        from database.connection import execute_query
        await execute_query(
            """
            INSERT INTO recommendation_feedback (user_id, product_id, feedback_type)
            VALUES ($1, $2, $3)
            """,
            [uuid.UUID(user_id), uuid.UUID(product_id), feedback_type]
        )
        
        return {
            "message": "Feedback recorded successfully",
            "user_id": user_id,
            "product_id": product_id,
            "feedback_type": feedback_type
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error recording feedback: {str(e)}")

@router.get("/popular", response_model=RecommendationsListResponse)
async def get_popular_products(limit: int = Query(10, description="Number of popular products to return")):
    """Get popular products based on interaction count"""
    try:
        recommendations = await recommendation_engine._get_popular_items(limit)
        
        recommendation_responses = [
            RecommendationResponse(
                product_id=rec['product_id'],
                score=rec['score'],
                algorithm=rec['algorithm']
            )
            for rec in recommendations
        ]
        
        return RecommendationsListResponse(
            recommendations=recommendation_responses,
            user_id="anonymous",
            algorithm="popular",
            total=len(recommendation_responses)
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error getting popular products: {str(e)}")

@router.get("/similar/{product_id}")
async def get_similar_products(
    product_id: str,
    limit: int = Query(5, description="Number of similar products to return")
):
    """Get products similar to a given product using content-based filtering"""
    try:
        # Validate product_id
        try:
            uuid.UUID(product_id)
        except ValueError:
            raise HTTPException(status_code=400, detail="Invalid product_id format")
        
        if not recommendation_engine.is_trained or recommendation_engine.content_similarity_matrix is None:
            await recommendation_engine.train_models()
        
        if recommendation_engine.product_features is None:
            raise HTTPException(status_code=500, detail="Content-based model not available")
        
        product_uuid = uuid.UUID(product_id)
        
        if product_uuid not in recommendation_engine.product_features.index:
            raise HTTPException(status_code=404, detail="Product not found")
        
        # Get similar products
        product_idx = list(recommendation_engine.product_features.index).index(product_uuid)
        similarities = recommendation_engine.content_similarity_matrix[product_idx]
        
        # Get top similar products (excluding the product itself)
        similar_indices = similarities.argsort()[::-1][1:limit+1]
        
        similar_products = []
        for idx in similar_indices:
            similar_product_id = recommendation_engine.product_features.index[idx]
            similarity_score = similarities[idx]
            
            similar_products.append({
                'product_id': str(similar_product_id),
                'similarity_score': float(similarity_score)
            })
        
        return {
            "product_id": product_id,
            "similar_products": similar_products,
            "total": len(similar_products)
        }
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error getting similar products: {str(e)}") 