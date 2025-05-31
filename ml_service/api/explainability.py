"""
Explainability API Endpoints
Provides detailed explanations for ML model decisions
"""

from fastapi import APIRouter, HTTPException, Query
from typing import Dict, Any, Optional
import logging

from models.explainability import explainability_engine

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/explainability", tags=["explainability"])

@router.post("/explain")
async def explain_model_decision(
    model_type: str,
    algorithm: str,
    decision_data: Dict[str, Any]
) -> Dict[str, Any]:
    """
    Get comprehensive explanation for any ML model decision
    
    Args:
        model_type: Type of ML model (recommendations, search, anomaly_detection, etc.)
        algorithm: Specific algorithm used (collaborative, tfidf, isolation_forest, etc.)
        decision_data: The decision data including reasoning, scores, etc.
    
    Returns:
        Comprehensive explanation with reasoning, confidence, and business impact
    """
    try:
        explanation = explainability_engine.explain_model_decision(
            model_type=model_type,
            algorithm=algorithm,
            decision_data=decision_data
        )
        
        return {
            "status": "success",
            "explanation": explanation,
            "model_info": {
                "type": model_type,
                "algorithm": algorithm,
                "explainability_available": True
            }
        }
        
    except Exception as e:
        logger.error(f"Error generating explanation: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to generate explanation: {str(e)}")

@router.get("/models")
async def get_supported_models() -> Dict[str, Any]:
    """
    Get list of all supported models and their explanation capabilities
    """
    try:
        templates = explainability_engine.explanation_templates
        
        supported_models = {}
        for model_type, algorithms in templates.items():
            supported_models[model_type] = {
                "algorithms": list(algorithms.keys()),
                "description": f"Explainability support for {model_type.replace('_', ' ').title()}",
                "explanation_features": [
                    "Model overview and methodology",
                    "Decision reasoning and factors",
                    "Confidence analysis",
                    "Business impact assessment",
                    "Technical implementation details",
                    "Improvement suggestions"
                ]
            }
        
        return {
            "status": "success",
            "supported_models": supported_models,
            "total_model_types": len(supported_models),
            "explainability_version": "1.0"
        }
        
    except Exception as e:
        logger.error(f"Error getting supported models: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get supported models: {str(e)}")

@router.get("/confidence-thresholds")
async def get_confidence_thresholds() -> Dict[str, Any]:
    """
    Get confidence thresholds for all model types
    """
    try:
        return {
            "status": "success",
            "confidence_thresholds": explainability_engine.confidence_thresholds,
            "description": "Confidence level thresholds for determining model certainty"
        }
        
    except Exception as e:
        logger.error(f"Error getting confidence thresholds: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get confidence thresholds: {str(e)}")

@router.get("/business-impact")
async def get_business_impact_rules() -> Dict[str, Any]:
    """
    Get business impact assessment rules for all model types
    """
    try:
        return {
            "status": "success",
            "business_impact_rules": explainability_engine.business_impact_rules,
            "description": "Business impact assessment rules for each ML model type"
        }
        
    except Exception as e:
        logger.error(f"Error getting business impact rules: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to get business impact rules: {str(e)}")

