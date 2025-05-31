import asyncio
import logging
from typing import List, Dict, Any, Optional, Tuple
import numpy as np
import pandas as pd
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import mean_absolute_error, mean_squared_error
from datetime import datetime, timedelta
import json

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class TrendAnalyzer:
    def __init__(self):
        self.models = {}
        self.scalers = {}
        self.trend_data = {}
        self.is_trained = False
        
    async def initialize(self):
        try:
            await self._load_historical_data()
            await self._analyze_trends()
            self.is_trained = True
            logger.info("Trend analyzer initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize trend analyzer: {e}")
            raise

    async def _load_historical_data(self):
        conn = await get_db_connection()
        try:
            sales_query = """
            SELECT 
                p.id as product_id,
                p.name as product_name,
                p.category,
                p.price,
                DATE(o.created_at) as date,
                SUM(oi.quantity) as units_sold,
                SUM(oi.quantity * oi.price) as revenue
            FROM orders o
            JOIN order_items oi ON o.id = oi.order_id
            JOIN products p ON oi.product_id = p.id
            WHERE o.status IN ('delivered', 'completed')
            AND o.created_at >= $1
            GROUP BY p.id, p.name, p.category, p.price, DATE(o.created_at)
            ORDER BY date
            """
            
            cutoff_date = datetime.now() - timedelta(days=90)
            sales_rows = await conn.fetch(sales_query, cutoff_date)
            
            interaction_query = """
            SELECT 
                p.id as product_id,
                p.name as product_name,
                p.category,
                DATE(ui.created_at) as date,
                ui.interaction_type,
                COUNT(*) as interaction_count
            FROM user_interactions ui
            JOIN products p ON ui.product_id = p.id
            WHERE ui.created_at >= $1
            GROUP BY p.id, p.name, p.category, DATE(ui.created_at), ui.interaction_type
            ORDER BY date
            """
            
            interaction_rows = await conn.fetch(interaction_query, cutoff_date)
            
            search_query = """
            SELECT 
                DATE(created_at) as date,
                query,
                results_count,
                COUNT(*) as search_count
            FROM search_queries
            WHERE created_at >= $1
            GROUP BY DATE(created_at), query, results_count
            ORDER BY date
            """
            
            search_rows = await conn.fetch(search_query, cutoff_date)
            
            self.sales_data = pd.DataFrame([dict(row) for row in sales_rows])
            self.interaction_data = pd.DataFrame([dict(row) for row in interaction_rows])
            self.search_data = pd.DataFrame([dict(row) for row in search_rows])
            
            logger.info(f"Loaded {len(self.sales_data)} sales records, "
                       f"{len(self.interaction_data)} interaction records, "
                       f"{len(self.search_data)} search records")
        finally:
            await release_db_connection(conn)

    async def _analyze_trends(self):
        await self._analyze_product_trends()
        
        await self._analyze_category_trends()
        
        await self._analyze_search_trends()
        
        await self._analyze_seasonal_patterns()

    async def _analyze_product_trends(self):
        if self.sales_data.empty:
            return
            
        product_trends = {}
        
        for product_id in self.sales_data['product_id'].unique():
            product_data = self.sales_data[
                self.sales_data['product_id'] == product_id
            ].copy()
            
            if len(product_data) < 3:
                continue
                
            product_data['date'] = pd.to_datetime(product_data['date'])
            product_data = product_data.sort_values('date')
            product_data['days_since_start'] = (
                product_data['date'] - product_data['date'].min()
            ).dt.days
            
            trend_metrics = self._calculate_trend_metrics(
                product_data['days_since_start'].values,
                product_data['units_sold'].values,
                product_data['revenue'].values
            )
            
            product_trends[str(product_id)] = {
                'product_name': product_data['product_name'].iloc[0],
                'category': product_data['category'].iloc[0],
                'total_units_sold': product_data['units_sold'].sum(),
                'total_revenue': product_data['revenue'].sum(),
                'trend_metrics': trend_metrics,
                'data_points': len(product_data)
            }
        
        self.trend_data['products'] = product_trends

    async def _analyze_category_trends(self):
        if self.sales_data.empty:
            return
            
        category_trends = {}
        
        category_data = self.sales_data.groupby(['category', 'date']).agg({
            'units_sold': 'sum',
            'revenue': 'sum'
        }).reset_index()
        
        for category in category_data['category'].unique():
            cat_data = category_data[
                category_data['category'] == category
            ].copy()
            
            if len(cat_data) < 3:
                continue
                
            cat_data['date'] = pd.to_datetime(cat_data['date'])
            cat_data = cat_data.sort_values('date')
            cat_data['days_since_start'] = (
                cat_data['date'] - cat_data['date'].min()
            ).dt.days
            
            trend_metrics = self._calculate_trend_metrics(
                cat_data['days_since_start'].values,
                cat_data['units_sold'].values,
                cat_data['revenue'].values
            )
            
            category_trends[category] = {
                'total_units_sold': cat_data['units_sold'].sum(),
                'total_revenue': cat_data['revenue'].sum(),
                'trend_metrics': trend_metrics,
                'data_points': len(cat_data)
            }
        
        self.trend_data['categories'] = category_trends

    async def _analyze_search_trends(self):
        if self.search_data.empty:
            return
            
        daily_searches = self.search_data.groupby('date').agg({
            'search_count': 'sum'
        }).reset_index()
        
        daily_searches['date'] = pd.to_datetime(daily_searches['date'])
        daily_searches = daily_searches.sort_values('date')
        
        if len(daily_searches) >= 3:
            daily_searches['days_since_start'] = (
                daily_searches['date'] - daily_searches['date'].min()
            ).dt.days
            
            search_trend = self._calculate_simple_trend(
                daily_searches['days_since_start'].values,
                daily_searches['search_count'].values
            )
            
            query_trends = self.search_data.groupby('query').agg({
                'search_count': 'sum'
            }).reset_index().sort_values('search_count', ascending=False)
            
            self.trend_data['search'] = {
                'daily_volume_trend': search_trend,
                'total_searches': daily_searches['search_count'].sum(),
                'top_queries': query_trends.head(10).to_dict('records')
            }

    async def _analyze_seasonal_patterns(self):
        """Analyze seasonal patterns in sales data"""
        if self.sales_data.empty:
            return
            
        sales_with_date = self.sales_data.copy()
        sales_with_date['date'] = pd.to_datetime(sales_with_date['date'])
        sales_with_date['day_of_week'] = sales_with_date['date'].dt.dayofweek
        sales_with_date['week_of_year'] = sales_with_date['date'].dt.isocalendar().week
        
        dow_patterns = sales_with_date.groupby('day_of_week').agg({
            'units_sold': 'mean',
            'revenue': 'mean'
        }).reset_index()
        
        dow_patterns['day_name'] = dow_patterns['day_of_week'].map({
            0: 'Monday', 1: 'Tuesday', 2: 'Wednesday', 3: 'Thursday',
            4: 'Friday', 5: 'Saturday', 6: 'Sunday'
        })
        
        weekly_patterns = sales_with_date.groupby('week_of_year').agg({
            'units_sold': 'mean',
            'revenue': 'mean'
        }).reset_index()
        
        self.trend_data['seasonal'] = {
            'day_of_week_patterns': dow_patterns.to_dict('records'),
            'weekly_patterns': weekly_patterns.to_dict('records')
        }

    def _calculate_trend_metrics(self, x_values: np.ndarray, units_sold: np.ndarray, revenue: np.ndarray) -> Dict[str, Any]:
        try:
            X = x_values.reshape(-1, 1)
            
            units_model = LinearRegression()
            units_model.fit(X, units_sold)
            units_slope = units_model.coef_[0]
            units_r2 = units_model.score(X, units_sold)
            
            revenue_model = LinearRegression()
            revenue_model.fit(X, revenue)
            revenue_slope = revenue_model.coef_[0]
            revenue_r2 = revenue_model.score(X, revenue)
            
            units_trend = self._classify_trend(units_slope, units_r2)
            revenue_trend = self._classify_trend(revenue_slope, revenue_r2)
            
            return {
                'units_sold_slope': float(units_slope),
                'units_sold_r2': float(units_r2),
                'units_sold_trend': units_trend,
                'revenue_slope': float(revenue_slope),
                'revenue_r2': float(revenue_r2),
                'revenue_trend': revenue_trend,
                'volatility': float(np.std(units_sold) / np.mean(units_sold)) if np.mean(units_sold) > 0 else 0
            }
            
        except Exception as e:
            logger.error(f"Failed to calculate trend metrics: {e}")
            return {}

    def _calculate_simple_trend(self, x_values: np.ndarray, y_values: np.ndarray) -> Dict[str, Any]:
        try:
            X = x_values.reshape(-1, 1)
            model = LinearRegression()
            model.fit(X, y_values)
            
            slope = model.coef_[0]
            r2 = model.score(X, y_values)
            trend = self._classify_trend(slope, r2)
            
            return {
                'slope': float(slope),
                'r2': float(r2),
                'trend': trend
            }
            
        except Exception as e:
            logger.error(f"Failed to calculate simple trend: {e}")
            return {}

    def _classify_trend(self, slope: float, r2: float) -> str:
        if r2 < 0.3:
            return 'stable'
        elif slope > 0.1:
            return 'growing'
        elif slope < -0.1:
            return 'declining'
        else:
            return 'stable'

    async def get_product_trends(self, limit: int = 20) -> Dict[str, Any]:
        if not self.is_trained:
            await self.initialize()
            
        product_trends = self.trend_data.get('products', {})
        
        trending_products = []
        for product_id, data in product_trends.items():
            metrics = data.get('trend_metrics', {})
            trend_score = self._calculate_trend_score(metrics)
            
            trending_products.append({
                'product_id': product_id,
                'product_name': data['product_name'],
                'category': data['category'],
                'total_units_sold': data['total_units_sold'],
                'total_revenue': data['total_revenue'],
                'trend_score': trend_score,
                'trend_metrics': metrics
            })
        
        trending_products.sort(key=lambda x: x['trend_score'], reverse=True)
        
        return {
            'trending_products': trending_products[:limit],
            'total_products_analyzed': len(product_trends)
        }

    async def get_category_trends(self) -> Dict[str, Any]:
        if not self.is_trained:
            await self.initialize()
            
        return self.trend_data.get('categories', {})

    async def get_search_trends(self) -> Dict[str, Any]:
        if not self.is_trained:
            await self.initialize()
            
        return self.trend_data.get('search', {})

    async def get_seasonal_patterns(self) -> Dict[str, Any]:
        if not self.is_trained:
            await self.initialize()
            
        return self.trend_data.get('seasonal', {})

    async def forecast_demand(self, product_id: str, days_ahead: int = 30) -> Dict[str, Any]:
        try:
            if not self.is_trained:
                await self.initialize()
                
            product_data = self.sales_data[
                self.sales_data['product_id'] == product_id
            ].copy()
            
            if len(product_data) < 5:
                return {
                    'error': 'Insufficient historical data for forecasting',
                    'required_data_points': 5,
                    'available_data_points': len(product_data)
                }
            
            product_data['date'] = pd.to_datetime(product_data['date'])
            product_data = product_data.sort_values('date')
            product_data['days_since_start'] = (
                product_data['date'] - product_data['date'].min()
            ).dt.days
            
            X = product_data['days_since_start'].values.reshape(-1, 1)
            y = product_data['units_sold'].values
            
            model = LinearRegression()
            model.fit(X, y)
            
            last_day = product_data['days_since_start'].max()
            future_days = np.arange(last_day + 1, last_day + days_ahead + 1).reshape(-1, 1)
            forecast = model.predict(future_days)
            
            forecast = np.maximum(forecast, 0)
            
            historical_mae = mean_absolute_error(y, model.predict(X))
            historical_rmse = np.sqrt(mean_squared_error(y, model.predict(X)))
            
            return {
                'product_id': product_id,
                'forecast_period_days': days_ahead,
                'forecasted_units': forecast.tolist(),
                'total_forecasted_units': float(forecast.sum()),
                'confidence_metrics': {
                    'historical_mae': float(historical_mae),
                    'historical_rmse': float(historical_rmse),
                    'model_r2': float(model.score(X, y))
                },
                'historical_data_points': len(product_data)
            }
            
        except Exception as e:
            logger.error(f"Forecasting failed for product {product_id}: {e}")
            return {'error': str(e)}

    def _calculate_trend_score(self, metrics: Dict[str, Any]) -> float:
        if not metrics:
            return 0.0
            
        units_slope = metrics.get('units_sold_slope', 0)
        units_r2 = metrics.get('units_sold_r2', 0)
        revenue_slope = metrics.get('revenue_slope', 0)
        revenue_r2 = metrics.get('revenue_r2', 0)
        volatility = metrics.get('volatility', 1)
        
        trend_score = (
            (units_slope * units_r2 * 0.4) +
            (revenue_slope * revenue_r2 * 0.4) +
            (1 / (1 + volatility) * 0.2)
        )
        
        return float(trend_score)

trend_analyzer = TrendAnalyzer() 