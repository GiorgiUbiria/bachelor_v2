from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel
from typing import Optional, Dict, Any
import logging

from models.anomaly_detection import anomaly_detector

logger = logging.getLogger(__name__)
router = APIRouter()

class RequestAnalysisRequest(BaseModel):
    ip_address: str
    user_agent: str
    method: str
    path: str
    query_params: Optional[str] = "{}"
    status_code: int
    response_time: float
    request_size: Optional[int] = 0
    response_size: Optional[int] = 0
    user_id: Optional[str] = None

@router.get("/")
async def anomaly_detection_status():
    """Get anomaly detection service status"""
    return {
        "service": "anomaly_detection",
        "status": "active",
        "is_trained": anomaly_detector.is_trained,
        "description": "Request Analysis for Anomalies - Security monitoring and threat detection"
    }

@router.post("/initialize")
async def initialize_anomaly_detector():
    """Initialize or retrain the anomaly detector"""
    try:
        await anomaly_detector.initialize()
        return {
            "message": "Anomaly detector initialized successfully",
            "requests_analyzed": len(anomaly_detector.request_data) if hasattr(anomaly_detector, 'request_data') else 0,
            "baseline_established": bool(anomaly_detector.baseline_metrics)
        }
    except Exception as e:
        logger.error(f"Failed to initialize anomaly detector: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/analyze")
async def analyze_request(request: RequestAnalysisRequest):
    """Analyze a single request for anomalies"""
    try:
        # Convert request to dictionary
        request_data = request.dict()
        
        # Perform anomaly analysis
        result = await anomaly_detector.analyze_request(request_data)
        
        return {
            "success": True,
            "data": result
        }
        
    except Exception as e:
        logger.error(f"Failed to analyze request: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/analyze/batch")
async def analyze_requests_batch(requests: list[RequestAnalysisRequest]):
    """Analyze multiple requests for anomalies"""
    try:
        results = []
        
        for req in requests:
            request_data = req.dict()
            analysis = await anomaly_detector.analyze_request(request_data)
            results.append(analysis)
        
        # Calculate summary statistics
        total_requests = len(results)
        anomalous_requests = len([r for r in results if r['is_anomaly']])
        high_risk_requests = len([r for r in results if r['risk_level'] in ['high', 'critical']])
        
        return {
            "success": True,
            "data": {
                "results": results,
                "summary": {
                    "total_requests": total_requests,
                    "anomalous_requests": anomalous_requests,
                    "high_risk_requests": high_risk_requests,
                    "anomaly_rate": round(anomalous_requests / total_requests, 3) if total_requests > 0 else 0
                }
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to analyze batch requests: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/insights")
async def get_anomaly_insights():
    """Get insights about detected anomalies"""
    try:
        result = await anomaly_detector.get_anomaly_insights()
        return {
            "success": True,
            "data": result
        }
    except Exception as e:
        logger.error(f"Failed to get anomaly insights: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/dashboard")
async def get_security_dashboard():
    """Get comprehensive security dashboard data"""
    try:
        result = await anomaly_detector.get_security_dashboard()
        return {
            "success": True,
            "data": result
        }
    except Exception as e:
        logger.error(f"Failed to get security dashboard: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/baseline")
async def get_baseline_metrics():
    """Get baseline metrics used for anomaly detection"""
    try:
        if not anomaly_detector.is_trained:
            await anomaly_detector.initialize()
            
        return {
            "success": True,
            "data": {
                "baseline_metrics": anomaly_detector.baseline_metrics,
                "suspicious_patterns": anomaly_detector.suspicious_patterns,
                "whitelist_ips": list(anomaly_detector.whitelist_ips),
                "blacklist_ips": list(anomaly_detector.blacklist_ips)
            }
        }
    except Exception as e:
        logger.error(f"Failed to get baseline metrics: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/simulate/attack")
async def simulate_attack_detection(
    attack_type: str = Query("sql_injection", description="Type of attack to simulate: sql_injection, xss, brute_force, ddos")
):
    """Simulate different types of attacks for demonstration"""
    try:
        # Generate sample attack requests based on type
        if attack_type == "sql_injection":
            sample_request = {
                "ip_address": "192.168.1.100",
                "user_agent": "sqlmap/1.0",
                "method": "GET",
                "path": "/api/products",
                "query_params": '{"id": "1\' OR 1=1--"}',
                "status_code": 500,
                "response_time": 1500,
                "request_size": 2048,
                "response_size": 512,
                "user_id": None
            }
        elif attack_type == "xss":
            sample_request = {
                "ip_address": "10.0.0.50",
                "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
                "method": "POST",
                "path": "/api/comments",
                "query_params": '{"content": "<script>alert(document.cookie)</script>"}',
                "status_code": 400,
                "response_time": 300,
                "request_size": 1024,
                "response_size": 256,
                "user_id": "user_123"
            }
        elif attack_type == "brute_force":
            sample_request = {
                "ip_address": "203.0.113.10",
                "user_agent": "Mozilla/5.0 (X11; Linux x86_64)",
                "method": "POST",
                "path": "/api/auth/login",
                "query_params": '{"username": "admin", "password": "password123"}',
                "status_code": 401,
                "response_time": 200,
                "request_size": 512,
                "response_size": 128,
                "user_id": None
            }
        elif attack_type == "ddos":
            sample_request = {
                "ip_address": "198.51.100.25",
                "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
                "method": "GET",
                "path": "/api/products",
                "query_params": "{}",
                "status_code": 200,
                "response_time": 5000,
                "request_size": 4096,
                "response_size": 8192,
                "user_id": None
            }
        else:
            raise HTTPException(status_code=400, detail="Invalid attack type")
        
        # Analyze the simulated attack
        result = await anomaly_detector.analyze_request(sample_request)
        
        return {
            "success": True,
            "data": {
                "attack_type": attack_type,
                "simulated_request": sample_request,
                "analysis_result": result,
                "demonstration": f"This simulates a {attack_type} attack and shows how the system detects it"
            }
        }
        
    except Exception as e:
        logger.error(f"Failed to simulate attack: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/patterns")
async def get_attack_patterns():
    """Get information about attack patterns the system can detect"""
    try:
        if not anomaly_detector.is_trained:
            await anomaly_detector.initialize()
            
        return {
            "success": True,
            "data": {
                "detection_methods": [
                    "IP-based analysis (blacklists, whitelists, suspicious ranges)",
                    "Pattern-based detection (SQL injection, XSS, path traversal)",
                    "Statistical analysis (response times, request sizes, error rates)",
                    "Machine Learning (Isolation Forest for outlier detection)",
                    "Behavioral analysis (time patterns, authentication context)"
                ],
                "attack_types_detected": [
                    "SQL Injection",
                    "Cross-Site Scripting (XSS)",
                    "Path Traversal",
                    "Brute Force Attacks",
                    "DDoS/High Volume Attacks",
                    "Admin Panel Access Attempts",
                    "Suspicious User Agents",
                    "Unusual Traffic Patterns"
                ],
                "suspicious_patterns": anomaly_detector.suspicious_patterns,
                "risk_levels": ["low", "medium", "high", "critical"]
            }
        }
    except Exception as e:
        logger.error(f"Failed to get attack patterns: {e}")
        raise HTTPException(status_code=500, detail=str(e)) 