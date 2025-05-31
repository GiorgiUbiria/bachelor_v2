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
        """Get recommendations using collaborative filtering with detailed reasoning"""
        if not self.is_trained or self.user_item_matrix is None:
            return []
        
        try:
            user_uuid = uuid.UUID(user_id)
            
            if user_uuid not in self.user_item_matrix.index:
                # Cold start problem - return popular items with reasoning
                popular_items = await self._get_popular_items(n_recommendations)
                for item in popular_items:
                    item['reasoning'] = {
                        'method': 'Cold Start Fallback',
                        'explanation': 'No interaction history found for user, showing popular items',
                        'confidence': 'Low',
                        'factors': ['New user', 'Popular among other users', 'General appeal'],
                        'technical_details': {
                            'algorithm': 'Popularity-based ranking',
                            'data_source': 'Global interaction counts',
                            'fallback_reason': 'Insufficient user data for collaborative filtering'
                        }
                    }
                return popular_items
            
            # Get user's ratings
            user_ratings = self.user_item_matrix.loc[user_uuid]
            user_interaction_count = (user_ratings > 0).sum()
            
            # Find similar users using SVD if available
            if self.svd_model is not None:
                user_vector = self.svd_model.transform([user_ratings])
                all_users_vectors = self.svd_model.transform(self.user_item_matrix)
                
                # Calculate similarities
                similarities = cosine_similarity(user_vector, all_users_vectors)[0]
                similar_users_indices = np.argsort(similarities)[::-1][1:11]  # Top 10 similar users
                
                # Get recommendations from similar users
                recommendations = {}
                user_similarity_scores = {}
                contributing_users = {}
                
                for idx in similar_users_indices:
                    similar_user_id = self.user_item_matrix.index[idx]
                    similar_user_ratings = self.user_item_matrix.loc[similar_user_id]
                    similarity_score = similarities[idx]
                    
                    # Find items the similar user liked but current user hasn't interacted with
                    for product_id, rating in similar_user_ratings.items():
                        if rating > 0 and user_ratings[product_id] == 0:
                            if product_id not in recommendations:
                                recommendations[product_id] = 0
                                contributing_users[product_id] = []
                            
                            weighted_score = rating * similarity_score
                            recommendations[product_id] += weighted_score
                            contributing_users[product_id].append({
                                'user_similarity': float(similarity_score),
                                'user_rating': float(rating),
                                'contribution': float(weighted_score)
                            })
                
                # Sort and return top recommendations with detailed reasoning
                sorted_recommendations = sorted(
                    recommendations.items(), 
                    key=lambda x: x[1], 
                    reverse=True
                )[:n_recommendations]
                
                result = []
                for product_id, score in sorted_recommendations:
                    # Calculate confidence based on number of similar users and their similarity scores
                    contributors = contributing_users[product_id]
                    avg_similarity = np.mean([c['user_similarity'] for c in contributors])
                    num_contributors = len(contributors)
                    
                    confidence = 'High' if avg_similarity > 0.7 and num_contributors >= 3 else \
                                'Medium' if avg_similarity > 0.5 and num_contributors >= 2 else 'Low'
                    
                    # Generate detailed reasoning
                    reasoning = {
                        'method': 'Collaborative Filtering (SVD-based)',
                        'explanation': f'Users with similar preferences (similarity: {avg_similarity:.2f}) also liked this item',
                        'confidence': confidence,
                        'factors': [
                            f'{num_contributors} similar users liked this item',
                            f'Average user similarity: {avg_similarity:.2f}',
                            f'Your interaction history: {user_interaction_count} items',
                            f'Recommendation score: {score:.3f}'
                        ],
                        'technical_details': {
                            'algorithm': 'Truncated SVD + Cosine Similarity',
                            'svd_components': self.svd_model.n_components,
                            'similar_users_analyzed': len(similar_users_indices),
                            'contributing_users': num_contributors,
                            'weighted_score_calculation': 'sum(user_rating * user_similarity)',
                            'data_sparsity': f'{(self.user_item_matrix > 0).sum().sum()} / {self.user_item_matrix.size} interactions'
                        },
                        'contributing_users': contributors[:3]  # Top 3 contributors for transparency
                    }
                    
                    result.append({
                        'product_id': str(product_id),
                        'score': float(score),
                        'algorithm': 'collaborative',
                        'reasoning': reasoning
                    })
                
                return result
            
        except Exception as e:
            print(f"❌ Error in collaborative filtering: {e}")
        
        return []
    
    async def get_content_based_recommendations(self, user_id: str, n_recommendations: int = 10) -> List[Dict]:
        """Get recommendations using content-based filtering with detailed reasoning"""
        if not self.is_trained or self.content_similarity_matrix is None:
            return []
        
        try:
            # Get user's interaction history
            interactions_df = await get_user_interactions()
            user_interactions = interactions_df[
                interactions_df['user_id'] == uuid.UUID(user_id)
            ]
            
            if user_interactions.empty:
                popular_items = await self._get_popular_items(n_recommendations)
                for item in popular_items:
                    item['reasoning'] = {
                        'method': 'Content-Based Fallback',
                        'explanation': 'No interaction history found, showing popular items',
                        'confidence': 'Low',
                        'factors': ['New user', 'Popular content', 'General appeal'],
                        'technical_details': {
                            'algorithm': 'Popularity-based ranking',
                            'fallback_reason': 'No user interaction data for content analysis'
                        }
                    }
                return popular_items
            
            # Get products the user has interacted with
            user_products = user_interactions['product_id'].unique()
            user_preferences = {}
            
            # Analyze user's content preferences
            for _, interaction in user_interactions.iterrows():
                product_id = interaction['product_id']
                interaction_type = interaction['interaction_type']
                
                # Weight different interaction types
                weight = {'view': 1, 'cart_add': 2, 'purchase': 3, 'like': 2}.get(interaction_type, 1)
                
                if product_id not in user_preferences:
                    user_preferences[product_id] = 0
                user_preferences[product_id] += weight
            
            # Calculate content-based scores
            recommendations = {}
            content_explanations = {}
            
            for product_id, preference_score in user_preferences.items():
                if product_id in self.product_features.index:
                    product_idx = list(self.product_features.index).index(product_id)
                    similarities = self.content_similarity_matrix[product_idx]
                    
                    # Get product details for reasoning
                    source_product = self.product_features.loc[product_id]
                    
                    for idx, similarity in enumerate(similarities):
                        similar_product_id = self.product_features.index[idx]
                        
                        # Don't recommend products user has already interacted with
                        if similar_product_id not in user_products and similarity > 0.1:
                            if similar_product_id not in recommendations:
                                recommendations[similar_product_id] = 0
                                content_explanations[similar_product_id] = {
                                    'similar_to': [],
                                    'shared_features': [],
                                    'total_similarity': 0,
                                    'preference_weight': 0
                                }
                            
                            weighted_score = similarity * preference_score
                            recommendations[similar_product_id] += weighted_score
                            
                            # Track reasoning details
                            similar_product = self.product_features.loc[similar_product_id]
                            content_explanations[similar_product_id]['similar_to'].append({
                                'product_name': source_product.get('name', 'Unknown'),
                                'product_category': source_product.get('category', 'Unknown'),
                                'similarity_score': float(similarity),
                                'your_interaction': preference_score,
                                'contribution': float(weighted_score)
                            })
                            
                            # Identify shared features
                            shared_features = []
                            if source_product.get('category') == similar_product.get('category'):
                                shared_features.append(f"Same category: {source_product.get('category')}")
                            
                            # Analyze text similarity (simplified)
                            source_text = f"{source_product.get('name', '')} {source_product.get('description', '')}"
                            similar_text = f"{similar_product.get('name', '')} {similar_product.get('description', '')}"
                            
                            if len(set(source_text.lower().split()) & set(similar_text.lower().split())) > 2:
                                shared_features.append("Similar keywords in description")
                            
                            content_explanations[similar_product_id]['shared_features'].extend(shared_features)
                            content_explanations[similar_product_id]['total_similarity'] += similarity
                            content_explanations[similar_product_id]['preference_weight'] += preference_score
            
            # Sort and return top recommendations with detailed reasoning
            sorted_recommendations = sorted(
                recommendations.items(), 
                key=lambda x: x[1], 
                reverse=True
            )[:n_recommendations]
            
            result = []
            for product_id, score in sorted_recommendations:
                explanation = content_explanations[product_id]
                num_similar_items = len(explanation['similar_to'])
                avg_similarity = explanation['total_similarity'] / num_similar_items if num_similar_items > 0 else 0
                
                # Determine confidence
                confidence = 'High' if avg_similarity > 0.6 and num_similar_items >= 2 else \
                            'Medium' if avg_similarity > 0.4 and num_similar_items >= 1 else 'Low'
                
                # Get unique shared features
                unique_features = list(set(explanation['shared_features']))
                
                reasoning = {
                    'method': 'Content-Based Filtering (TF-IDF)',
                    'explanation': f'Similar to {num_similar_items} items you\'ve interacted with (avg similarity: {avg_similarity:.2f})',
                    'confidence': confidence,
                    'factors': [
                        f'Similar to {num_similar_items} of your preferred items',
                        f'Average content similarity: {avg_similarity:.2f}',
                        f'Your preference strength: {explanation["preference_weight"]:.1f}',
                        f'Recommendation score: {score:.3f}'
                    ] + unique_features[:3],  # Add top 3 shared features
                    'technical_details': {
                        'algorithm': 'TF-IDF Vectorization + Cosine Similarity',
                        'features_analyzed': 'Product name, description, category',
                        'similarity_threshold': 0.1,
                        'interaction_weights': {'view': 1, 'cart_add': 2, 'purchase': 3, 'like': 2},
                        'tfidf_features': self.tfidf_vectorizer.get_feature_names_out().shape[0] if self.tfidf_vectorizer else 0
                    },
                    'similar_items': explanation['similar_to'][:3]  # Top 3 similar items for transparency
                }
                
                result.append({
                    'product_id': str(product_id),
                    'score': float(score),
                    'algorithm': 'content_based',
                    'reasoning': reasoning
                })
            
            return result
            
        except Exception as e:
            print(f"❌ Error in content-based filtering: {e}")
        
        return []
    
    async def get_hybrid_recommendations(self, user_id: str, n_recommendations: int = 10) -> List[Dict]:
        """Get hybrid recommendations combining collaborative and content-based with detailed reasoning"""
        # Get recommendations from both methods
        collaborative_recs = await self.get_collaborative_recommendations(user_id, n_recommendations)
        content_recs = await self.get_content_based_recommendations(user_id, n_recommendations)
        
        # Combine and weight the recommendations
        combined_recommendations = {}
        reasoning_details = {}
        
        # Weight collaborative filtering more heavily (70%)
        for rec in collaborative_recs:
            product_id = rec['product_id']
            collaborative_score = rec['score'] * 0.7
            combined_recommendations[product_id] = collaborative_score
            
            reasoning_details[product_id] = {
                'collaborative': {
                    'score': rec['score'],
                    'weighted_score': collaborative_score,
                    'reasoning': rec.get('reasoning', {}),
                    'weight': 0.7
                },
                'content_based': None
            }
        
        # Add content-based recommendations (30%)
        for rec in content_recs:
            product_id = rec['product_id']
            content_score = rec['score'] * 0.3
            
            if product_id in combined_recommendations:
                combined_recommendations[product_id] += content_score
                reasoning_details[product_id]['content_based'] = {
                    'score': rec['score'],
                    'weighted_score': content_score,
                    'reasoning': rec.get('reasoning', {}),
                    'weight': 0.3
                }
            else:
                combined_recommendations[product_id] = content_score
                reasoning_details[product_id] = {
                    'collaborative': None,
                    'content_based': {
                        'score': rec['score'],
                        'weighted_score': content_score,
                        'reasoning': rec.get('reasoning', {}),
                        'weight': 0.3
                    }
                }
        
        # Sort and return top recommendations
        sorted_recommendations = sorted(
            combined_recommendations.items(), 
            key=lambda x: x[1], 
            reverse=True
        )[:n_recommendations]
        
        result = []
        for product_id, final_score in sorted_recommendations:
            details = reasoning_details[product_id]
            
            # Determine which algorithms contributed
            contributing_algorithms = []
            algorithm_explanations = []
            confidence_scores = []
            
            if details['collaborative']:
                contributing_algorithms.append('Collaborative Filtering')
                collab_reasoning = details['collaborative']['reasoning']
                algorithm_explanations.append(f"Collaborative: {collab_reasoning.get('explanation', 'Similar users liked this')}")
                confidence_scores.append(collab_reasoning.get('confidence', 'Medium'))
            
            if details['content_based']:
                contributing_algorithms.append('Content-Based Filtering')
                content_reasoning = details['content_based']['reasoning']
                algorithm_explanations.append(f"Content: {content_reasoning.get('explanation', 'Similar to your preferences')}")
                confidence_scores.append(content_reasoning.get('confidence', 'Medium'))
            
            # Calculate overall confidence
            high_confidence_count = confidence_scores.count('High')
            medium_confidence_count = confidence_scores.count('Medium')
            
            if high_confidence_count >= 1 and len(contributing_algorithms) >= 2:
                overall_confidence = 'High'
            elif high_confidence_count >= 1 or (medium_confidence_count >= 1 and len(contributing_algorithms) >= 2):
                overall_confidence = 'Medium'
            else:
                overall_confidence = 'Low'
            
            # Create comprehensive reasoning
            reasoning = {
                'method': 'Hybrid Recommendation (Collaborative + Content-Based)',
                'explanation': f'Combines {" and ".join(contributing_algorithms)} for personalized recommendations',
                'confidence': overall_confidence,
                'factors': [
                    f'Final score: {final_score:.3f}',
                    f'Algorithms used: {", ".join(contributing_algorithms)}',
                    f'Collaborative weight: 70%, Content weight: 30%'
                ] + algorithm_explanations,
                'technical_details': {
                    'algorithm': 'Weighted Hybrid (Linear Combination)',
                    'collaborative_weight': 0.7,
                    'content_based_weight': 0.3,
                    'combination_method': 'Linear weighted sum',
                    'contributing_algorithms': len(contributing_algorithms)
                },
                'algorithm_breakdown': {
                    'collaborative': details['collaborative'],
                    'content_based': details['content_based']
                }
            }
            
            result.append({
                'product_id': product_id,
                'score': float(final_score),
                'algorithm': 'hybrid',
                'reasoning': reasoning
            })
        
        return result
    
    async def _get_popular_items(self, n_recommendations: int = 10) -> List[Dict]:
        """Get popular items for cold start users with detailed reasoning"""
        try:
            interactions_df = await get_user_interactions()
            
            if interactions_df.empty:
                return []
            
            # Count interactions per product with different weights
            interaction_weights = {'view': 1, 'cart_add': 3, 'purchase': 5, 'like': 2}
            
            # Calculate weighted popularity scores
            popularity_scores = {}
            interaction_details = {}
            
            for _, interaction in interactions_df.iterrows():
                product_id = interaction['product_id']
                interaction_type = interaction['interaction_type']
                weight = interaction_weights.get(interaction_type, 1)
                
                if product_id not in popularity_scores:
                    popularity_scores[product_id] = 0
                    interaction_details[product_id] = {
                        'total_interactions': 0,
                        'unique_users': set(),
                        'interaction_breakdown': {}
                    }
                
                popularity_scores[product_id] += weight
                interaction_details[product_id]['total_interactions'] += 1
                interaction_details[product_id]['unique_users'].add(interaction['user_id'])
                
                if interaction_type not in interaction_details[product_id]['interaction_breakdown']:
                    interaction_details[product_id]['interaction_breakdown'][interaction_type] = 0
                interaction_details[product_id]['interaction_breakdown'][interaction_type] += 1
            
            # Sort by popularity score
            sorted_popular = sorted(popularity_scores.items(), key=lambda x: x[1], reverse=True)
            
            result = []
            for i, (product_id, score) in enumerate(sorted_popular[:n_recommendations]):
                details = interaction_details[product_id]
                unique_users_count = len(details['unique_users'])
                total_interactions = details['total_interactions']
                
                # Determine confidence based on interaction diversity and user count
                interaction_types = len(details['interaction_breakdown'])
                confidence = 'High' if unique_users_count >= 10 and interaction_types >= 3 else \
                            'Medium' if unique_users_count >= 5 and interaction_types >= 2 else 'Low'
                
                # Create detailed reasoning
                top_interaction_type = max(details['interaction_breakdown'].items(), key=lambda x: x[1])
                
                reasoning = {
                    'method': 'Popularity-Based Ranking',
                    'explanation': f'Popular among {unique_users_count} users with {total_interactions} total interactions',
                    'confidence': confidence,
                    'factors': [
                        f'Popularity rank: #{i+1}',
                        f'Weighted popularity score: {score:.1f}',
                        f'Unique users: {unique_users_count}',
                        f'Total interactions: {total_interactions}',
                        f'Most common interaction: {top_interaction_type[0]} ({top_interaction_type[1]} times)'
                    ],
                    'technical_details': {
                        'algorithm': 'Weighted interaction counting',
                        'interaction_weights': interaction_weights,
                        'ranking_method': 'Descending by weighted score',
                        'interaction_diversity': interaction_types
                    },
                    'interaction_breakdown': details['interaction_breakdown']
                }
                
                result.append({
                    'product_id': str(product_id),
                    'score': float(score),
                    'algorithm': 'popular',
                    'reasoning': reasoning
                })
            
            return result
            
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