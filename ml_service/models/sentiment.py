import asyncio
import logging
from typing import List, Dict, Any, Optional
import numpy as np
import pandas as pd
from textblob import TextBlob
import re
from datetime import datetime, timedelta

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class SentimentAnalyzer:
    def __init__(self):
        self.is_trained = False
        self.sentiment_data = {}
        
    async def initialize(self):
        """Initialize the sentiment analyzer"""
        try:
            # Initialize empty sentiment data first
            self.sentiment_data = {
                'comments': [],
                'last_updated': datetime.now(),
                'total_analyzed': 0
            }
            
            await self._load_comment_data()
            await self._analyze_sentiments()
            self.is_trained = True
            logger.info("Sentiment analyzer initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize sentiment analyzer: {e}")
            # Ensure we have a valid sentiment_data structure even if initialization fails
            self.sentiment_data = {
                'comments': [],
                'last_updated': datetime.now(),
                'total_analyzed': 0
            }
            self.is_trained = False
            # Don't raise the exception, just log it and continue with empty data

    async def _load_comment_data(self):
        """Load comment data from database"""
        try:
            conn = await get_db_connection()
            try:
                query = """
                SELECT 
                    c.id,
                    c.content,
                    c.rating,
                    c.created_at,
                    p.id as product_id,
                    p.name as product_name,
                    p.category,
                    u.name as user_name
                FROM comments c
                JOIN products p ON c.product_id = p.id
                JOIN users u ON c.user_id = u.id
                WHERE c.created_at >= $1
                ORDER BY c.created_at DESC
                """
                
                cutoff_date = datetime.now() - timedelta(days=90)
                rows = await conn.fetch(query, cutoff_date)
                
                self.comment_data = pd.DataFrame([dict(row) for row in rows])
                
                logger.info(f"Loaded {len(self.comment_data)} comments for sentiment analysis")
            finally:
                await release_db_connection(conn)
        except Exception as e:
            logger.error(f"Failed to load comment data: {e}")
            # Initialize empty DataFrame if loading fails
            self.comment_data = pd.DataFrame()

    async def _analyze_sentiments(self):
        """Analyze sentiments of comments"""
        if self.comment_data.empty:
            # Initialize empty sentiment data if no comments
            self.sentiment_data = {
                'comments': [],
                'last_updated': datetime.now(),
                'total_analyzed': 0
            }
            return
            
        sentiments = []
        
        for _, comment in self.comment_data.iterrows():
            sentiment_score = self._analyze_text_sentiment(comment['content'])
            
            sentiments.append({
                'comment_id': comment['id'],
                'product_id': comment['product_id'],
                'product_name': comment['product_name'],
                'category': comment['category'],
                'content': comment['content'],
                'rating': comment['rating'],
                'sentiment_score': sentiment_score['compound'],
                'sentiment_label': sentiment_score['label'],
                'polarity': sentiment_score['polarity'],
                'subjectivity': sentiment_score['subjectivity'],
                'created_at': comment['created_at']
            })
        
        self.sentiment_data = {
            'comments': sentiments,
            'last_updated': datetime.now(),
            'total_analyzed': len(sentiments)
        }

    def _analyze_text_sentiment(self, text: str) -> Dict[str, Any]:
        """Analyze sentiment of a single text"""
        try:
            # Clean text
            cleaned_text = self._clean_text(text)
            
            # Use TextBlob for sentiment analysis
            blob = TextBlob(cleaned_text)
            
            polarity = blob.sentiment.polarity  # -1 to 1
            subjectivity = blob.sentiment.subjectivity  # 0 to 1
            
            # Convert to compound score similar to VADER
            compound = polarity
            
            # Determine sentiment label
            if compound >= 0.05:
                label = 'positive'
            elif compound <= -0.05:
                label = 'negative'
            else:
                label = 'neutral'
            
            return {
                'polarity': polarity,
                'subjectivity': subjectivity,
                'compound': compound,
                'label': label
            }
            
        except Exception as e:
            logger.error(f"Error analyzing sentiment: {e}")
            return {
                'polarity': 0.0,
                'subjectivity': 0.0,
                'compound': 0.0,
                'label': 'neutral'
            }

    def _clean_text(self, text: str) -> str:
        """Clean text for sentiment analysis"""
        # Remove special characters and extra whitespace
        text = re.sub(r'[^\w\s]', ' ', text)
        text = re.sub(r'\s+', ' ', text).strip()
        return text.lower()

    async def analyze_product_sentiment(self, product_id: str) -> Dict[str, Any]:
        """Analyze sentiment for a specific product"""
        if not self.is_trained:
            await self.initialize()
            
        # Ensure sentiment_data exists and has comments
        if not hasattr(self, 'sentiment_data') or not self.sentiment_data or 'comments' not in self.sentiment_data:
            return {
                'product_id': product_id,
                'total_comments': 0,
                'sentiment_summary': {
                    'positive': 0,
                    'negative': 0,
                    'neutral': 0
                },
                'average_sentiment': 0.0,
                'sentiment_trend': 'no_data'
            }
            
        product_sentiments = [
            s for s in self.sentiment_data['comments'] 
            if str(s['product_id']) == product_id
        ]
        
        if not product_sentiments:
            return {
                'product_id': product_id,
                'total_comments': 0,
                'sentiment_summary': {
                    'positive': 0,
                    'negative': 0,
                    'neutral': 0
                },
                'average_sentiment': 0.0,
                'sentiment_trend': 'no_data'
            }
        
        # Calculate sentiment distribution
        positive_count = len([s for s in product_sentiments if s['sentiment_label'] == 'positive'])
        negative_count = len([s for s in product_sentiments if s['sentiment_label'] == 'negative'])
        neutral_count = len([s for s in product_sentiments if s['sentiment_label'] == 'neutral'])
        
        # Calculate average sentiment
        avg_sentiment = np.mean([s['sentiment_score'] for s in product_sentiments])
        
        # Determine overall trend
        if avg_sentiment > 0.1:
            trend = 'positive'
        elif avg_sentiment < -0.1:
            trend = 'negative'
        else:
            trend = 'neutral'
        
        return {
            'product_id': product_id,
            'total_comments': len(product_sentiments),
            'sentiment_summary': {
                'positive': positive_count,
                'negative': negative_count,
                'neutral': neutral_count
            },
            'average_sentiment': float(avg_sentiment),
            'sentiment_trend': trend,
            'recent_comments': product_sentiments[:5]  # Most recent 5 comments
        }

    async def get_category_sentiment(self, category: str) -> Dict[str, Any]:
        """Get sentiment analysis for a product category"""
        if not self.is_trained:
            await self.initialize()
            
        # Ensure sentiment_data exists and has comments
        if not hasattr(self, 'sentiment_data') or not self.sentiment_data or 'comments' not in self.sentiment_data:
            return {
                'category': category,
                'total_comments': 0,
                'average_sentiment': 0.0
            }
            
        category_sentiments = [
            s for s in self.sentiment_data['comments'] 
            if s['category'].lower() == category.lower()
        ]
        
        if not category_sentiments:
            return {
                'category': category,
                'total_comments': 0,
                'average_sentiment': 0.0
            }
        
        avg_sentiment = np.mean([s['sentiment_score'] for s in category_sentiments])
        
        # Group by product
        product_sentiments = {}
        for sentiment in category_sentiments:
            product_id = sentiment['product_id']
            if product_id not in product_sentiments:
                product_sentiments[product_id] = {
                    'product_name': sentiment['product_name'],
                    'sentiments': []
                }
            product_sentiments[product_id]['sentiments'].append(sentiment['sentiment_score'])
        
        # Calculate average sentiment per product
        product_averages = []
        for product_id, data in product_sentiments.items():
            avg = np.mean(data['sentiments'])
            product_averages.append({
                'product_id': product_id,
                'product_name': data['product_name'],
                'average_sentiment': float(avg),
                'comment_count': len(data['sentiments'])
            })
        
        # Sort by sentiment
        product_averages.sort(key=lambda x: x['average_sentiment'], reverse=True)
        
        return {
            'category': category,
            'total_comments': len(category_sentiments),
            'average_sentiment': float(avg_sentiment),
            'top_products': product_averages[:10],
            'bottom_products': product_averages[-5:] if len(product_averages) > 5 else []
        }

    async def get_sentiment_trends(self, days: int = 30) -> Dict[str, Any]:
        """Get sentiment trends over time"""
        if not self.is_trained:
            await self.initialize()
            
        # Ensure sentiment_data exists and has comments
        if not hasattr(self, 'sentiment_data') or not self.sentiment_data or 'comments' not in self.sentiment_data:
            return {
                'period_days': days,
                'total_comments': 0,
                'trends': [],
                'overall_trend': 'no_data'
            }
            
        cutoff_date = datetime.now() - timedelta(days=days)
        
        # Filter recent comments
        recent_sentiments = [
            s for s in self.sentiment_data['comments']
            if s['created_at'] >= cutoff_date
        ]
        
        if not recent_sentiments:
            return {
                'period_days': days,
                'total_comments': 0,
                'trends': []
            }
        
        # Group by date
        daily_sentiments = {}
        for sentiment in recent_sentiments:
            date_key = sentiment['created_at'].date()
            if date_key not in daily_sentiments:
                daily_sentiments[date_key] = []
            daily_sentiments[date_key].append(sentiment['sentiment_score'])
        
        # Calculate daily averages
        trends = []
        for date, scores in sorted(daily_sentiments.items()):
            avg_sentiment = np.mean(scores)
            trends.append({
                'date': date.isoformat(),
                'average_sentiment': float(avg_sentiment),
                'comment_count': len(scores)
            })
        
        return {
            'period_days': days,
            'total_comments': len(recent_sentiments),
            'trends': trends,
            'overall_trend': self._calculate_trend_direction(trends)
        }

    def _calculate_trend_direction(self, trends: List[Dict]) -> str:
        """Calculate overall trend direction"""
        if len(trends) < 2:
            return 'insufficient_data'
        
        # Compare first half with second half
        mid_point = len(trends) // 2
        first_half_avg = np.mean([t['average_sentiment'] for t in trends[:mid_point]])
        second_half_avg = np.mean([t['average_sentiment'] for t in trends[mid_point:]])
        
        diff = second_half_avg - first_half_avg
        
        if diff > 0.05:
            return 'improving'
        elif diff < -0.05:
            return 'declining'
        else:
            return 'stable'

    async def get_sentiment_insights(self) -> Dict[str, Any]:
        """Get actionable sentiment insights"""
        if not self.is_trained:
            await self.initialize()
            
        # Ensure sentiment_data exists and has comments
        if not hasattr(self, 'sentiment_data') or not self.sentiment_data or 'comments' not in self.sentiment_data:
            return {
                'insights': [],
                'summary': {
                    'total_products_analyzed': 0,
                    'products_with_negative_sentiment': 0,
                    'products_with_positive_sentiment': 0,
                    'overall_sentiment_health': 'no_data'
                },
                'generated_at': datetime.now().isoformat()
            }
            
        insights = []
        
        # Find products with consistently negative sentiment
        negative_products = []
        positive_products = []
        
        # Group sentiments by product
        product_groups = {}
        for sentiment in self.sentiment_data['comments']:
            product_id = sentiment['product_id']
            if product_id not in product_groups:
                product_groups[product_id] = {
                    'product_name': sentiment['product_name'],
                    'category': sentiment['category'],
                    'sentiments': []
                }
            product_groups[product_id]['sentiments'].append(sentiment['sentiment_score'])
        
        # Analyze each product
        for product_id, data in product_groups.items():
            if len(data['sentiments']) >= 3:  # Minimum 3 comments
                avg_sentiment = np.mean(data['sentiments'])
                
                if avg_sentiment < -0.2:
                    negative_products.append({
                        'product_id': product_id,
                        'product_name': data['product_name'],
                        'category': data['category'],
                        'average_sentiment': float(avg_sentiment),
                        'comment_count': len(data['sentiments'])
                    })
                elif avg_sentiment > 0.3:
                    positive_products.append({
                        'product_id': product_id,
                        'product_name': data['product_name'],
                        'category': data['category'],
                        'average_sentiment': float(avg_sentiment),
                        'comment_count': len(data['sentiments'])
                    })
        
        # Generate insights
        if negative_products:
            insights.append({
                'type': 'warning',
                'title': 'Products with Negative Sentiment',
                'description': f'Found {len(negative_products)} products with consistently negative reviews',
                'products': sorted(negative_products, key=lambda x: x['average_sentiment'])[:5],
                'action': 'Review product quality, customer service, or pricing'
            })
        
        if positive_products:
            insights.append({
                'type': 'success',
                'title': 'Top Performing Products',
                'description': f'Found {len(positive_products)} products with excellent customer sentiment',
                'products': sorted(positive_products, key=lambda x: x['average_sentiment'], reverse=True)[:5],
                'action': 'Consider promoting these products or expanding similar offerings'
            })
        
        return {
            'insights': insights,
            'summary': {
                'total_products_analyzed': len(product_groups),
                'products_with_negative_sentiment': len(negative_products),
                'products_with_positive_sentiment': len(positive_products),
                'overall_sentiment_health': 'good' if len(positive_products) > len(negative_products) else 'needs_attention'
            },
            'generated_at': datetime.now().isoformat()
        }

# Global sentiment analyzer instance
sentiment_analyzer = SentimentAnalyzer() 