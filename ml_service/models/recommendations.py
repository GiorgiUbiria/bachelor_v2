import pandas as pd
import numpy as np
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity
from sklearn.decomposition import TruncatedSVD
from typing import List, Dict, Tuple
import uuid
from datetime import datetime

from database.connection import (
    get_user_interactions, 
    get_products_data, 
    get_user_product_matrix,
    execute_query
)

class RecommendationEngine:
    def __init__(self):
        self.user_item_matrix = None
        self.product_features = None
        self.tfidf_vectorizer = None
        self.content_similarity_matrix = None
        self.svd_model = None
        self.is_trained = False
    
    async def train_models(self):
        """Train both collaborative and content-based models"""
        print("🤖 Training recommendation models...")
        
        # Load data
        interactions_df = await get_user_interactions()
        products_df = await get_products_data()
        
        if interactions_df.empty or products_df.empty:
            print("⚠️ No data available for training")
            return
        
        # Train collaborative filtering
        await self._train_collaborative_filtering(interactions_df)
        
        # Train content-based filtering
        await self._train_content_based_filtering(products_df)
        
        self.is_trained = True
        print("✅ Recommendation models trained successfully")
    
    async def _train_collaborative_filtering(self, interactions_df: pd.DataFrame):
        """Train collaborative filtering model using SVD"""
        try:
            # Create user-item matrix
            matrix_df = await get_user_product_matrix()
            
            if matrix_df.empty:
                print("⚠️ No interaction data for collaborative filtering")
                return
            
            # Pivot to create user-item matrix
            self.user_item_matrix = matrix_df.pivot_table(
                index='user_id', 
                columns='product_id', 
                values='max_score', 
                fill_value=0
            )
            
            # Apply SVD for dimensionality reduction
            if self.user_item_matrix.shape[0] > 1 and self.user_item_matrix.shape[1] > 1:
                n_components = min(50, min(self.user_item_matrix.shape) - 1)
                self.svd_model = TruncatedSVD(n_components=n_components, random_state=42)
                self.svd_model.fit(self.user_item_matrix)
                print(f"✅ SVD model trained with {n_components} components")
            
        except Exception as e:
            print(f"❌ Error training collaborative filtering: {e}")
    
    async def _train_content_based_filtering(self, products_df: pd.DataFrame):
        """Train content-based filtering model using TF-IDF"""
        try:
            # Combine text features
            products_df['combined_features'] = (
                products_df['name'].fillna('') + ' ' + 
                products_df['description'].fillna('') + ' ' + 
                products_df['category'].fillna('')
            )
            
            # Create TF-IDF vectors
            self.tfidf_vectorizer = TfidfVectorizer(
                max_features=1000,
                stop_words='english',
                ngram_range=(1, 2)
            )
            
            tfidf_matrix = self.tfidf_vectorizer.fit_transform(products_df['combined_features'])
            
            # Calculate cosine similarity
            self.content_similarity_matrix = cosine_similarity(tfidf_matrix)
            
            # Store product features
            self.product_features = products_df.set_index('id')
            
            print("✅ Content-based model trained successfully")
            
        except Exception as e:
            print(f"❌ Error training content-based filtering: {e}")
    
    async def get_collaborative_recommendations(self, user_id: str, n_recommendations: int = 10) -> List[Dict]:
        """Get recommendations using collaborative filtering"""
        if not self.is_trained or self.user_item_matrix is None:
            return []
        
        try:
            user_uuid = uuid.UUID(user_id)
            
            if user_uuid not in self.user_item_matrix.index:
                # Cold start problem - return popular items
                return await self._get_popular_items(n_recommendations)
            
            # Get user's ratings
            user_ratings = self.user_item_matrix.loc[user_uuid]
            
            # Find similar users using SVD if available
            if self.svd_model is not None:
                user_vector = self.svd_model.transform([user_ratings])
                all_users_vectors = self.svd_model.transform(self.user_item_matrix)
                
                # Calculate similarities
                similarities = cosine_similarity(user_vector, all_users_vectors)[0]
                similar_users_indices = np.argsort(similarities)[::-1][1:11]  # Top 10 similar users
                
                # Get recommendations from similar users
                recommendations = {}
                for idx in similar_users_indices:
                    similar_user_id = self.user_item_matrix.index[idx]
                    similar_user_ratings = self.user_item_matrix.loc[similar_user_id]
                    
                    # Find items the similar user liked but current user hasn't interacted with
                    for product_id, rating in similar_user_ratings.items():
                        if rating > 0 and user_ratings[product_id] == 0:
                            if product_id not in recommendations:
                                recommendations[product_id] = 0
                            recommendations[product_id] += rating * similarities[idx]
                
                # Sort and return top recommendations
                sorted_recommendations = sorted(
                    recommendations.items(), 
                    key=lambda x: x[1], 
                    reverse=True
                )[:n_recommendations]
                
                return [
                    {
                        'product_id': str(product_id),
                        'score': float(score),
                        'algorithm': 'collaborative'
                    }
                    for product_id, score in sorted_recommendations
                ]
            
        except Exception as e:
            print(f"❌ Error in collaborative filtering: {e}")
        
        return []
    
    async def get_content_based_recommendations(self, user_id: str, n_recommendations: int = 10) -> List[Dict]:
        """Get recommendations using content-based filtering"""
        if not self.is_trained or self.content_similarity_matrix is None:
            return []
        
        try:
            # Get user's interaction history
            interactions_df = await get_user_interactions()
            user_interactions = interactions_df[
                interactions_df['user_id'] == uuid.UUID(user_id)
            ]
            
            if user_interactions.empty:
                return await self._get_popular_items(n_recommendations)
            
            # Get products the user has interacted with
            user_products = user_interactions['product_id'].unique()
            
            # Calculate content-based scores
            recommendations = {}
            
            for product_id in user_products:
                if product_id in self.product_features.index:
                    product_idx = list(self.product_features.index).index(product_id)
                    similarities = self.content_similarity_matrix[product_idx]
                    
                    for idx, similarity in enumerate(similarities):
                        similar_product_id = self.product_features.index[idx]
                        
                        # Don't recommend products user has already interacted with
                        if similar_product_id not in user_products:
                            if similar_product_id not in recommendations:
                                recommendations[similar_product_id] = 0
                            recommendations[similar_product_id] += similarity
            
            # Sort and return top recommendations
            sorted_recommendations = sorted(
                recommendations.items(), 
                key=lambda x: x[1], 
                reverse=True
            )[:n_recommendations]
            
            return [
                {
                    'product_id': str(product_id),
                    'score': float(score),
                    'algorithm': 'content_based'
                }
                for product_id, score in sorted_recommendations
            ]
            
        except Exception as e:
            print(f"❌ Error in content-based filtering: {e}")
        
        return []
    
    async def get_hybrid_recommendations(self, user_id: str, n_recommendations: int = 10) -> List[Dict]:
        """Get hybrid recommendations combining collaborative and content-based"""
        # Get recommendations from both methods
        collaborative_recs = await self.get_collaborative_recommendations(user_id, n_recommendations)
        content_recs = await self.get_content_based_recommendations(user_id, n_recommendations)
        
        # Combine and weight the recommendations
        combined_recommendations = {}
        
        # Weight collaborative filtering more heavily
        for rec in collaborative_recs:
            product_id = rec['product_id']
            combined_recommendations[product_id] = rec['score'] * 0.7
        
        # Add content-based recommendations
        for rec in content_recs:
            product_id = rec['product_id']
            if product_id in combined_recommendations:
                combined_recommendations[product_id] += rec['score'] * 0.3
            else:
                combined_recommendations[product_id] = rec['score'] * 0.3
        
        # Sort and return top recommendations
        sorted_recommendations = sorted(
            combined_recommendations.items(), 
            key=lambda x: x[1], 
            reverse=True
        )[:n_recommendations]
        
        return [
            {
                'product_id': product_id,
                'score': float(score),
                'algorithm': 'hybrid'
            }
            for product_id, score in sorted_recommendations
        ]
    
    async def _get_popular_items(self, n_recommendations: int = 10) -> List[Dict]:
        """Get popular items for cold start users"""
        try:
            interactions_df = await get_user_interactions()
            
            if interactions_df.empty:
                return []
            
            # Count interactions per product
            popular_products = interactions_df.groupby('product_id').size().sort_values(ascending=False)
            
            return [
                {
                    'product_id': str(product_id),
                    'score': float(count),
                    'algorithm': 'popular'
                }
                for product_id, count in popular_products.head(n_recommendations).items()
            ]
            
        except Exception as e:
            print(f"❌ Error getting popular items: {e}")
            return []
    
    async def save_recommendations_to_db(self, user_id: str, recommendations: List[Dict]):
        """Save recommendations to database"""
        try:
            # Clear existing recommendations for user
            await execute_query(
                "DELETE FROM recommendations WHERE user_id = $1",
                [uuid.UUID(user_id)]
            )
            
            # Insert new recommendations
            for rec in recommendations:
                await execute_query(
                    """
                    INSERT INTO recommendations (user_id, product_id, algorithm_type, score)
                    VALUES ($1, $2, $3, $4)
                    """,
                    [
                        uuid.UUID(user_id),
                        uuid.UUID(rec['product_id']),
                        rec['algorithm'],
                        rec['score']
                    ]
                )
            
            print(f"✅ Saved {len(recommendations)} recommendations for user {user_id}")
            
        except Exception as e:
            print(f"❌ Error saving recommendations: {e}")

# Global recommendation engine instance
recommendation_engine = RecommendationEngine() 