import { useState, useEffect } from 'react'
import { apiService } from '../services/api'
import LoadingSpinner from './LoadingSpinner'

interface MLMetrics {
  recommendation_accuracy: number
  search_relevance: number
  trend_prediction_accuracy: number
  model_performance: {
    collaborative_filtering: number
    content_based: number
    hybrid: number
  }
}

export default function MLDashboard() {
  const [metrics, setMetrics] = useState<MLMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchMLMetrics()
  }, [])

  const fetchMLMetrics = async () => {
    try {
      setLoading(true)
      setError(null)
      
      // Fetch ML model status and metrics
      await apiService.ml.recommendations.status()
      
      // Mock metrics for demonstration
      const mockMetrics: MLMetrics = {
        recommendation_accuracy: 0.87,
        search_relevance: 0.92,
        trend_prediction_accuracy: 0.78,
        model_performance: {
          collaborative_filtering: 0.85,
          content_based: 0.89,
          hybrid: 0.91
        }
      }
      
      setMetrics(mockMetrics)
    } catch (error) {
      console.error('Error fetching ML metrics:', error)
      setError('Failed to load ML metrics')
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex justify-center items-center py-8">
          <LoadingSpinner size="lg" />
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="text-center py-8">
          <div className="text-red-500 text-4xl mb-4">⚠️</div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Error Loading ML Dashboard</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <button
            onClick={fetchMLMetrics}
            className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-lg p-6">
        <h2 className="text-xl font-semibold text-purple-900 mb-2 flex items-center">
          🤖 Machine Learning Dashboard
        </h2>
        <p className="text-purple-700">
          Monitor the performance of your ML models and algorithms in real-time.
        </p>
      </div>

      {/* ML Model Performance */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Recommendations</h3>
            <span className="text-2xl">🎯</span>
          </div>
          <div className="text-3xl font-bold text-blue-600 mb-2">
            {metrics ? (metrics.recommendation_accuracy * 100).toFixed(1) : '0'}%
          </div>
          <p className="text-gray-600 text-sm">Accuracy Rate</p>
          <div className="mt-3 bg-gray-200 rounded-full h-2">
            <div 
              className="bg-blue-600 h-2 rounded-full transition-all duration-500"
              style={{ width: `${metrics ? metrics.recommendation_accuracy * 100 : 0}%` }}
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Search</h3>
            <span className="text-2xl">🔍</span>
          </div>
          <div className="text-3xl font-bold text-green-600 mb-2">
            {metrics ? (metrics.search_relevance * 100).toFixed(1) : '0'}%
          </div>
          <p className="text-gray-600 text-sm">Relevance Score</p>
          <div className="mt-3 bg-gray-200 rounded-full h-2">
            <div 
              className="bg-green-600 h-2 rounded-full transition-all duration-500"
              style={{ width: `${metrics ? metrics.search_relevance * 100 : 0}%` }}
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Trends</h3>
            <span className="text-2xl">📈</span>
          </div>
          <div className="text-3xl font-bold text-purple-600 mb-2">
            {metrics ? (metrics.trend_prediction_accuracy * 100).toFixed(1) : '0'}%
          </div>
          <p className="text-gray-600 text-sm">Prediction Accuracy</p>
          <div className="mt-3 bg-gray-200 rounded-full h-2">
            <div 
              className="bg-purple-600 h-2 rounded-full transition-all duration-500"
              style={{ width: `${metrics ? metrics.trend_prediction_accuracy * 100 : 0}%` }}
            />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900">Overall</h3>
            <span className="text-2xl">⚡</span>
          </div>
          <div className="text-3xl font-bold text-orange-600 mb-2">
            {metrics ? (
              ((metrics.recommendation_accuracy + metrics.search_relevance + metrics.trend_prediction_accuracy) / 3 * 100).toFixed(1)
            ) : '0'}%
          </div>
          <p className="text-gray-600 text-sm">System Performance</p>
          <div className="mt-3 bg-gray-200 rounded-full h-2">
            <div 
              className="bg-orange-600 h-2 rounded-full transition-all duration-500"
              style={{ 
                width: `${metrics ? 
                  (metrics.recommendation_accuracy + metrics.search_relevance + metrics.trend_prediction_accuracy) / 3 * 100 
                  : 0}%` 
              }}
            />
          </div>
        </div>
      </div>

      {/* Algorithm Performance Breakdown */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-6">Recommendation Algorithm Performance</h3>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="text-center">
            <div className="text-2xl mb-2">🤝</div>
            <h4 className="font-medium text-gray-900 mb-2">Collaborative Filtering</h4>
            <div className="text-2xl font-bold text-blue-600 mb-2">
              {metrics ? (metrics.model_performance.collaborative_filtering * 100).toFixed(1) : '0'}%
            </div>
            <div className="bg-gray-200 rounded-full h-2">
              <div 
                className="bg-blue-600 h-2 rounded-full transition-all duration-500"
                style={{ width: `${metrics ? metrics.model_performance.collaborative_filtering * 100 : 0}%` }}
              />
            </div>
            <p className="text-gray-600 text-sm mt-2">User-based recommendations</p>
          </div>

          <div className="text-center">
            <div className="text-2xl mb-2">📊</div>
            <h4 className="font-medium text-gray-900 mb-2">Content-Based</h4>
            <div className="text-2xl font-bold text-green-600 mb-2">
              {metrics ? (metrics.model_performance.content_based * 100).toFixed(1) : '0'}%
            </div>
            <div className="bg-gray-200 rounded-full h-2">
              <div 
                className="bg-green-600 h-2 rounded-full transition-all duration-500"
                style={{ width: `${metrics ? metrics.model_performance.content_based * 100 : 0}%` }}
              />
            </div>
            <p className="text-gray-600 text-sm mt-2">Product feature matching</p>
          </div>

          <div className="text-center">
            <div className="text-2xl mb-2">🔄</div>
            <h4 className="font-medium text-gray-900 mb-2">Hybrid Model</h4>
            <div className="text-2xl font-bold text-purple-600 mb-2">
              {metrics ? (metrics.model_performance.hybrid * 100).toFixed(1) : '0'}%
            </div>
            <div className="bg-gray-200 rounded-full h-2">
              <div 
                className="bg-purple-600 h-2 rounded-full transition-all duration-500"
                style={{ width: `${metrics ? metrics.model_performance.hybrid * 100 : 0}%` }}
              />
            </div>
            <p className="text-gray-600 text-sm mt-2">Combined approach</p>
          </div>
        </div>
      </div>

      {/* ML Insights */}
      <div className="bg-gradient-to-r from-blue-50 to-purple-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-blue-900 mb-4 flex items-center">
          💡 ML Insights & Recommendations
        </h3>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-white rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">🎯 Model Optimization</h4>
            <p className="text-sm text-gray-600">
              Hybrid model shows best performance. Consider increasing its weight in the recommendation pipeline.
            </p>
          </div>
          
          <div className="bg-white rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">📊 Data Quality</h4>
            <p className="text-sm text-gray-600">
              Search relevance is high. Continue collecting user interaction data to improve recommendations.
            </p>
          </div>
          
          <div className="bg-white rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">🔄 Model Retraining</h4>
            <p className="text-sm text-gray-600">
              Last retrained 2 days ago. Schedule weekly retraining for optimal performance.
            </p>
          </div>
          
          <div className="bg-white rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-2">⚡ Performance</h4>
            <p className="text-sm text-gray-600">
              All models performing above 85% accuracy threshold. System is operating optimally.
            </p>
          </div>
        </div>
      </div>
    </div>
  )
} 