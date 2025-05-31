import asyncio
import logging
from typing import List, Dict, Any, Optional, Tuple
import numpy as np
import pandas as pd
from datetime import datetime, timedelta
from collections import defaultdict
import math

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class SmartDiscountEngine:
    def __init__(self):
        self.is_trained = False
        self.product_performance = {}
        self.user_behavior_patterns = {}
        self.seasonal_trends = {}
        self.price_elasticity = {}
        
    async def initialize(self):
        """Initialize the smart discount engine"""
        try:
            await self._load_sales_data()
            await self._load_user_behavior_data()
            await self._analyze_product_performance()
            await self._calculate_price_elasticity()
            await self._analyze_seasonal_trends()
            self.is_trained = True
            logger.info("Smart discount engine initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize smart discount engine: {e}")
            raise

    async def _load_sales_data(self):
        """Load sales and order data"""
        conn = await get_db_connection()
        try:
            query = """
            SELECT 
                oi.product_id,
                oi.quantity,
                oi.price,
                o.created_at,
                p.name as product_name,
                p.category,
                p.price as current_price,
                COALESCE(d.discount_value, 0) as discount_applied,
                COALESCE(d.discount_type, 'none') as discount_type
            FROM order_items oi
            JOIN orders o ON oi.order_id = o.id
            JOIN products p ON oi.product_id = p.id
            LEFT JOIN discounts d ON (d.product_id = p.id OR d.category = p.category)
                AND d.is_active = true
                AND o.created_at BETWEEN d.start_date AND d.end_date
            WHERE o.created_at >= $1
            ORDER BY o.created_at DESC
            """
            
            cutoff_date = datetime.now() - timedelta(days=90)
            rows = await conn.fetch(query, cutoff_date)
            
            self.sales_data = pd.DataFrame([dict(row) for row in rows])
            
            logger.info(f"Loaded {len(self.sales_data)} sales records")
        finally:
            await release_db_connection(conn)

    async def _load_user_behavior_data(self):
        """Load user interaction and behavior data"""
        conn = await get_db_connection()
        try:
            # Load user interactions
            interaction_query = """
            SELECT 
                ui.user_id,
                ui.product_id,
                ui.interaction_type,
                ui.created_at,
                p.category,
                p.price
            FROM user_interactions ui
            JOIN products p ON ui.product_id = p.id
            WHERE ui.created_at >= $1
            """
            
            # Load cart abandonment data - simplified query to avoid complex joins
            cart_query = """
            SELECT 
                c.user_id,
                ci.product_id,
                ci.quantity,
                ci.created_at,
                p.price,
                p.category,
                CASE 
                    WHEN EXISTS (
                        SELECT 1 FROM orders o 
                        JOIN order_items oi ON o.id = oi.order_id 
                        WHERE o.user_id = c.user_id 
                        AND oi.product_id = ci.product_id 
                        AND o.created_at > ci.created_at
                    ) THEN false 
                    ELSE true 
                END as abandoned
            FROM cart_items ci
            JOIN shopping_carts c ON ci.cart_id = c.id
            JOIN products p ON ci.product_id = p.id
            WHERE ci.created_at >= $1
            """
            
            cutoff_date = datetime.now() - timedelta(days=60)
            
            interaction_rows = await conn.fetch(interaction_query, cutoff_date)
            cart_rows = await conn.fetch(cart_query, cutoff_date)
            
            self.interaction_data = pd.DataFrame([dict(row) for row in interaction_rows])
            self.cart_data = pd.DataFrame([dict(row) for row in cart_rows])
            
            logger.info(f"Loaded {len(self.interaction_data)} interactions and {len(self.cart_data)} cart records")
        finally:
            await release_db_connection(conn)

    async def _analyze_product_performance(self):
        """Analyze product performance metrics"""
        if self.sales_data.empty:
            return
            
        # Group by product
        product_stats = self.sales_data.groupby('product_id').agg({
            'quantity': ['sum', 'count'],
            'price': 'mean',
            'current_price': 'first',
            'product_name': 'first',
            'category': 'first',
            'created_at': ['min', 'max']
        }).reset_index()
        
        # Flatten column names
        product_stats.columns = [
            'product_id', 'total_quantity', 'order_count', 'avg_sale_price',
            'current_price', 'product_name', 'category', 'first_sale', 'last_sale'
        ]
        
        # Calculate performance metrics
        for _, product in product_stats.iterrows():
            product_id = product['product_id']
            
            # Calculate sales velocity (units per day)
            days_active = max((product['last_sale'] - product['first_sale']).days, 1)
            sales_velocity = product['total_quantity'] / days_active
            
            # Calculate conversion rate from interactions
            if not self.interaction_data.empty:
                product_interactions = self.interaction_data[
                    self.interaction_data['product_id'] == product_id
                ]
                view_count = len(product_interactions[product_interactions['interaction_type'] == 'view'])
                conversion_rate = product['order_count'] / max(view_count, 1)
            else:
                conversion_rate = 0.1  # Default assumption
            
            # Calculate inventory turnover (assuming monthly restocking)
            inventory_turnover = product['total_quantity'] / 30  # Simplified
            
            self.product_performance[str(product_id)] = {
                'product_name': product['product_name'],
                'category': product['category'],
                'total_sales': product['total_quantity'],
                'order_count': product['order_count'],
                'sales_velocity': sales_velocity,
                'conversion_rate': conversion_rate,
                'inventory_turnover': inventory_turnover,
                'avg_sale_price': product['avg_sale_price'],
                'current_price': product['current_price'],
                'performance_score': self._calculate_performance_score(
                    sales_velocity, conversion_rate, inventory_turnover
                )
            }

    def _calculate_performance_score(self, velocity: float, conversion: float, turnover: float) -> float:
        """Calculate overall performance score"""
        # Normalize and weight the metrics
        velocity_score = min(velocity / 10, 1.0) * 0.4  # Weight: 40%
        conversion_score = min(conversion, 1.0) * 0.4   # Weight: 40%
        turnover_score = min(turnover / 5, 1.0) * 0.2   # Weight: 20%
        
        return velocity_score + conversion_score + turnover_score

    async def _calculate_price_elasticity(self):
        """Calculate price elasticity for products"""
        if self.sales_data.empty:
            return
            
        for product_id in self.sales_data['product_id'].unique():
            product_sales = self.sales_data[self.sales_data['product_id'] == product_id]
            
            if len(product_sales) < 5:  # Need minimum data points
                continue
                
            # Group by price ranges to analyze elasticity
            price_ranges = pd.cut(product_sales['price'], bins=5, duplicates='drop')
            price_analysis = product_sales.groupby(price_ranges).agg({
                'quantity': 'sum',
                'price': 'mean'
            }).reset_index()
            
            if len(price_analysis) >= 2:
                # Calculate elasticity using simple regression
                prices = price_analysis['price'].values
                quantities = price_analysis['quantity'].values
                
                # Calculate percentage changes
                price_changes = np.diff(prices) / prices[:-1]
                quantity_changes = np.diff(quantities) / quantities[:-1]
                
                # Avoid division by zero
                valid_indices = price_changes != 0
                if np.any(valid_indices):
                    elasticity = np.mean(quantity_changes[valid_indices] / price_changes[valid_indices])
                    self.price_elasticity[str(product_id)] = abs(elasticity)
                else:
                    self.price_elasticity[str(product_id)] = 1.0  # Default elasticity
            else:
                self.price_elasticity[str(product_id)] = 1.0

    async def _analyze_seasonal_trends(self):
        """Analyze seasonal and temporal trends"""
        if self.sales_data.empty:
            return
            
        # Add time features
        self.sales_data['month'] = self.sales_data['created_at'].dt.month
        self.sales_data['day_of_week'] = self.sales_data['created_at'].dt.dayofweek
        self.sales_data['hour'] = self.sales_data['created_at'].dt.hour
        
        # Analyze by category
        for category in self.sales_data['category'].unique():
            category_sales = self.sales_data[self.sales_data['category'] == category]
            
            # Monthly trends
            monthly_sales = category_sales.groupby('month')['quantity'].sum()
            peak_months = monthly_sales.nlargest(3).index.tolist()
            low_months = monthly_sales.nsmallest(3).index.tolist()
            
            # Day of week trends
            daily_sales = category_sales.groupby('day_of_week')['quantity'].sum()
            peak_days = daily_sales.nlargest(2).index.tolist()
            
            # Hour trends
            hourly_sales = category_sales.groupby('hour')['quantity'].sum()
            peak_hours = hourly_sales.nlargest(3).index.tolist()
            
            self.seasonal_trends[category] = {
                'peak_months': peak_months,
                'low_months': low_months,
                'peak_days': peak_days,
                'peak_hours': peak_hours,
                'monthly_pattern': monthly_sales.to_dict(),
                'seasonal_factor': monthly_sales.std() / monthly_sales.mean()
            }

    async def suggest_product_discount(self, product_id: str) -> Dict[str, Any]:
        """Suggest optimal discount for a specific product"""
        if not self.is_trained:
            await self.initialize()
            
        # Check if product exists in performance data
        if product_id not in self.product_performance:
            return {
                'product_id': product_id,
                'error': 'Product not found in performance data',
                'recommendation': 'no_data',
                'suggested_discount_percentage': 0,
                'reasoning': 'Insufficient data to make discount recommendation'
            }
            
        performance = self.product_performance[product_id]
        category = performance['category']
        
        # Get performance metrics
        current_performance = performance['performance_score']
        conversion_rate = performance['conversion_rate']
        sales_velocity = performance['sales_velocity']
        
        # Get price elasticity
        elasticity = self.price_elasticity.get(product_id, 1.0)
        
        # Calculate base discount recommendation
        discount_recommendation = self._calculate_optimal_discount(
            current_performance, conversion_rate, sales_velocity, elasticity
        )
        
        # Apply seasonal adjustments
        seasonal_adjustment = self._get_seasonal_adjustment(category)
        final_discount = discount_recommendation['discount_percentage'] * seasonal_adjustment
        
        # Cap discount at reasonable limits
        final_discount = min(final_discount, 50)  # Max 50% discount
        
        # Determine discount type
        discount_type = self._determine_discount_type(final_discount, performance)
        
        return {
            'product_id': product_id,
            'product_name': performance['product_name'],
            'category': category,
            'current_performance_score': current_performance,
            'recommendation': discount_recommendation['strategy'],
            'suggested_discount_percentage': round(final_discount, 1),
            'discount_type': discount_type,
            'expected_impact': self._calculate_expected_impact(final_discount, elasticity, performance),
            'reasoning': discount_recommendation['reasoning'],
            'seasonal_factor': seasonal_adjustment,
            'urgency': self._calculate_urgency(performance),
            'generated_at': datetime.now().isoformat()
        }

    def _calculate_optimal_discount(self, performance: float, conversion: float, 
                                  velocity: float, elasticity: float) -> Dict[str, Any]:
        """Calculate optimal discount based on performance metrics"""
        
        if performance >= 0.7:  # High performing product
            if conversion < 0.1:  # Low conversion despite good performance
                return {
                    'strategy': 'conversion_boost',
                    'discount_percentage': 5 + (0.1 - conversion) * 100,
                    'reasoning': 'High performance but low conversion rate - small discount to boost conversions'
                }
            else:
                return {
                    'strategy': 'maintain_momentum',
                    'discount_percentage': 0,
                    'reasoning': 'Product performing well - no discount needed'
                }
        
        elif performance >= 0.4:  # Medium performing product
            if velocity < 1.0:  # Slow moving
                discount = 10 + (1.0 - velocity) * 10
                return {
                    'strategy': 'velocity_increase',
                    'discount_percentage': min(discount, 25),
                    'reasoning': 'Medium performance with slow sales - moderate discount to increase velocity'
                }
            else:
                return {
                    'strategy': 'steady_promotion',
                    'discount_percentage': 8,
                    'reasoning': 'Steady performance - small promotional discount'
                }
        
        else:  # Low performing product
            if elasticity > 1.5:  # Price sensitive
                discount = 20 + (0.4 - performance) * 50
                return {
                    'strategy': 'clearance_boost',
                    'discount_percentage': min(discount, 40),
                    'reasoning': 'Low performance with high price sensitivity - significant discount needed'
                }
            else:
                return {
                    'strategy': 'inventory_clearance',
                    'discount_percentage': 15,
                    'reasoning': 'Low performance - moderate discount to clear inventory'
                }

    def _get_seasonal_adjustment(self, category: str) -> float:
        """Get seasonal adjustment factor"""
        if category not in self.seasonal_trends:
            return 1.0
            
        current_month = datetime.now().month
        trends = self.seasonal_trends[category]
        
        if current_month in trends['low_months']:
            return 1.3  # Increase discount during low season
        elif current_month in trends['peak_months']:
            return 0.7  # Reduce discount during peak season
        else:
            return 1.0  # Normal season

    def _determine_discount_type(self, discount_percentage: float, performance: Dict) -> str:
        """Determine the type of discount to apply"""
        if discount_percentage == 0:
            return 'none'
        elif discount_percentage <= 5:
            return 'percentage'
        elif discount_percentage <= 15:
            return 'percentage'
        elif performance['current_price'] > 100:
            return 'fixed_amount'  # Fixed amount works better for expensive items
        else:
            return 'percentage'

    def _calculate_expected_impact(self, discount: float, elasticity: float, 
                                 performance: Dict) -> Dict[str, Any]:
        """Calculate expected impact of discount"""
        if discount == 0:
            return {
                'sales_increase_percentage': 0,
                'revenue_impact_percentage': 0,
                'units_increase': 0
            }
        
        # Calculate expected sales increase based on elasticity
        price_reduction = discount / 100
        sales_increase = elasticity * price_reduction
        
        # Calculate revenue impact
        revenue_impact = sales_increase - price_reduction
        
        # Calculate expected unit increase
        current_velocity = performance['sales_velocity']
        units_increase = current_velocity * sales_increase
        
        return {
            'sales_increase_percentage': round(sales_increase * 100, 1),
            'revenue_impact_percentage': round(revenue_impact * 100, 1),
            'units_increase': round(units_increase, 1)
        }

    def _calculate_urgency(self, performance: Dict) -> str:
        """Calculate urgency level for discount application"""
        score = performance['performance_score']
        velocity = performance['sales_velocity']
        
        if score < 0.2 and velocity < 0.5:
            return 'high'
        elif score < 0.4 or velocity < 1.0:
            return 'medium'
        else:
            return 'low'

    async def suggest_category_discounts(self, category: str) -> Dict[str, Any]:
        """Suggest discounts for an entire category"""
        if not self.is_trained:
            await self.initialize()
            
        # Check if we have any product performance data
        if not self.product_performance:
            return {
                'category': category,
                'error': 'No product performance data available',
                'recommendations': [],
                'category_strategy': 'no_data'
            }
            
        category_products = [
            (pid, perf) for pid, perf in self.product_performance.items()
            if perf['category'].lower() == category.lower()
        ]
        
        if not category_products:
            return {
                'category': category,
                'error': 'No products found in this category',
                'recommendations': [],
                'category_strategy': 'no_data'
            }
        
        recommendations = []
        for product_id, _ in category_products:
            suggestion = await self.suggest_product_discount(product_id)
            recommendations.append(suggestion)
        
        # Analyze category-wide strategy
        valid_recommendations = [r for r in recommendations if 'error' not in r]
        
        if not valid_recommendations:
            return {
                'category': category,
                'error': 'Unable to generate recommendations for any products in category',
                'recommendations': recommendations,
                'category_strategy': 'no_data'
            }
        
        avg_performance = np.mean([r['current_performance_score'] for r in valid_recommendations])
        high_urgency_count = len([r for r in valid_recommendations if r['urgency'] == 'high'])
        
        if avg_performance < 0.3:
            category_strategy = 'category_wide_sale'
        elif high_urgency_count > len(valid_recommendations) * 0.3:
            category_strategy = 'selective_discounts'
        else:
            category_strategy = 'maintain_current'
        
        return {
            'category': category,
            'total_products': len(recommendations),
            'valid_recommendations': len(valid_recommendations),
            'average_performance': round(avg_performance, 2),
            'high_urgency_products': high_urgency_count,
            'category_strategy': category_strategy,
            'recommendations': recommendations,
            'seasonal_trends': self.seasonal_trends.get(category, {}),
            'generated_at': datetime.now().isoformat()
        }

    async def get_discount_insights(self) -> Dict[str, Any]:
        """Get overall discount insights and recommendations"""
        if not self.is_trained:
            await self.initialize()
            
        # Check if we have any product performance data
        if not self.product_performance:
            return {
                'insights': [],
                'summary': {
                    'total_products_analyzed': 0,
                    'low_performing_products': 0,
                    'high_performing_products': 0,
                    'average_performance': 0.0,
                    'discount_opportunities': 0
                },
                'generated_at': datetime.now().isoformat(),
                'error': 'No product performance data available'
            }
            
        insights = []
        
        # Analyze overall performance
        all_scores = [perf['performance_score'] for perf in self.product_performance.values()]
        avg_performance = np.mean(all_scores)
        
        low_performers = [
            pid for pid, perf in self.product_performance.items()
            if perf['performance_score'] < 0.3
        ]
        
        high_performers = [
            pid for pid, perf in self.product_performance.items()
            if perf['performance_score'] > 0.7
        ]
        
        # Generate insights
        if len(low_performers) > len(self.product_performance) * 0.2:
            insights.append({
                'type': 'warning',
                'title': 'High Number of Low-Performing Products',
                'description': f'{len(low_performers)} products need attention',
                'action': 'Consider implementing targeted discount campaigns',
                'affected_products': len(low_performers)
            })
        
        if avg_performance > 0.6:
            insights.append({
                'type': 'success',
                'title': 'Strong Overall Performance',
                'description': 'Most products are performing well',
                'action': 'Focus on maintaining momentum with selective promotions',
                'performance_score': round(avg_performance, 2)
            })
        
        # Seasonal recommendations
        current_month = datetime.now().month
        seasonal_recommendations = []
        
        for category, trends in self.seasonal_trends.items():
            if current_month in trends['low_months']:
                seasonal_recommendations.append({
                    'category': category,
                    'recommendation': 'increase_discounts',
                    'reason': 'Low season for this category'
                })
            elif current_month in trends['peak_months']:
                seasonal_recommendations.append({
                    'category': category,
                    'recommendation': 'reduce_discounts',
                    'reason': 'Peak season for this category'
                })
        
        return {
            'insights': insights,
            'overall_performance': {
                'average_score': round(avg_performance, 2),
                'total_products': len(self.product_performance),
                'low_performers': len(low_performers),
                'high_performers': len(high_performers)
            },
            'seasonal_recommendations': seasonal_recommendations,
            'discount_opportunities': len(low_performers),
            'generated_at': datetime.now().isoformat()
        }

# Global smart discount engine instance
smart_discount_engine = SmartDiscountEngine() 