@router.post("/batch-explain")
async def batch_explain_decisions(
    explanations_request: Dict[str, Any]
) -> Dict[str, Any]:
    """
    Generate explanations for multiple ML model decisions in batch
    
    Args:
        explanations_request: Dictionary containing multiple decision requests
        Format: {
            "decisions": [
                {
                    "id": "unique_id",
                    "model_type": "recommendations",
                    "algorithm": "hybrid",
                    "decision_data": {...}
                },
                ...
            ]
        }
    """
    try:
        decisions = explanations_request.get("decisions", [])
        
        if not decisions:
            raise HTTPException(status_code=400, detail="No decisions provided for explanation")
        
        if len(decisions) > 50:  # Limit batch size
            raise HTTPException(status_code=400, detail="Batch size too large. Maximum 50 decisions per request")
        
        explanations = {}
        errors = {}
        
        for decision in decisions:
            decision_id = decision.get("id", f"decision_{len(explanations)}")
            
            try:
                explanation = explainability_engine.explain_model_decision(
                    model_type=decision["model_type"],
                    algorithm=decision["algorithm"],
                    decision_data=decision["decision_data"]
                )
                
                explanations[decision_id] = explanation
                
            except Exception as e:
                errors[decision_id] = str(e)
                logger.error(f"Error explaining decision {decision_id}: {e}")
        
        return {
            "status": "success",
            "explanations": explanations,
            "errors": errors,
            "total_processed": len(decisions),
            "successful_explanations": len(explanations),
            "failed_explanations": len(errors)
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in batch explanation: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to process batch explanations: {str(e)}")

@router.get("/model-comparison")
async def compare_model_approaches(
    model_type: str = Query(..., description="Model type to compare approaches for")
) -> Dict[str, Any]:
    """
    Compare different algorithmic approaches for a specific model type
    """
    try:
        templates = explainability_engine.explanation_templates.get(model_type, {})
        
        if not templates:
            raise HTTPException(status_code=404, detail=f"Model type '{model_type}' not found")
        
        comparison = {
            "model_type": model_type,
            "available_algorithms": {},
            "comparison_matrix": {},
            "recommendations": {}
        }
        
        # Get algorithm details
        for algorithm, template in templates.items():
            comparison["available_algorithms"][algorithm] = {
                "name": template.get("method_name", algorithm),
                "description": template.get("description", ""),
                "strengths": template.get("strengths", []),
                "limitations": template.get("limitations", []),
                "use_cases": template.get("use_cases", [])
            }
        
        # Create comparison matrix
        comparison_criteria = ["accuracy", "interpretability", "scalability", "cold_start_handling", "computational_cost"]
        
        for algorithm in templates.keys():
            comparison["comparison_matrix"][algorithm] = {}
            
            # Simplified scoring based on algorithm characteristics
            if algorithm == "collaborative":
                scores = {"accuracy": "High", "interpretability": "Medium", "scalability": "Medium", 
                         "cold_start_handling": "Poor", "computational_cost": "Medium"}
            elif algorithm == "content_based":
                scores = {"accuracy": "Medium", "interpretability": "High", "scalability": "High", 
                         "cold_start_handling": "Good", "computational_cost": "Low"}
            elif algorithm == "hybrid":
                scores = {"accuracy": "Very High", "interpretability": "Medium", "scalability": "Medium", 
                         "cold_start_handling": "Good", "computational_cost": "High"}
            elif algorithm == "tfidf":
                scores = {"accuracy": "Medium", "interpretability": "High", "scalability": "High", 
                         "cold_start_handling": "Good", "computational_cost": "Low"}
            elif algorithm == "isolation_forest":
                scores = {"accuracy": "High", "interpretability": "Low", "scalability": "High", 
                         "cold_start_handling": "Good", "computational_cost": "Medium"}
            else:
                scores = {"accuracy": "Medium", "interpretability": "Medium", "scalability": "Medium", 
                         "cold_start_handling": "Medium", "computational_cost": "Medium"}
            
            comparison["comparison_matrix"][algorithm] = scores
        
        # Generate recommendations
        if len(templates) > 1:
            comparison["recommendations"] = {
                "production_use": "hybrid" if "hybrid" in templates else list(templates.keys())[0],
                "development_start": "content_based" if "content_based" in templates else list(templates.keys())[0],
                "high_scale": "tfidf" if "tfidf" in templates else list(templates.keys())[0],
                "best_accuracy": "hybrid" if "hybrid" in templates else "collaborative" if "collaborative" in templates else list(templates.keys())[0]
            }
        
        return {
            "status": "success",
            "comparison": comparison
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error comparing model approaches: {e}")
        raise HTTPException(status_code=500, detail=f"Failed to compare model approaches: {str(e)}")

@router.get("/health")
async def health_check() -> Dict[str, Any]:
    """Health check for explainability service"""
    try:
        # Test basic functionality
        test_decision = {
            "score": 0.75,
            "reasoning": {
                "method": "Test Method",
                "explanation": "Test explanation",
                "confidence": "High"
            }
        }
        
        test_explanation = explainability_engine.explain_model_decision(
            model_type="recommendations",
            algorithm="hybrid",
            decision_data=test_decision
        )
        
        return {
            "status": "healthy",
            "explainability_engine": "operational",
            "supported_model_types": len(explainability_engine.explanation_templates),
            "test_explanation_generated": bool(test_explanation),
            "version": "1.0"
        }
        
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        return {
            "status": "unhealthy",
            "error": str(e),
            "explainability_engine": "error"
        } 