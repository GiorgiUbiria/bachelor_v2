import asyncio
import logging
from typing import List, Dict, Any, Optional
import numpy as np
import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.decomposition import TruncatedSVD
import re
import json
from datetime import datetime, timedelta

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class EnhancedSearchEngine:
    def __init__(self):
        self.tfidf_vectorizer = None
        self.product_features = None
        self.product_data = None
        self.search_history = []
        self.query_expansion_dict = {}
        self.is_trained = False
        
    async def initialize(self):
        """Initialize the search engine with product data"""
        try:
            await self._load_product_data()
            await self._build_search_index()
            await self._load_search_history()
            await self._build_query_expansion()
            self.is_trained = True
            logger.info("Enhanced search engine initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize search engine: {e}")
            raise

    async def _load_product_data(self):
        """Load product data from database"""
        conn = await get_db_connection()
        try:
            query = """
            SELECT 
                id, name, description, category, price, stock,
                created_at
            FROM products 
            WHERE stock > 0
            ORDER BY created_at DESC
            """
            
            rows = await conn.fetch(query)
            
            self.product_data = pd.DataFrame([
                {
                    'id': str(row['id']),
                    'name': row['name'],
                    'description': row['description'] or '',
                    'category': row['category'],
                    'price': float(row['price']),
                    'stock': row['stock'],
                    'created_at': row['created_at']
                }
                for row in rows
            ])
            
            logger.info(f"Loaded {len(self.product_data)} products for search indexing")
        finally:
            await release_db_connection(conn)

    async def _build_search_index(self):
        """Build TF-IDF search index"""
        if self.product_data.empty:
            return
            
        # Combine text fields for indexing
        self.product_data['search_text'] = (
            self.product_data['name'] + ' ' +
            self.product_data['description'] + ' ' +
            self.product_data['category']
        ).str.lower()
        
        # Build TF-IDF vectorizer
        self.tfidf_vectorizer = TfidfVectorizer(
            max_features=5000,
            stop_words='english',
            ngram_range=(1, 2),
            min_df=1,
            max_df=0.95
        )
        
        self.product_features = self.tfidf_vectorizer.fit_transform(
            self.product_data['search_text']
        )
        
        logger.info("Search index built successfully")

    async def _load_search_history(self):
        """Load recent search history for query expansion"""
        conn = await get_db_connection()
        try:
            query = """
            SELECT query, results_count, created_at
            FROM search_queries 
            WHERE created_at >= $1
            ORDER BY created_at DESC
            LIMIT 1000
            """
            
            cutoff_date = datetime.now() - timedelta(days=30)
            rows = await conn.fetch(query, cutoff_date)
            
            self.search_history = [
                {
                    'query': row['query'],
                    'results_count': row['results_count'],
                    'created_at': row['created_at']
                }
                for row in rows
            ]
        except Exception as e:
            # If search_queries table doesn't exist, just continue
            logger.warning(f"Could not load search history: {e}")
            self.search_history = []
        finally:
            await release_db_connection(conn)

    async def _build_query_expansion(self):
        """Build query expansion dictionary from search history"""
        # Simple query expansion based on common search patterns
        self.query_expansion_dict = {
            'phone': ['smartphone', 'mobile', 'cell phone'],
            'laptop': ['notebook', 'computer', 'pc'],
            'headphones': ['earphones', 'earbuds', 'headset'],
            'watch': ['smartwatch', 'timepiece'],
            'camera': ['photography', 'photo', 'lens'],
            'gaming': ['game', 'gamer', 'console'],
            'wireless': ['bluetooth', 'wifi', 'cordless'],
            'portable': ['mobile', 'travel', 'compact'],
            'professional': ['pro', 'business', 'work'],
            'budget': ['cheap', 'affordable', 'low cost']
        }

    def _expand_query(self, query: str) -> str:
        """Expand query with synonyms and related terms"""
        expanded_terms = [query]
        query_lower = query.lower()
        
        for key, synonyms in self.query_expansion_dict.items():
            if key in query_lower:
                expanded_terms.extend(synonyms)
            elif any(syn in query_lower for syn in synonyms):
                expanded_terms.append(key)
                
        return ' '.join(expanded_terms)

    def _preprocess_query(self, query: str) -> str:
        """Preprocess search query"""
        # Remove special characters and normalize
        query = re.sub(r'[^\w\s]', ' ', query)
        query = re.sub(r'\s+', ' ', query).strip().lower()
        
        # Expand query
        query = self._expand_query(query)
        
        return query

    async def search(
        self, 
        query: str, 
        user_id: Optional[str] = None,
        category: Optional[str] = None,
        min_price: Optional[float] = None,
        max_price: Optional[float] = None,
        sort_by: str = 'relevance',
        limit: int = 20
    ) -> Dict[str, Any]:
        """
        Perform enhanced search with ML features
        """
        try:
            if not self.is_trained:
                await self.initialize()
                
            # Preprocess query
            processed_query = self._preprocess_query(query)
            
            # Get base results
            results = await self._semantic_search(processed_query, limit * 2)
            
            # Apply filters
            if category:
                results = [r for r in results if r['category'].lower() == category.lower()]
            
            if min_price is not None:
                results = [r for r in results if r['price'] >= min_price]
                
            if max_price is not None:
                results = [r for r in results if r['price'] <= max_price]
            
            # Apply personalization if user provided
            if user_id:
                results = await self._personalize_results(results, user_id)
            
            # Sort results
            results = self._sort_results(results, sort_by)
            
            # Limit results
            results = results[:limit]
            
            # Log search query
            await self._log_search_query(query, user_id, len(results))
            
            # Generate search suggestions
            suggestions = await self._generate_suggestions(query)
            
            return {
                'results': results,
                'total_count': len(results),
                'query': query,
                'processed_query': processed_query,
                'suggestions': suggestions,
                'search_time': datetime.now().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Search failed: {e}")
            return {
                'results': [],
                'total_count': 0,
                'query': query,
                'error': str(e)
            }

    async def _semantic_search(self, query: str, limit: int) -> List[Dict[str, Any]]:
        """Perform semantic search using TF-IDF"""
        if self.tfidf_vectorizer is None or self.product_features is None:
            return []
            
        # Transform query
        query_vector = self.tfidf_vectorizer.transform([query])
        
        # Calculate similarities
        similarities = cosine_similarity(query_vector, self.product_features).flatten()
        
        # Get top results
        top_indices = similarities.argsort()[-limit:][::-1]
        
        results = []
        for idx in top_indices:
            if similarities[idx] > 0:  # Only include relevant results
                product = self.product_data.iloc[idx]
                results.append({
                    'id': str(product['id']),
                    'name': str(product['name']),
                    'description': str(product['description']),
                    'category': str(product['category']),
                    'price': float(product['price']),
                    'stock': int(product['stock']),
                    'relevance_score': float(similarities[idx]),
                    'created_at': product['created_at'].isoformat() if pd.notna(product['created_at']) else None
                })
                
        return results

    async def _personalize_results(self, results: List[Dict[str, Any]], user_id: str) -> List[Dict[str, Any]]:
        """Personalize search results based on user behavior"""
        try:
            conn = await get_db_connection()
            try:
                # Get user interaction history
                query = """
                SELECT product_id, interaction_type, COUNT(*) as interaction_count
                FROM user_interactions 
                WHERE user_id = $1 
                AND created_at >= $2
                GROUP BY product_id, interaction_type
                """
                
                cutoff_date = datetime.now() - timedelta(days=30)
                rows = await conn.fetch(query, user_id, cutoff_date)
                
                # Build user preference scores
                user_preferences = {}
                for row in rows:
                    product_id = str(row['product_id'])
                    interaction_type = row['interaction_type']
                    count = row['interaction_count']
                    
                    if product_id not in user_preferences:
                        user_preferences[product_id] = 0
                    
                    # Weight different interaction types
                    weights = {
                        'view': 1,
                        'cart_add': 3,
                        'purchase': 5,
                        'like': 2
                    }
                    
                    user_preferences[product_id] += weights.get(interaction_type, 1) * count
                
                # Boost scores for products user has interacted with
                for result in results:
                    if result['id'] in user_preferences:
                        boost = min(user_preferences[result['id']] * 0.1, 0.5)
                        result['relevance_score'] += boost
                        result['personalized'] = True
                    else:
                        result['personalized'] = False
            finally:
                await release_db_connection(conn)
                        
        except Exception as e:
            logger.error(f"Personalization failed: {e}")
            
        return results

    def _sort_results(self, results: List[Dict[str, Any]], sort_by: str) -> List[Dict[str, Any]]:
        """Sort search results"""
        if sort_by == 'price_low':
            return sorted(results, key=lambda x: x['price'])
        elif sort_by == 'price_high':
            return sorted(results, key=lambda x: x['price'], reverse=True)
        elif sort_by == 'newest':
            return sorted(results, key=lambda x: x['created_at'] or '', reverse=True)
        elif sort_by == 'name':
            return sorted(results, key=lambda x: x['name'])
        else:  # relevance (default)
            return sorted(results, key=lambda x: x['relevance_score'], reverse=True)

    async def _log_search_query(self, query: str, user_id: Optional[str], results_count: int):
        """Log search query for analytics"""
        try:
            conn = await get_db_connection()
            try:
                await conn.execute("""
                    INSERT INTO search_queries (query, user_id, results_count, created_at)
                    VALUES ($1, $2, $3, $4)
                """, query, user_id, results_count, datetime.now())
            finally:
                await release_db_connection(conn)
        except Exception as e:
            logger.error(f"Failed to log search query: {e}")

    async def _generate_suggestions(self, query: str) -> List[str]:
        """Generate search suggestions based on query"""
        suggestions = []
        
        # Add query expansions
        expanded = self._expand_query(query)
        if expanded != query:
            suggestions.extend(expanded.split())
        
        # Add popular searches from history
        popular_queries = [
            item['query'] for item in self.search_history[:10]
            if query.lower() in item['query'].lower() and item['query'] != query
        ]
        suggestions.extend(popular_queries[:3])
        
        # Remove duplicates and limit
        suggestions = list(dict.fromkeys(suggestions))[:5]
        
        return suggestions

    async def get_search_analytics(self, days: int = 30) -> Dict[str, Any]:
        """Get search analytics"""
        try:
            conn = await get_db_connection()
            try:
                cutoff_date = datetime.now() - timedelta(days=days)
                
                # Top search queries
                top_queries = await conn.fetch("""
                    SELECT query, COUNT(*) as search_count
                    FROM search_queries 
                    WHERE created_at >= $1
                    GROUP BY query
                    ORDER BY search_count DESC
                    LIMIT 10
                """, cutoff_date)
                
                # Search volume over time
                search_volume = await conn.fetch("""
                    SELECT DATE(created_at) as date, COUNT(*) as searches
                    FROM search_queries 
                    WHERE created_at >= $1
                    GROUP BY DATE(created_at)
                    ORDER BY date
                """, cutoff_date)
                
                # Zero result queries
                zero_results = await conn.fetch("""
                    SELECT query, COUNT(*) as count
                    FROM search_queries 
                    WHERE created_at >= $1 AND results_count = 0
                    GROUP BY query
                    ORDER BY count DESC
                    LIMIT 10
                """, cutoff_date)
                
                return {
                    'top_queries': [dict(row) for row in top_queries],
                    'search_volume': [dict(row) for row in search_volume],
                    'zero_results': [dict(row) for row in zero_results],
                    'total_searches': sum(row['search_count'] for row in top_queries)
                }
            finally:
                await release_db_connection(conn)
                
        except Exception as e:
            logger.error(f"Failed to get search analytics: {e}")
            return {}

# Global search engine instance
search_engine = EnhancedSearchEngine() 