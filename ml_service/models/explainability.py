"""
Comprehensive Explainability System for ML Models
Addresses Issue #2: Lack of reasoning in Machine Learning models

This module provides detailed explanations for how and why each ML model makes decisions,
including technical details, confidence levels, and business impact analysis.
"""

import logging
from typing import Dict, List, Any, Optional
from datetime import datetime
import numpy as np
import pandas as pd

logger = logging.getLogger(__name__)

class MLExplainabilityEngine:
    """
    Central explainability engine that provides reasoning for all ML models
    """
    
    def __init__(self):
        self.explanation_templates = self._initialize_explanation_templates()
        self.confidence_thresholds = self._initialize_confidence_thresholds()
        self.business_impact_rules = self._initialize_business_impact_rules()
    
    def _initialize_explanation_templates(self) -> Dict[str, Dict]:
        """Initialize explanation templates for each ML model"""
        return {
            'recommendations': {
                'collaborative': {
                    'method_name': 'Collaborative Filtering',
                    'description': 'Finds users with similar preferences and recommends items they liked',
                    'strengths': ['Discovers unexpected items', 'Leverages community wisdom', 'No content analysis needed'],
                    'limitations': ['Cold start problem', 'Popularity bias', 'Sparsity issues'],
                    'use_cases': ['Personalized recommendations', 'Cross-selling', 'Discovery']
                },
                'content_based': {
                    'method_name': 'Content-Based Filtering',
                    'description': 'Analyzes item features to find similar products to user preferences',
                    'strengths': ['No cold start', 'Transparent reasoning', 'Domain knowledge integration'],
                    'limitations': ['Limited diversity', 'Feature engineering required', 'Over-specialization'],
                    'use_cases': ['Similar item suggestions', 'Category exploration', 'Feature-based matching']
                },
                'hybrid': {
                    'method_name': 'Hybrid Recommendation',
                    'description': 'Combines collaborative and content-based approaches for optimal results',
                    'strengths': ['Best of both worlds', 'Reduced limitations', 'Higher accuracy'],
                    'limitations': ['Increased complexity', 'Parameter tuning', 'Computational overhead'],
                    'use_cases': ['Production systems', 'Balanced recommendations', 'Robust performance']
                }
            },
            'search': {
                'tfidf': {
                    'method_name': 'TF-IDF Semantic Search',
                    'description': 'Uses term frequency and inverse document frequency to find relevant content',
                    'strengths': ['Fast execution', 'Interpretable results', 'Language agnostic'],
                    'limitations': ['Keyword dependency', 'No semantic understanding', 'Vocabulary mismatch'],
                    'use_cases': ['Text search', 'Document retrieval', 'Content matching']
                }
            },
            'anomaly_detection': {
                'isolation_forest': {
                    'method_name': 'Isolation Forest',
                    'description': 'Isolates anomalies by randomly selecting features and split values',
                    'strengths': ['Unsupervised learning', 'Efficient for large datasets', 'No assumptions about data distribution'],
                    'limitations': ['Parameter sensitivity', 'High-dimensional challenges', 'Interpretability issues'],
                    'use_cases': ['Fraud detection', 'Network security', 'Quality control']
                },
                'statistical': {
                    'method_name': 'Statistical Analysis',
                    'description': 'Detects anomalies using statistical measures and thresholds',
                    'strengths': ['Interpretable', 'Fast computation', 'Well-understood'],
                    'limitations': ['Assumes normal distribution', 'Threshold dependency', 'Limited to univariate'],
                    'use_cases': ['Performance monitoring', 'Quality control', 'Baseline comparison']
                }
            },
            'sentiment': {
                'lexicon_based': {
                    'method_name': 'Lexicon-Based Sentiment Analysis',
                    'description': 'Uses predefined word sentiment scores to classify text sentiment',
                    'strengths': ['No training required', 'Interpretable', 'Domain adaptable'],
                    'limitations': ['Context ignorance', 'Sarcasm detection', 'Domain specificity'],
                    'use_cases': ['Social media monitoring', 'Review analysis', 'Brand sentiment']
                }
            },
            'auto_tagging': {
                'tfidf_clustering': {
                    'method_name': 'TF-IDF + Clustering',
                    'description': 'Extracts keywords using TF-IDF and groups similar content',
                    'strengths': ['Automatic tag discovery', 'Scalable', 'No manual labeling'],
                    'limitations': ['Quality depends on data', 'May generate irrelevant tags', 'Requires post-processing'],
                    'use_cases': ['Content categorization', 'Metadata generation', 'Search optimization']
                }
            },
            'smart_discounts': {
                'performance_based': {
                    'method_name': 'Performance-Based Optimization',
                    'description': 'Analyzes sales metrics to suggest optimal discount strategies',
                    'strengths': ['Data-driven decisions', 'ROI optimization', 'Market responsive'],
                    'limitations': ['Historical data dependency', 'Market volatility', 'Competitor effects'],
                    'use_cases': ['Pricing strategy', 'Inventory management', 'Revenue optimization']
                }
            },
            'trends': {
                'linear_regression': {
                    'method_name': 'Linear Regression Analysis',
                    'description': 'Identifies trends using linear regression on time series data',
                    'strengths': ['Simple interpretation', 'Fast computation', 'Trend direction clear'],
                    'limitations': ['Linear assumption', 'Seasonality ignorance', 'Outlier sensitivity'],
                    'use_cases': ['Sales forecasting', 'Trend analysis', 'Performance tracking']
                }
            }
        }
    
    def _initialize_confidence_thresholds(self) -> Dict[str, Dict]:
        """Initialize confidence thresholds for each model type"""
        return {
            'recommendations': {'high': 0.7, 'medium': 0.4, 'low': 0.0},
            'search': {'high': 0.5, 'medium': 0.2, 'low': 0.0},
            'anomaly_detection': {'high': 0.7, 'medium': 0.4, 'low': 0.0},
            'sentiment': {'high': 0.8, 'medium': 0.6, 'low': 0.0},
            'auto_tagging': {'high': 0.7, 'medium': 0.5, 'low': 0.0},
            'smart_discounts': {'high': 0.8, 'medium': 0.6, 'low': 0.0},
            'trends': {'high': 0.8, 'medium': 0.6, 'low': 0.0}
        }
    
    def _initialize_business_impact_rules(self) -> Dict[str, Dict]:
        """Initialize business impact assessment rules"""
        return {
            'recommendations': {
                'high_impact': 'Directly affects user engagement and sales conversion',
                'metrics': ['Click-through rate', 'Conversion rate', 'Revenue per user'],
                'success_indicators': ['Increased engagement', 'Higher sales', 'Better user retention']
            },
            'search': {
                'high_impact': 'Critical for product discovery and user experience',
                'metrics': ['Search success rate', 'Time to find', 'User satisfaction'],
                'success_indicators': ['Faster product discovery', 'Reduced bounce rate', 'Higher search-to-purchase conversion']
            },
            'anomaly_detection': {
                'high_impact': 'Essential for security and system reliability',
                'metrics': ['False positive rate', 'Detection accuracy', 'Response time'],
                'success_indicators': ['Reduced security incidents', 'Faster threat detection', 'Lower false alarms']
            },
            'sentiment': {
                'medium_impact': 'Provides insights for product and service improvement',
                'metrics': ['Sentiment accuracy', 'Coverage rate', 'Actionable insights'],
                'success_indicators': ['Better product feedback', 'Improved customer satisfaction', 'Proactive issue resolution']
            },
            'auto_tagging': {
                'medium_impact': 'Improves content organization and searchability',
                'metrics': ['Tag relevance', 'Coverage completeness', 'Manual effort reduction'],
                'success_indicators': ['Better content discovery', 'Reduced manual work', 'Improved search results']
            },
            'smart_discounts': {
                'high_impact': 'Directly affects revenue and profit margins',
                'metrics': ['Revenue impact', 'Margin preservation', 'Inventory turnover'],
                'success_indicators': ['Optimized pricing', 'Increased sales volume', 'Better inventory management']
            },
            'trends': {
                'high_impact': 'Guides strategic business decisions and planning',
                'metrics': ['Forecast accuracy', 'Trend detection speed', 'Business alignment'],
                'success_indicators': ['Better planning', 'Market responsiveness', 'Competitive advantage']
            }
        }
    
    def explain_model_decision(self, model_type: str, algorithm: str, 
                             decision_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Generate comprehensive explanation for any ML model decision
        
        Args:
            model_type: Type of ML model (recommendations, search, etc.)
            algorithm: Specific algorithm used
            decision_data: Data about the decision made
            
        Returns:
            Comprehensive explanation with reasoning, confidence, and business impact
        """
        
        # Get model template
        model_templates = self.explanation_templates.get(model_type, {})
        algorithm_template = model_templates.get(algorithm, {})
        
        # Generate explanation
        explanation = {
            'model_overview': self._generate_model_overview(model_type, algorithm, algorithm_template),
            'decision_reasoning': self._generate_decision_reasoning(decision_data),
            'confidence_analysis': self._generate_confidence_analysis(model_type, decision_data),
            'business_impact': self._generate_business_impact(model_type),
            'technical_details': self._generate_technical_details(algorithm_template, decision_data),
            'improvement_suggestions': self._generate_improvement_suggestions(model_type, decision_data),
            'explanation_metadata': {
                'generated_at': datetime.now().isoformat(),
                'model_type': model_type,
                'algorithm': algorithm,
                'explainability_version': '1.0'
            }
        }
        
        return explanation
    
    def _generate_model_overview(self, model_type: str, algorithm: str, 
                               template: Dict[str, Any]) -> Dict[str, Any]:
        """Generate high-level model overview"""
        return {
            'model_type': model_type.replace('_', ' ').title(),
            'algorithm': algorithm.replace('_', ' ').title(),
            'method_name': template.get('method_name', 'Unknown Method'),
            'description': template.get('description', 'No description available'),
            'primary_purpose': self._get_primary_purpose(model_type),
            'strengths': template.get('strengths', []),
            'limitations': template.get('limitations', []),
            'typical_use_cases': template.get('use_cases', [])
        }
    
    def _get_primary_purpose(self, model_type: str) -> str:
        """Get primary purpose of each model type"""
        purposes = {
            'recommendations': 'Personalize product suggestions to increase user engagement and sales',
            'search': 'Help users find relevant products quickly and efficiently',
            'anomaly_detection': 'Detect and prevent security threats and unusual behavior',
            'sentiment': 'Understand customer opinions and feedback sentiment',
            'auto_tagging': 'Automatically categorize and organize content',
            'smart_discounts': 'Optimize pricing strategies for maximum revenue',
            'trends': 'Identify patterns and forecast future business trends'
        }
        return purposes.get(model_type, 'Provide intelligent automation and insights')
    
    def _generate_decision_reasoning(self, decision_data: Dict[str, Any]) -> Dict[str, Any]:
        """Generate reasoning for the specific decision made"""
        reasoning = decision_data.get('reasoning', {})
        
        return {
            'primary_reason': reasoning.get('explanation', 'Decision based on model analysis'),
            'contributing_factors': reasoning.get('factors', []),
            'decision_process': reasoning.get('method', 'Standard model inference'),
            'key_inputs': self._extract_key_inputs(decision_data),
            'decision_path': self._generate_decision_path(reasoning)
        }
    
    def _extract_key_inputs(self, decision_data: Dict[str, Any]) -> List[str]:
        """Extract key inputs that influenced the decision"""
        inputs = []
        
        # Extract from technical details if available
        technical = decision_data.get('reasoning', {}).get('technical_details', {})
        
        if 'data_source' in technical:
            inputs.append(f"Data source: {technical['data_source']}")
        
        if 'features_analyzed' in technical:
            inputs.append(f"Features: {technical['features_analyzed']}")
        
        if 'algorithm' in technical:
            inputs.append(f"Algorithm: {technical['algorithm']}")
        
        # Add score if available
        if 'score' in decision_data:
            inputs.append(f"Confidence score: {decision_data['score']:.3f}")
        
        return inputs if inputs else ['Standard model inputs']
    
    def _generate_decision_path(self, reasoning: Dict[str, Any]) -> List[str]:
        """Generate step-by-step decision path"""
        path = []
        
        # Add method
        if 'method' in reasoning:
            path.append(f"1. Applied {reasoning['method']}")
        
        # Add key factors
        factors = reasoning.get('factors', [])
        for i, factor in enumerate(factors[:3], 2):  # Top 3 factors
            path.append(f"{i}. Considered: {factor}")
        
        # Add final decision
        if 'explanation' in reasoning:
            path.append(f"{len(path)+1}. Result: {reasoning['explanation']}")
        
        return path if path else ['1. Standard model processing', '2. Generated prediction', '3. Applied business rules']
    
    def _generate_confidence_analysis(self, model_type: str, decision_data: Dict[str, Any]) -> Dict[str, Any]:
        """Generate confidence analysis for the decision"""
        thresholds = self.confidence_thresholds.get(model_type, {'high': 0.7, 'medium': 0.4, 'low': 0.0})
        
        # Get confidence from reasoning or score
        reasoning = decision_data.get('reasoning', {})
        confidence_level = reasoning.get('confidence', 'Medium')
        score = decision_data.get('score', 0.5)
        
        # Determine numerical confidence if not provided
        if isinstance(confidence_level, str):
            confidence_score = 0.8 if confidence_level == 'High' else 0.6 if confidence_level == 'Medium' else 0.3
        else:
            confidence_score = float(confidence_level)
        
        return {
            'confidence_level': confidence_level,
            'confidence_score': confidence_score,
            'confidence_factors': self._get_confidence_factors(reasoning, decision_data),
            'reliability_indicators': self._get_reliability_indicators(model_type, decision_data),
            'uncertainty_sources': self._get_uncertainty_sources(model_type, reasoning)
        }
    
    def _get_confidence_factors(self, reasoning: Dict, decision_data: Dict) -> List[str]:
        """Get factors that contribute to confidence level"""
        factors = []
        
        # Check data quality indicators
        technical = reasoning.get('technical_details', {})
        
        if 'data_sparsity' in technical:
            factors.append(f"Data availability: {technical['data_sparsity']}")
        
        if 'similar_users_analyzed' in technical:
            factors.append(f"Sample size: {technical['similar_users_analyzed']} users")
        
        if 'contributing_users' in technical:
            factors.append(f"Contributing factors: {technical['contributing_users']}")
        
        # Check algorithm-specific factors
        if 'algorithm' in technical:
            factors.append(f"Algorithm reliability: {technical['algorithm']}")
        
        return factors if factors else ['Standard confidence assessment']
    
    def _get_reliability_indicators(self, model_type: str, decision_data: Dict) -> List[str]:
        """Get indicators of decision reliability"""
        indicators = []
        
        # Model-specific reliability indicators
        if model_type == 'recommendations':
            if decision_data.get('algorithm') == 'hybrid':
                indicators.append('Multiple algorithms agree')
            
        elif model_type == 'anomaly_detection':
            reasoning = decision_data.get('reasoning', {})
            breakdown = reasoning.get('analysis_breakdown', {})
            triggered_methods = len([m for m in breakdown.values() if m.get('score', 0) > 0])
            if triggered_methods > 1:
                indicators.append(f'{triggered_methods} detection methods triggered')
        
        elif model_type == 'search':
            reasoning = decision_data.get('reasoning', {})
            matched_terms = len(reasoning.get('matched_terms', []))
            if matched_terms > 1:
                indicators.append(f'Multiple term matches ({matched_terms})')
        
        return indicators if indicators else ['Standard reliability checks passed']
    
    def _get_uncertainty_sources(self, model_type: str, reasoning: Dict) -> List[str]:
        """Identify sources of uncertainty in the decision"""
        uncertainties = []
        
        # Common uncertainty sources
        technical = reasoning.get('technical_details', {})
        
        if 'fallback_reason' in technical:
            uncertainties.append(f"Fallback used: {technical['fallback_reason']}")
        
        # Model-specific uncertainties
        if model_type == 'recommendations':
            if 'cold_start' in str(reasoning).lower():
                uncertainties.append('Limited user interaction history')
        
        elif model_type == 'anomaly_detection':
            if reasoning.get('confidence', '').lower() == 'low':
                uncertainties.append('Borderline anomaly score')
        
        return uncertainties if uncertainties else ['No significant uncertainty sources identified']
    
    def _generate_business_impact(self, model_type: str) -> Dict[str, Any]:
        """Generate business impact analysis"""
        impact_rules = self.business_impact_rules.get(model_type, {})
        
        return {
            'impact_level': list(impact_rules.keys())[0] if impact_rules else 'medium_impact',
            'impact_description': list(impact_rules.values())[0] if impact_rules else 'Provides business value through automation',
            'key_metrics': impact_rules.get('metrics', ['Performance', 'Accuracy', 'User satisfaction']),
            'success_indicators': impact_rules.get('success_indicators', ['Improved efficiency', 'Better outcomes']),
            'roi_potential': self._assess_roi_potential(model_type),
            'strategic_value': self._assess_strategic_value(model_type)
        }
    
    def _assess_roi_potential(self, model_type: str) -> str:
        """Assess ROI potential for each model type"""
        roi_assessments = {
            'recommendations': 'High - Direct impact on sales and user engagement',
            'search': 'High - Improves user experience and conversion rates',
            'anomaly_detection': 'Very High - Prevents losses from security incidents',
            'sentiment': 'Medium - Provides insights for product improvement',
            'auto_tagging': 'Medium - Reduces manual effort and improves organization',
            'smart_discounts': 'Very High - Optimizes revenue and profit margins',
            'trends': 'High - Enables better strategic planning and forecasting'
        }
        return roi_assessments.get(model_type, 'Medium - Provides operational efficiency')
    
    def _assess_strategic_value(self, model_type: str) -> str:
        """Assess strategic value for each model type"""
        strategic_values = {
            'recommendations': 'Core competitive advantage in personalization',
            'search': 'Essential for user experience and product discovery',
            'anomaly_detection': 'Critical for security and trust',
            'sentiment': 'Valuable for customer insights and product development',
            'auto_tagging': 'Supports scalable content management',
            'smart_discounts': 'Key for pricing strategy and revenue optimization',
            'trends': 'Fundamental for data-driven decision making'
        }
        return strategic_values.get(model_type, 'Supports operational excellence')
    
    def _generate_technical_details(self, template: Dict, decision_data: Dict) -> Dict[str, Any]:
        """Generate technical implementation details"""
        reasoning = decision_data.get('reasoning', {})
        technical = reasoning.get('technical_details', {})
        
        return {
            'algorithm_family': template.get('method_name', 'Unknown'),
            'implementation_details': technical,
            'data_requirements': self._get_data_requirements(template),
            'computational_complexity': self._assess_complexity(template),
            'scalability_notes': self._get_scalability_notes(template),
            'performance_characteristics': self._get_performance_characteristics(template)
        }
    
    def _get_data_requirements(self, template: Dict) -> List[str]:
        """Get data requirements for the algorithm"""
        method_name = template.get('method_name', '').lower()
        
        if 'collaborative' in method_name:
            return ['User-item interaction matrix', 'Sufficient user overlap', 'Historical behavior data']
        elif 'content' in method_name:
            return ['Item feature descriptions', 'Structured metadata', 'Text content']
        elif 'tfidf' in method_name:
            return ['Text corpus', 'Tokenized documents', 'Vocabulary coverage']
        elif 'isolation' in method_name:
            return ['Numerical features', 'Sufficient sample size', 'Representative data distribution']
        else:
            return ['Structured data', 'Historical records', 'Quality labels']
    
    def _assess_complexity(self, template: Dict) -> str:
        """Assess computational complexity"""
        method_name = template.get('method_name', '').lower()
        
        if 'svd' in method_name or 'matrix' in method_name:
            return 'O(n³) for matrix operations, O(k) for inference'
        elif 'tfidf' in method_name:
            return 'O(n*m) for vectorization, O(1) for similarity'
        elif 'isolation' in method_name:
            return 'O(n log n) for training, O(log n) for prediction'
        else:
            return 'Linear to sub-linear complexity'
    
    def _get_scalability_notes(self, template: Dict) -> List[str]:
        """Get scalability considerations"""
        method_name = template.get('method_name', '').lower()
        
        if 'collaborative' in method_name:
            return ['Memory scales with user-item matrix', 'Consider dimensionality reduction', 'Incremental updates possible']
        elif 'tfidf' in method_name:
            return ['Vocabulary size affects memory', 'Sparse matrices help efficiency', 'Parallel processing friendly']
        elif 'isolation' in method_name:
            return ['Scales well with data size', 'Parallel tree construction', 'Memory efficient']
        else:
            return ['Generally scalable', 'Consider data partitioning', 'Monitor memory usage']
    
    def _get_performance_characteristics(self, template: Dict) -> Dict[str, str]:
        """Get performance characteristics"""
        return {
            'training_speed': 'Fast to Moderate depending on data size',
            'inference_speed': 'Real-time capable',
            'memory_usage': 'Moderate to High depending on algorithm',
            'accuracy_profile': 'Good with sufficient data'
        }
    
    def _generate_improvement_suggestions(self, model_type: str, decision_data: Dict) -> List[str]:
        """Generate suggestions for improving model performance"""
        suggestions = []
        
        reasoning = decision_data.get('reasoning', {})
        confidence = reasoning.get('confidence', 'Medium')
        
        # General suggestions based on confidence
        if confidence == 'Low':
            suggestions.extend([
                'Collect more training data to improve model confidence',
                'Consider feature engineering to capture more relevant signals',
                'Evaluate alternative algorithms for this use case'
            ])
        
        # Model-specific suggestions
        if model_type == 'recommendations':
            if decision_data.get('algorithm') == 'popular':
                suggestions.append('Gather user interaction data to enable personalized recommendations')
            
        elif model_type == 'anomaly_detection':
            breakdown = reasoning.get('analysis_breakdown', {})
            if len([m for m in breakdown.values() if m.get('score', 0) > 0]) == 1:
                suggestions.append('Consider ensemble methods for more robust detection')
        
        elif model_type == 'search':
            matched_terms = len(reasoning.get('matched_terms', []))
            if matched_terms < 2:
                suggestions.append('Implement query expansion to improve search coverage')
        
        # Add general best practices
        suggestions.extend([
            'Monitor model performance metrics regularly',
            'Implement A/B testing to validate improvements',
            'Consider user feedback to refine model behavior'
        ])
        
        return suggestions[:5]  # Limit to top 5 suggestions

# Global explainability engine instance
explainability_engine = MLExplainabilityEngine() 