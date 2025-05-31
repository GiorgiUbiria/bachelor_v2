import asyncio
import logging
from typing import List, Dict, Any, Optional, Set
import numpy as np
import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.cluster import KMeans
from sklearn.metrics.pairwise import cosine_similarity
import re
from collections import Counter
from datetime import datetime, timedelta

from database.connection import get_db_connection, release_db_connection

logger = logging.getLogger(__name__)

class AutoTagger:
    def __init__(self):
        self.is_trained = False
        self.tfidf_vectorizer = None
        self.product_features = None
        self.tag_suggestions = {}
        self.category_keywords = {}
        self.interaction_patterns = {}
        
    async def initialize(self):
        """Initialize the auto-tagger"""
        try:
            await self._load_product_data()
            await self._load_interaction_data()
            await self._build_keyword_models()
            await self._analyze_interaction_patterns()
            self.is_trained = True
            logger.info("Auto-tagger initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize auto-tagger: {e}")
            raise

    async def _load_product_data(self):
        """Load product data for analysis"""
        conn = await get_db_connection()
        try:
            query = """
            SELECT 
                p.id,
                p.name,
                p.description,
                p.category,
                p.price,
                COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}') as existing_tags
            FROM products p
            LEFT JOIN product_tags pt ON p.id = pt.product_id
            LEFT JOIN tags t ON pt.tag_id = t.id
            GROUP BY p.id, p.name, p.description, p.category, p.price
            """
            
            rows = await conn.fetch(query)
            
            self.product_data = pd.DataFrame([
                {
                    'id': str(row['id']),
                    'name': row['name'],
                    'description': row['description'] or '',
                    'category': row['category'],
                    'price': float(row['price']),
                    'existing_tags': list(row['existing_tags']) if row['existing_tags'] else []
                }
                for row in rows
            ])
            
            logger.info(f"Loaded {len(self.product_data)} products for auto-tagging")
        finally:
            await release_db_connection(conn)

    async def _load_interaction_data(self):
        """Load user interaction data"""
        conn = await get_db_connection()
        try:
            query = """
            SELECT 
                ui.product_id,
                ui.interaction_type,
                COUNT(*) as interaction_count,
                p.category
            FROM user_interactions ui
            JOIN products p ON ui.product_id = p.id
            WHERE ui.created_at >= $1
            GROUP BY ui.product_id, ui.interaction_type, p.category
            """
            
            cutoff_date = datetime.now() - timedelta(days=30)
            rows = await conn.fetch(query, cutoff_date)
            
            self.interaction_data = pd.DataFrame([dict(row) for row in rows])
            
            logger.info(f"Loaded {len(self.interaction_data)} interaction records")
        finally:
            await release_db_connection(conn)

    async def _build_keyword_models(self):
        """Build keyword extraction models"""
        if self.product_data.empty:
            return
            
        # Combine text fields for analysis
        self.product_data['combined_text'] = (
            self.product_data['name'] + ' ' + 
            self.product_data['description'] + ' ' + 
            self.product_data['category']
        ).str.lower()
        
        # Build TF-IDF vectorizer
        self.tfidf_vectorizer = TfidfVectorizer(
            max_features=1000,
            stop_words='english',
            ngram_range=(1, 2),
            min_df=2,
            max_df=0.8
        )
        
        self.product_features = self.tfidf_vectorizer.fit_transform(
            self.product_data['combined_text']
        )
        
        # Extract category-specific keywords
        await self._extract_category_keywords()

    async def _extract_category_keywords(self):
        """Extract keywords for each category"""
        categories = self.product_data['category'].unique()
        
        for category in categories:
            category_products = self.product_data[
                self.product_data['category'] == category
            ]
            
            if len(category_products) < 2:
                continue
                
            # Get category-specific text
            category_text = ' '.join(category_products['combined_text'])
            
            # Extract keywords using TF-IDF
            category_vectorizer = TfidfVectorizer(
                max_features=50,
                stop_words='english',
                ngram_range=(1, 2),
                min_df=1
            )
            
            try:
                category_features = category_vectorizer.fit_transform([category_text])
                feature_names = category_vectorizer.get_feature_names_out()
                scores = category_features.toarray()[0]
                
                # Get top keywords
                keyword_scores = list(zip(feature_names, scores))
                keyword_scores.sort(key=lambda x: x[1], reverse=True)
                
                self.category_keywords[category] = [
                    keyword for keyword, score in keyword_scores[:20]
                    if len(keyword) > 2 and not keyword.isdigit()
                ]
                
            except Exception as e:
                logger.warning(f"Failed to extract keywords for category {category}: {e}")
                self.category_keywords[category] = []

    async def _analyze_interaction_patterns(self):
        """Analyze interaction patterns for tagging insights"""
        if self.interaction_data.empty:
            return
            
        # Group by product and calculate interaction scores
        product_interactions = self.interaction_data.groupby('product_id').agg({
            'interaction_count': 'sum'
        }).reset_index()
        
        # Merge with product data
        interaction_analysis = product_interactions.merge(
            self.product_data[['id', 'category', 'name']],
            left_on='product_id',
            right_on='id',
            how='left'
        )
        
        # Identify high-interaction products by category
        for category in interaction_analysis['category'].unique():
            category_data = interaction_analysis[
                interaction_analysis['category'] == category
            ]
            
            if len(category_data) > 0:
                # Find top interacted products
                top_products = category_data.nlargest(5, 'interaction_count')
                
                self.interaction_patterns[category] = {
                    'top_products': top_products['name'].tolist(),
                    'avg_interactions': category_data['interaction_count'].mean(),
                    'total_products': len(category_data)
                }

    async def suggest_tags_for_product(self, product_id: str) -> Dict[str, Any]:
        """Suggest tags for a specific product"""
        if not self.is_trained:
            await self.initialize()
            
        # Find product
        product = self.product_data[self.product_data['id'] == product_id]
        
        if product.empty:
            return {
                'product_id': product_id,
                'suggested_tags': [],
                'confidence_scores': [],
                'reasoning': 'Product not found'
            }
        
        product_info = product.iloc[0]
        suggested_tags = []
        confidence_scores = []
        reasoning = []
        
        # 1. Category-based tags
        category = product_info['category']
        if category in self.category_keywords:
            category_tags = self._extract_relevant_keywords(
                product_info['combined_text'],
                self.category_keywords[category]
            )
            
            for tag, score in category_tags[:5]:
                suggested_tags.append(tag)
                confidence_scores.append(score)
                reasoning.append(f"Category '{category}' keyword")
        
        # 2. Content-based tags
        content_tags = self._extract_content_tags(product_info['combined_text'])
        for tag, score in content_tags[:3]:
            if tag not in suggested_tags:
                suggested_tags.append(tag)
                confidence_scores.append(score)
                reasoning.append("Content analysis")
        
        # 3. Price-based tags
        price_tags = self._get_price_based_tags(product_info['price'])
        for tag in price_tags:
            if tag not in suggested_tags:
                suggested_tags.append(tag)
                confidence_scores.append(0.7)
                reasoning.append("Price range analysis")
        
        # 4. Interaction-based tags
        interaction_tags = await self._get_interaction_based_tags(product_id)
        for tag, score in interaction_tags:
            if tag not in suggested_tags:
                suggested_tags.append(tag)
                confidence_scores.append(score)
                reasoning.append("User interaction patterns")
        
        # 5. Similar product tags
        similar_tags = await self._get_similar_product_tags(product_id)
        for tag, score in similar_tags[:2]:
            if tag not in suggested_tags:
                suggested_tags.append(tag)
                confidence_scores.append(score)
                reasoning.append("Similar product analysis")
        
        return {
            'product_id': product_id,
            'product_name': product_info['name'],
            'category': category,
            'existing_tags': product_info['existing_tags'],
            'suggested_tags': suggested_tags[:10],  # Limit to top 10
            'confidence_scores': confidence_scores[:10],
            'reasoning': reasoning[:10],
            'generated_at': datetime.now().isoformat()
        }

    def _extract_relevant_keywords(self, text: str, keywords: List[str]) -> List[tuple]:
        """Extract relevant keywords from text"""
        text_lower = text.lower()
        relevant = []
        
        for keyword in keywords:
            if keyword in text_lower:
                # Calculate relevance score based on frequency and position
                count = text_lower.count(keyword)
                # Boost score if keyword appears in product name (beginning of text)
                position_boost = 1.5 if text_lower.find(keyword) < 50 else 1.0
                score = min(count * position_boost * 0.1, 1.0)
                relevant.append((keyword, score))
        
        return sorted(relevant, key=lambda x: x[1], reverse=True)

    def _extract_content_tags(self, text: str) -> List[tuple]:
        """Extract tags from product content"""
        # Common product attribute patterns
        patterns = {
            'wireless': r'\b(wireless|bluetooth|wifi)\b',
            'portable': r'\b(portable|mobile|travel|compact)\b',
            'professional': r'\b(professional|pro|business|commercial)\b',
            'gaming': r'\b(gaming|gamer|game)\b',
            'smart': r'\b(smart|intelligent|ai)\b',
            'premium': r'\b(premium|luxury|high-end|deluxe)\b',
            'eco-friendly': r'\b(eco|green|sustainable|organic)\b',
            'waterproof': r'\b(waterproof|water-resistant|splash-proof)\b',
            'rechargeable': r'\b(rechargeable|battery|usb)\b',
            'adjustable': r'\b(adjustable|customizable|flexible)\b'
        }
        
        tags = []
        text_lower = text.lower()
        
        for tag, pattern in patterns.items():
            if re.search(pattern, text_lower):
                # Calculate confidence based on pattern strength
                matches = len(re.findall(pattern, text_lower))
                confidence = min(matches * 0.3, 1.0)
                tags.append((tag, confidence))
        
        return sorted(tags, key=lambda x: x[1], reverse=True)

    def _get_price_based_tags(self, price: float) -> List[str]:
        """Get tags based on price range"""
        tags = []
        
        if price < 25:
            tags.extend(['budget', 'affordable', 'value'])
        elif price < 100:
            tags.extend(['mid-range', 'reasonable'])
        elif price < 500:
            tags.extend(['premium', 'quality'])
        else:
            tags.extend(['luxury', 'high-end', 'professional'])
        
        return tags

    async def _get_interaction_based_tags(self, product_id: str) -> List[tuple]:
        """Get tags based on user interactions"""
        if self.interaction_data.empty:
            return []
        
        product_interactions = self.interaction_data[
            self.interaction_data['product_id'] == product_id
        ]
        
        if product_interactions.empty:
            return []
        
        tags = []
        total_interactions = product_interactions['interaction_count'].sum()
        
        # Analyze interaction types
        interaction_summary = product_interactions.groupby('interaction_type')['interaction_count'].sum()
        
        if 'view' in interaction_summary and interaction_summary['view'] > 50:
            tags.append(('popular', 0.8))
        
        if 'cart_add' in interaction_summary and interaction_summary['cart_add'] > 10:
            tags.append(('in-demand', 0.7))
        
        if 'purchase' in interaction_summary and interaction_summary['purchase'] > 5:
            tags.append(('bestseller', 0.9))
        
        if 'favorite' in interaction_summary and interaction_summary['favorite'] > 3:
            tags.append(('customer-favorite', 0.8))
        
        return tags

    async def _get_similar_product_tags(self, product_id: str) -> List[tuple]:
        """Get tags from similar products"""
        if self.product_features is None:
            return []
        
        # Find product index
        product_idx = None
        for idx, pid in enumerate(self.product_data['id']):
            if pid == product_id:
                product_idx = idx
                break
        
        if product_idx is None:
            return []
        
        # Calculate similarities
        product_vector = self.product_features[product_idx]
        similarities = cosine_similarity(product_vector, self.product_features).flatten()
        
        # Get top 5 similar products (excluding self)
        similar_indices = similarities.argsort()[::-1][1:6]
        
        # Collect tags from similar products
        tag_counts = Counter()
        
        for idx in similar_indices:
            similar_product = self.product_data.iloc[idx]
            for tag in similar_product['existing_tags']:
                tag_counts[tag] += similarities[idx]
        
        # Return top tags with confidence scores
        return [(tag, min(count, 1.0)) for tag, count in tag_counts.most_common(5)]

    async def auto_tag_products(self, limit: int = 50) -> Dict[str, Any]:
        """Automatically suggest tags for products that need them"""
        if not self.is_trained:
            await self.initialize()
        
        # Find products with few or no tags
        products_needing_tags = []
        
        for _, product in self.product_data.iterrows():
            if len(product['existing_tags']) < 3:  # Products with fewer than 3 tags
                products_needing_tags.append(product['id'])
        
        # Limit the number of products to process
        products_to_process = products_needing_tags[:limit]
        
        results = []
        for product_id in products_to_process:
            suggestions = await self.suggest_tags_for_product(product_id)
            
            # Only include products with high-confidence suggestions
            high_confidence_tags = [
                tag for tag, score in zip(suggestions['suggested_tags'], suggestions['confidence_scores'])
                if score > 0.6
            ]
            
            if high_confidence_tags:
                results.append({
                    'product_id': product_id,
                    'product_name': suggestions['product_name'],
                    'category': suggestions['category'],
                    'existing_tags': suggestions['existing_tags'],
                    'suggested_tags': high_confidence_tags,
                    'auto_apply_recommended': len(high_confidence_tags) >= 2
                })
        
        return {
            'total_products_analyzed': len(products_to_process),
            'products_with_suggestions': len(results),
            'suggestions': results,
            'generated_at': datetime.now().isoformat()
        }

    async def get_tagging_insights(self) -> Dict[str, Any]:
        """Get insights about current tagging status"""
        if not self.is_trained:
            await self.initialize()
        
        insights = []
        
        # Analyze tag coverage by category
        category_stats = {}
        for category in self.product_data['category'].unique():
            category_products = self.product_data[self.product_data['category'] == category]
            
            total_products = len(category_products)
            products_with_tags = len([
                p for _, p in category_products.iterrows() 
                if len(p['existing_tags']) > 0
            ])
            
            avg_tags_per_product = np.mean([
                len(p['existing_tags']) for _, p in category_products.iterrows()
            ])
            
            category_stats[category] = {
                'total_products': total_products,
                'products_with_tags': products_with_tags,
                'coverage_percentage': (products_with_tags / total_products) * 100,
                'avg_tags_per_product': avg_tags_per_product
            }
        
        # Find categories with low tag coverage
        low_coverage_categories = [
            cat for cat, stats in category_stats.items()
            if stats['coverage_percentage'] < 50
        ]
        
        if low_coverage_categories:
            insights.append({
                'type': 'warning',
                'title': 'Low Tag Coverage Categories',
                'description': f'Found {len(low_coverage_categories)} categories with less than 50% tag coverage',
                'categories': low_coverage_categories,
                'action': 'Consider running auto-tagging for these categories'
            })
        
        # Find most common tags
        all_tags = []
        for _, product in self.product_data.iterrows():
            all_tags.extend(product['existing_tags'])
        
        tag_counts = Counter(all_tags)
        
        return {
            'insights': insights,
            'category_stats': category_stats,
            'most_common_tags': tag_counts.most_common(20),
            'total_unique_tags': len(tag_counts),
            'total_tag_applications': sum(tag_counts.values()),
            'generated_at': datetime.now().isoformat()
        }

# Global auto-tagger instance
auto_tagger = AutoTagger() 