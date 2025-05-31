import asyncio
import logging
from typing import List, Dict, Any, Optional, Tuple
import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler, LabelEncoder
from sklearn.cluster import DBSCAN
from datetime import datetime, timedelta
import re
import json
from collections import Counter, defaultdict
import ipaddress
from urllib.parse import urlparse, parse_qs

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class AnomalyDetector:
    def __init__(self):
        self.is_trained = False
        self.isolation_forest = None
        self.scaler = StandardScaler()
        self.label_encoders = {}
        self.request_patterns = {}
        self.baseline_metrics = {}
        self.suspicious_patterns = {}
        self.whitelist_ips = set()
        self.blacklist_ips = set()
        
    async def initialize(self):
        """Initialize the anomaly detector"""
        try:
            await self._load_request_logs()
            await self._load_security_data()
            await self._build_baseline_patterns()
            await self._train_anomaly_models()
            self.is_trained = True
            logger.info("Anomaly detector initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize anomaly detector: {e}")
            raise

    async def _load_request_logs(self):
        """Load HTTP request logs from database"""
        conn = await get_db_connection()
        try:
            # Load request logs - assuming we have a request_logs table
            query = """
            SELECT 
                id,
                ip_address,
                user_agent,
                method,
                path,
                query_params,
                status_code,
                response_time,
                request_size,
                response_size,
                timestamp,
                user_id
            FROM request_logs 
            WHERE timestamp >= $1
            ORDER BY timestamp DESC
            LIMIT 10000
            """
            
            cutoff_date = datetime.now() - timedelta(days=30)
            rows = await conn.fetch(query, cutoff_date)
            
            if rows:
                self.request_data = pd.DataFrame([dict(row) for row in rows])
            else:
                # Create sample data if no logs exist
                self.request_data = self._generate_sample_request_data()
            
            logger.info(f"Loaded {len(self.request_data)} request records for anomaly detection")
        except Exception as e:
            logger.warning(f"Could not load request logs from database: {e}")
            # Generate sample data for demonstration
            self.request_data = self._generate_sample_request_data()
        finally:
            await release_db_connection(conn)

    def _generate_sample_request_data(self) -> pd.DataFrame:
        """Generate sample request data for demonstration"""
        np.random.seed(42)
        
        # Normal request patterns
        normal_ips = [f"192.168.1.{i}" for i in range(1, 50)]
        normal_paths = ['/api/products', '/api/users', '/api/orders', '/api/search', '/api/recommendations']
        normal_methods = ['GET', 'POST', 'PUT', 'DELETE']
        normal_user_agents = [
            'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
            'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36',
            'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36'
        ]
        
        # Generate normal requests
        normal_requests = []
        for i in range(800):
            timestamp = datetime.now() - timedelta(
                days=np.random.randint(0, 30),
                hours=np.random.randint(0, 24),
                minutes=np.random.randint(0, 60)
            )
            
            normal_requests.append({
                'id': i + 1,
                'ip_address': np.random.choice(normal_ips),
                'user_agent': np.random.choice(normal_user_agents),
                'method': np.random.choice(normal_methods, p=[0.6, 0.2, 0.1, 0.1]),
                'path': np.random.choice(normal_paths),
                'query_params': '{}' if np.random.random() > 0.3 else '{"limit": "10"}',
                'status_code': np.random.choice([200, 201, 404], p=[0.8, 0.1, 0.1]),
                'response_time': np.random.normal(200, 50),
                'request_size': np.random.normal(1024, 200),
                'response_size': np.random.normal(5120, 1000),
                'timestamp': timestamp,
                'user_id': f"user_{np.random.randint(1, 100)}" if np.random.random() > 0.2 else None
            })
        
        # Generate anomalous requests
        anomalous_requests = []
        for i in range(200):
            timestamp = datetime.now() - timedelta(
                days=np.random.randint(0, 30),
                hours=np.random.randint(0, 24),
                minutes=np.random.randint(0, 60)
            )
            
            # Different types of anomalies
            anomaly_type = np.random.choice(['sql_injection', 'brute_force', 'ddos', 'suspicious_path'])
            
            if anomaly_type == 'sql_injection':
                path = '/api/products'
                query_params = '{"id": "1\' OR 1=1--"}'
                status_code = 500
            elif anomaly_type == 'brute_force':
                path = '/api/auth/login'
                query_params = '{}'
                status_code = 401
            elif anomaly_type == 'ddos':
                path = np.random.choice(normal_paths)
                query_params = '{}'
                status_code = 200
            else:  # suspicious_path
                path = '/admin/config'
                query_params = '{}'
                status_code = 403
            
            anomalous_requests.append({
                'id': 800 + i + 1,
                'ip_address': f"10.0.0.{np.random.randint(1, 255)}" if anomaly_type == 'ddos' else np.random.choice(normal_ips),
                'user_agent': 'sqlmap/1.0' if anomaly_type == 'sql_injection' else np.random.choice(normal_user_agents),
                'method': 'POST' if anomaly_type == 'brute_force' else np.random.choice(normal_methods),
                'path': path,
                'query_params': query_params,
                'status_code': status_code,
                'response_time': np.random.normal(1000, 200) if anomaly_type == 'ddos' else np.random.normal(200, 50),
                'request_size': np.random.normal(2048, 400),
                'response_size': np.random.normal(1024, 200),
                'timestamp': timestamp,
                'user_id': None if anomaly_type in ['sql_injection', 'ddos'] else f"user_{np.random.randint(1, 100)}"
            })
        
        all_requests = normal_requests + anomalous_requests
        return pd.DataFrame(all_requests)

    async def _load_security_data(self):
        """Load security-related data like IP whitelists/blacklists"""
        # In a real implementation, this would load from security databases
        # For now, we'll use some sample data
        
        # Known good IPs (internal network)
        self.whitelist_ips = {
            '192.168.1.1', '192.168.1.10', '192.168.1.20',
            '10.0.0.1', '172.16.0.1'
        }
        
        # Known bad IPs (from threat intelligence)
        self.blacklist_ips = {
            '1.2.3.4', '5.6.7.8', '9.10.11.12'
        }

    async def _build_baseline_patterns(self):
        """Build baseline patterns from normal traffic"""
        if self.request_data.empty:
            return
            
        # Convert decimal columns to float to avoid type mismatch issues
        numeric_columns = ['response_time', 'request_size', 'response_size']
        for col in numeric_columns:
            if col in self.request_data.columns:
                self.request_data[col] = pd.to_numeric(self.request_data[col], errors='coerce').fillna(0)
            
        # Calculate baseline metrics
        self.baseline_metrics = {
            'avg_response_time': float(self.request_data['response_time'].mean()),
            'std_response_time': float(self.request_data['response_time'].std()),
            'avg_request_size': float(self.request_data['request_size'].mean()),
            'std_request_size': float(self.request_data['request_size'].std()),
            'common_paths': self.request_data['path'].value_counts().head(20).to_dict(),
            'common_user_agents': self.request_data['user_agent'].value_counts().head(10).to_dict(),
            'status_code_distribution': self.request_data['status_code'].value_counts().to_dict(),
            'method_distribution': self.request_data['method'].value_counts().to_dict()
        }
        
        # Analyze request patterns by hour
        self.request_data['hour'] = pd.to_datetime(self.request_data['timestamp']).dt.hour
        hourly_patterns = self.request_data.groupby('hour').size()
        self.baseline_metrics['hourly_pattern'] = hourly_patterns.to_dict()
        
        # Identify suspicious patterns
        self._identify_suspicious_patterns()

    def _identify_suspicious_patterns(self):
        """Identify patterns that might indicate attacks"""
        self.suspicious_patterns = {
            'sql_injection_keywords': [
                'union', 'select', 'drop', 'insert', 'update', 'delete',
                'or 1=1', 'or 1=2', '--', ';--', '/*', '*/', 'xp_',
                'sp_', 'exec', 'execute', 'cast', 'convert'
            ],
            'xss_patterns': [
                '<script', '</script>', 'javascript:', 'onload=',
                'onerror=', 'onclick=', 'alert(', 'document.cookie'
            ],
            'path_traversal': [
                '../', '..\\', '/etc/passwd', '/etc/shadow',
                'boot.ini', 'win.ini', 'system32'
            ],
            'suspicious_user_agents': [
                'sqlmap', 'nikto', 'nmap', 'masscan', 'zap',
                'burp', 'w3af', 'acunetix', 'nessus'
            ],
            'admin_paths': [
                '/admin', '/administrator', '/wp-admin', '/phpmyadmin',
                '/config', '/backup', '/test', '/debug'
            ]
        }

    async def _train_anomaly_models(self):
        """Train machine learning models for anomaly detection"""
        if self.request_data.empty:
            return
            
        # Prepare features for ML models
        features_df = self._extract_features(self.request_data)
        
        # Train Isolation Forest for outlier detection
        self.isolation_forest = IsolationForest(
            contamination=0.1,  # Expect 10% anomalies
            random_state=42,
            n_estimators=100
        )
        
        # Scale numerical features
        numerical_features = ['response_time', 'request_size', 'response_size', 'hour']
        features_scaled = self.scaler.fit_transform(features_df[numerical_features])
        
        # Train the model
        self.isolation_forest.fit(features_scaled)
        
        logger.info("Anomaly detection models trained successfully")

    def _extract_features(self, data: pd.DataFrame) -> pd.DataFrame:
        """Extract features from request data for ML models"""
        features = data.copy()
        
        # Convert decimal columns to float to avoid type mismatch issues
        numeric_columns = ['response_time', 'request_size', 'response_size']
        for col in numeric_columns:
            if col in features.columns:
                features[col] = pd.to_numeric(features[col], errors='coerce').fillna(0)
        
        # Add time-based features
        features['hour'] = pd.to_datetime(features['timestamp']).dt.hour
        features['day_of_week'] = pd.to_datetime(features['timestamp']).dt.dayofweek
        
        # Add categorical encodings
        for col in ['method', 'path']:
            if col not in self.label_encoders:
                self.label_encoders[col] = LabelEncoder()
                features[f'{col}_encoded'] = self.label_encoders[col].fit_transform(features[col].astype(str))
            else:
                # Handle unseen categories
                unique_values = self.label_encoders[col].classes_
                features[f'{col}_encoded'] = features[col].apply(
                    lambda x: self.label_encoders[col].transform([x])[0] 
                    if x in unique_values else -1
                )
        
        # Add derived features
        features['is_error'] = (features['status_code'] >= 400).astype(int)
        
        # Safely calculate z-score with proper type conversion
        avg_response_time = float(self.baseline_metrics.get('avg_response_time', 200))
        std_response_time = float(self.baseline_metrics.get('std_response_time', 50))
        
        features['response_time_zscore'] = np.abs(
            (features['response_time'].astype(float) - avg_response_time) / std_response_time
        )
        
        return features

    async def analyze_request(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze a single request for anomalies with comprehensive reasoning"""
        if not self.is_trained:
            await self.initialize()
            
        anomaly_score = 0.0
        anomaly_reasons = []
        risk_level = 'low'
        detailed_analysis = {}
        
        # 1. IP-based analysis
        ip_analysis = self._analyze_ip(request_data.get('ip_address', ''))
        anomaly_score += ip_analysis['score']
        if ip_analysis['reasons']:
            anomaly_reasons.extend(ip_analysis['reasons'])
        detailed_analysis['ip_analysis'] = ip_analysis
        
        # 2. Pattern-based analysis
        pattern_analysis = self._analyze_patterns(request_data)
        anomaly_score += pattern_analysis['score']
        if pattern_analysis['reasons']:
            anomaly_reasons.extend(pattern_analysis['reasons'])
        detailed_analysis['pattern_analysis'] = pattern_analysis
        
        # 3. Statistical analysis
        stats_analysis = self._analyze_statistics(request_data)
        anomaly_score += stats_analysis['score']
        if stats_analysis['reasons']:
            anomaly_reasons.extend(stats_analysis['reasons'])
        detailed_analysis['stats_analysis'] = stats_analysis
        
        # 4. ML-based analysis
        ml_analysis = {'score': 0.0, 'reasons': []}
        if self.isolation_forest:
            ml_analysis = self._analyze_with_ml(request_data)
            anomaly_score += ml_analysis['score']
            if ml_analysis['reasons']:
                anomaly_reasons.extend(ml_analysis['reasons'])
        detailed_analysis['ml_analysis'] = ml_analysis
        
        # 5. Behavioral analysis
        behavior_analysis = await self._analyze_behavior(request_data)
        anomaly_score += behavior_analysis['score']
        if behavior_analysis['reasons']:
            anomaly_reasons.extend(behavior_analysis['reasons'])
        detailed_analysis['behavior_analysis'] = behavior_analysis
        
        # Determine risk level with reasoning
        risk_thresholds = {
            'critical': 0.8,
            'high': 0.6,
            'medium': 0.4,
            'low': 0.0
        }
        
        for level, threshold in risk_thresholds.items():
            if anomaly_score >= threshold:
                risk_level = level
                break
        
        # Generate comprehensive reasoning
        reasoning = self._generate_comprehensive_reasoning(
            anomaly_score, anomaly_reasons, detailed_analysis, risk_level
        )
        
        # Generate recommendations
        recommendations = self._generate_recommendations(anomaly_score, anomaly_reasons)
        
        return {
            'request_id': request_data.get('id', 'unknown'),
            'anomaly_score': round(anomaly_score, 3),
            'risk_level': risk_level,
            'is_anomaly': anomaly_score > 0.5,
            'anomaly_reasons': anomaly_reasons,
            'recommendations': recommendations,
            'analysis_timestamp': datetime.now().isoformat(),
            'reasoning': reasoning,
            'detailed_analysis': detailed_analysis
        }

    def _generate_comprehensive_reasoning(self, anomaly_score: float, reasons: List[str], 
                                        detailed_analysis: Dict, risk_level: str) -> Dict[str, Any]:
        """Generate comprehensive reasoning for anomaly detection decision"""
        
        # Count analysis contributions
        analysis_contributions = {}
        for analysis_type, analysis_data in detailed_analysis.items():
            score = analysis_data.get('score', 0)
            analysis_contributions[analysis_type] = {
                'score': score,
                'percentage': (score / anomaly_score * 100) if anomaly_score > 0 else 0,
                'reasons_count': len(analysis_data.get('reasons', []))
            }
        
        # Identify primary detection method
        primary_method = max(analysis_contributions.items(), key=lambda x: x[1]['score'])
        
        # Create explanation based on primary detection method
        method_explanations = {
            'ip_analysis': 'IP address reputation and geolocation analysis',
            'pattern_analysis': 'Known attack pattern recognition',
            'stats_analysis': 'Statistical deviation from normal behavior',
            'ml_analysis': 'Machine learning anomaly detection',
            'behavior_analysis': 'Behavioral pattern analysis'
        }
        
        primary_explanation = method_explanations.get(primary_method[0], 'Multi-factor analysis')
        
        # Determine confidence
        confidence = 'High' if anomaly_score > 0.7 and len(reasons) >= 3 else \
                    'Medium' if anomaly_score > 0.4 and len(reasons) >= 2 else \
                    'Low'
        
        return {
            'method': 'Multi-Layered Anomaly Detection',
            'explanation': f'Primary detection via {primary_explanation} (score: {primary_method[1]["score"]:.3f})',
            'confidence': confidence,
            'factors': [
                f'Total anomaly score: {anomaly_score:.3f}',
                f'Risk level: {risk_level.upper()}',
                f'Detection methods triggered: {len([a for a in analysis_contributions.values() if a["score"] > 0])}',
                f'Primary method: {primary_method[0].replace("_", " ").title()}'
            ],
            'technical_details': {
                'algorithm': 'Ensemble of 5 detection methods',
                'methods_used': list(detailed_analysis.keys()),
                'scoring_method': 'Additive scoring across all methods',
                'risk_thresholds': {'low': '0.0-0.4', 'medium': '0.4-0.6', 'high': '0.6-0.8', 'critical': '0.8+'},
                'baseline_established': bool(self.baseline_metrics)
            },
            'analysis_breakdown': analysis_contributions,
            'detection_timeline': {
                'analysis_order': ['IP Analysis', 'Pattern Recognition', 'Statistical Analysis', 'ML Detection', 'Behavioral Analysis'],
                'primary_trigger': primary_method[0].replace('_', ' ').title()
            }
        }

    def _analyze_ip(self, ip_address: str) -> Dict[str, Any]:
        """Analyze IP address for anomalies"""
        score = 0.0
        reasons = []
        
        if not ip_address:
            return {'score': 0.1, 'reasons': ['Missing IP address']}
        
        # Check blacklist
        if ip_address in self.blacklist_ips:
            score += 0.9
            reasons.append(f'IP {ip_address} is blacklisted')
        
        # Check whitelist
        if ip_address in self.whitelist_ips:
            score -= 0.2  # Reduce suspicion for whitelisted IPs
            reasons.append(f'IP {ip_address} is whitelisted')
        
        # Check if IP is from suspicious ranges
        try:
            ip_obj = ipaddress.ip_address(ip_address)
            
            # Check for private vs public IP patterns
            if ip_obj.is_private:
                # Private IPs are generally less suspicious
                score -= 0.1
            else:
                # Public IPs might be more suspicious depending on context
                score += 0.1
                
            # Check for specific suspicious ranges (example)
            suspicious_ranges = ['1.2.3.0/24', '5.6.7.0/24']
            for range_str in suspicious_ranges:
                if ip_obj in ipaddress.ip_network(range_str):
                    score += 0.5
                    reasons.append(f'IP {ip_address} is in suspicious range {range_str}')
                    
        except ValueError:
            score += 0.3
            reasons.append(f'Invalid IP address format: {ip_address}')
        
        return {'score': min(score, 1.0), 'reasons': reasons}

    def _analyze_patterns(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze request patterns for known attack signatures"""
        score = 0.0
        reasons = []
        
        path = request_data.get('path', '').lower()
        query_params = request_data.get('query_params', '{}')
        user_agent = request_data.get('user_agent', '').lower()
        
        # Check for SQL injection patterns
        combined_text = f"{path} {query_params}".lower()
        for keyword in self.suspicious_patterns['sql_injection_keywords']:
            if keyword in combined_text:
                score += 0.3
                reasons.append(f'SQL injection pattern detected: {keyword}')
        
        # Check for XSS patterns
        for pattern in self.suspicious_patterns['xss_patterns']:
            if pattern in combined_text:
                score += 0.3
                reasons.append(f'XSS pattern detected: {pattern}')
        
        # Check for path traversal
        for pattern in self.suspicious_patterns['path_traversal']:
            if pattern in combined_text:
                score += 0.4
                reasons.append(f'Path traversal pattern detected: {pattern}')
        
        # Check for suspicious user agents
        for agent in self.suspicious_patterns['suspicious_user_agents']:
            if agent in user_agent:
                score += 0.5
                reasons.append(f'Suspicious user agent detected: {agent}')
        
        # Check for admin path access
        for admin_path in self.suspicious_patterns['admin_paths']:
            if admin_path in path:
                score += 0.2
                reasons.append(f'Admin path access detected: {admin_path}')
        
        return {'score': min(score, 1.0), 'reasons': reasons}

    def _analyze_statistics(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze request using statistical methods"""
        score = 0.0
        reasons = []
        
        # Check response time anomalies
        response_time = float(request_data.get('response_time', 0))
        avg_response_time = float(self.baseline_metrics.get('avg_response_time', 200))
        std_response_time = float(self.baseline_metrics.get('std_response_time', 50))
        
        if response_time > avg_response_time + 3 * std_response_time:
            score += 0.3
            reasons.append(f'Unusually high response time: {response_time}ms')
        
        # Check request size anomalies
        request_size = float(request_data.get('request_size', 0))
        avg_request_size = float(self.baseline_metrics.get('avg_request_size', 1024))
        std_request_size = float(self.baseline_metrics.get('std_request_size', 200))
        
        if request_size > avg_request_size + 3 * std_request_size:
            score += 0.2
            reasons.append(f'Unusually large request size: {request_size} bytes')
        
        # Check status code patterns
        status_code = request_data.get('status_code', 200)
        if status_code >= 500:
            score += 0.2
            reasons.append(f'Server error status code: {status_code}')
        elif status_code == 401 or status_code == 403:
            score += 0.1
            reasons.append(f'Authentication/authorization error: {status_code}')
        
        return {'score': min(score, 1.0), 'reasons': reasons}

    def _analyze_with_ml(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze request using trained ML models"""
        score = 0.0
        reasons = []
        
        try:
            # Convert request to feature format
            temp_df = pd.DataFrame([request_data])
            features_df = self._extract_features(temp_df)
            
            # Extract numerical features for ML model
            numerical_features = ['response_time', 'request_size', 'response_size', 'hour']
            
            # Handle missing features
            for feature in numerical_features:
                if feature not in features_df.columns:
                    if feature == 'hour':
                        features_df[feature] = datetime.now().hour
                    else:
                        features_df[feature] = 0
            
            features_scaled = self.scaler.transform(features_df[numerical_features])
            
            # Get anomaly score from Isolation Forest
            anomaly_score = self.isolation_forest.decision_function(features_scaled)[0]
            is_outlier = self.isolation_forest.predict(features_scaled)[0] == -1
            
            if is_outlier:
                score += 0.4
                reasons.append(f'ML model detected anomaly (score: {anomaly_score:.3f})')
            
        except Exception as e:
            logger.warning(f"ML analysis failed: {e}")
            score += 0.1
            reasons.append('ML analysis failed - using fallback detection')
        
        return {'score': min(score, 1.0), 'reasons': reasons}

    async def _analyze_behavior(self, request_data: Dict[str, Any]) -> Dict[str, Any]:
        """Analyze behavioral patterns"""
        score = 0.0
        reasons = []
        
        ip_address = request_data.get('ip_address', '')
        user_id = request_data.get('user_id')
        path = request_data.get('path', '')
        
        # Check for rapid requests from same IP (simple rate limiting check)
        # In a real implementation, this would check recent request history
        current_hour = datetime.now().hour
        
        # Check for unusual time patterns
        if current_hour < 6 or current_hour > 22:  # Outside normal business hours
            score += 0.1
            reasons.append('Request outside normal business hours')
        
        # Check for missing user context on protected endpoints
        if '/api/' in path and not user_id and path not in ['/api/auth/login', '/api/auth/register']:
            score += 0.2
            reasons.append('Unauthenticated access to protected endpoint')
        
        return {'score': min(score, 1.0), 'reasons': reasons}

    def _generate_recommendations(self, anomaly_score: float, reasons: List[str]) -> List[str]:
        """Generate security recommendations based on analysis"""
        recommendations = []
        
        if anomaly_score >= 0.8:
            recommendations.extend([
                'IMMEDIATE ACTION: Block this IP address',
                'Review and strengthen authentication mechanisms',
                'Implement additional monitoring for this request pattern'
            ])
        elif anomaly_score >= 0.6:
            recommendations.extend([
                'Increase monitoring for this IP address',
                'Consider implementing CAPTCHA for suspicious requests',
                'Review access logs for similar patterns'
            ])
        elif anomaly_score >= 0.4:
            recommendations.extend([
                'Monitor this request pattern',
                'Consider rate limiting for this IP',
                'Review security policies'
            ])
        
        # Specific recommendations based on detected patterns
        if any('SQL injection' in reason for reason in reasons):
            recommendations.append('Implement parameterized queries and input validation')
        
        if any('XSS' in reason for reason in reasons):
            recommendations.append('Implement output encoding and CSP headers')
        
        if any('suspicious user agent' in reason for reason in reasons):
            recommendations.append('Block known attack tool user agents')
        
        return recommendations

    async def get_anomaly_insights(self) -> Dict[str, Any]:
        """Get insights about detected anomalies"""
        if not self.is_trained:
            await self.initialize()
            
        # Analyze recent anomalies (this would typically query a database)
        insights = []
        
        # Sample insights based on baseline data
        total_requests = len(self.request_data)
        error_rate = len(self.request_data[self.request_data['status_code'] >= 400]) / total_requests
        
        if error_rate > 0.1:
            insights.append({
                'type': 'warning',
                'title': 'High Error Rate Detected',
                'description': f'Error rate is {error_rate:.1%}, which is above normal threshold',
                'recommendation': 'Investigate server issues or potential attacks'
            })
        
        # Check for unusual traffic patterns
        hourly_distribution = self.request_data.groupby(
            pd.to_datetime(self.request_data['timestamp']).dt.hour
        ).size()
        
        if hourly_distribution.std() > hourly_distribution.mean():
            insights.append({
                'type': 'info',
                'title': 'Irregular Traffic Patterns',
                'description': 'Traffic distribution shows unusual patterns',
                'recommendation': 'Review traffic sources and implement rate limiting'
            })
        
        return {
            'insights': insights,
            'summary': {
                'total_requests_analyzed': total_requests,
                'error_rate': round(error_rate, 3),
                'baseline_established': True,
                'models_trained': self.is_trained
            },
            'generated_at': datetime.now().isoformat()
        }

    async def get_security_dashboard(self) -> Dict[str, Any]:
        """Get comprehensive security dashboard data"""
        if not self.is_trained:
            await self.initialize()
            
        # Calculate security metrics
        total_requests = len(self.request_data)
        
        # Status code distribution
        status_distribution = self.request_data['status_code'].value_counts().to_dict()
        
        # Top IPs by request count
        top_ips = self.request_data['ip_address'].value_counts().head(10).to_dict()
        
        # Top paths by request count
        top_paths = self.request_data['path'].value_counts().head(10).to_dict()
        
        # Error rate over time
        self.request_data['date'] = pd.to_datetime(self.request_data['timestamp']).dt.date
        daily_errors = self.request_data[self.request_data['status_code'] >= 400].groupby('date').size()
        daily_total = self.request_data.groupby('date').size()
        error_rate_trend = (daily_errors / daily_total).fillna(0).to_dict()
        
        return {
            'summary': {
                'total_requests': total_requests,
                'unique_ips': self.request_data['ip_address'].nunique(),
                'error_rate': len(self.request_data[self.request_data['status_code'] >= 400]) / total_requests,
                'avg_response_time': self.request_data['response_time'].mean()
            },
            'distributions': {
                'status_codes': status_distribution,
                'top_ips': top_ips,
                'top_paths': top_paths
            },
            'trends': {
                'error_rate_by_date': {str(k): v for k, v in error_rate_trend.items()}
            },
            'security_status': {
                'models_trained': self.is_trained,
                'baseline_established': bool(self.baseline_metrics),
                'last_updated': datetime.now().isoformat()
            }
        }

# Global anomaly detector instance
anomaly_detector = AnomalyDetector() 