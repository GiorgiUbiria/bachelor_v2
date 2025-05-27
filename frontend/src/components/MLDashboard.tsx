import { useState, useEffect } from 'react'
import { apiService } from '../services/api'
import LoadingSpinner from './LoadingSpinner'

interface MLStatus {
  is_trained: boolean
  has_user_item_matrix: boolean
  has_content_similarity: boolean
  has_svd_model: boolean
}

export default function MLDashboard() {
  const [status, setStatus] = useState<MLStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [training, setTraining] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)

  useEffect(() => {
    fetchMLStatus()
  }, [])

  const fetchMLStatus = async () => {
    try {
      setLoading(true)
      const response = await apiService.ml.recommendations.getUser('status')
      setStatus(response.data)
    } catch (error) {
      console.error('Error fetching ML status:', error)
      setMessage({ type: 'error', text: 'Failed to fetch ML service status' })
    } finally {
      setLoading(false)
    }
  }

  const handleTrainModels = async () => {
    try {
      setTraining(true)
      setMessage(null)
      
      // Call the train endpoint
      await fetch('/api/v1/ml/train', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${JSON.parse(localStorage.getItem('auth-storage') || '{}').state?.token}`,
          'Content-Type': 'application/json'
        }
      })
      
      setMessage({ type: 'success', text: 'ML model training initiated successfully!' })
      
      // Refresh status after a delay
      setTimeout(() => {
        fetchMLStatus()
      }, 2000)
      
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to train ML models' })
    } finally {
      setTraining(false)
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

  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-gray-900 flex items-center">
          🤖 ML Service Dashboard
        </h2>
        <button
          onClick={fetchMLStatus}
          className="text-blue-600 hover:text-blue-700 text-sm"
        >
          Refresh Status
        </button>
      </div>

      {message && (
        <div className={`mb-4 p-3 rounded-md ${
          message.type === 'success' 
            ? 'bg-green-50 border border-green-200 text-green-700' 
            : 'bg-red-50 border border-red-200 text-red-700'
        }`}>
          {message.text}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className={`w-3 h-3 rounded-full mr-2 ${
              status?.is_trained ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            <span className="text-sm font-medium text-gray-700">Models Trained</span>
          </div>
          <p className="text-xs text-gray-500 mt-1">
            {status?.is_trained ? 'Ready' : 'Not trained'}
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className={`w-3 h-3 rounded-full mr-2 ${
              status?.has_user_item_matrix ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            <span className="text-sm font-medium text-gray-700">User Matrix</span>
          </div>
          <p className="text-xs text-gray-500 mt-1">
            {status?.has_user_item_matrix ? 'Available' : 'Not available'}
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className={`w-3 h-3 rounded-full mr-2 ${
              status?.has_content_similarity ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            <span className="text-sm font-medium text-gray-700">Content Similarity</span>
          </div>
          <p className="text-xs text-gray-500 mt-1">
            {status?.has_content_similarity ? 'Available' : 'Not available'}
          </p>
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center">
            <div className={`w-3 h-3 rounded-full mr-2 ${
              status?.has_svd_model ? 'bg-green-500' : 'bg-red-500'
            }`}></div>
            <span className="text-sm font-medium text-gray-700">SVD Model</span>
          </div>
          <p className="text-xs text-gray-500 mt-1">
            {status?.has_svd_model ? 'Available' : 'Not available'}
          </p>
        </div>
      </div>

      <div className="border-t pt-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-medium text-gray-900">Model Training</h3>
            <p className="text-sm text-gray-600">
              Train or retrain the recommendation models with latest data
            </p>
          </div>
          <button
            onClick={handleTrainModels}
            disabled={training}
            className="bg-purple-600 text-white px-6 py-2 rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
          >
            {training && <LoadingSpinner size="sm" className="mr-2" />}
            {training ? 'Training...' : 'Train Models'}
          </button>
        </div>
      </div>

      <div className="mt-4 p-4 bg-blue-50 rounded-lg">
        <h4 className="text-sm font-medium text-blue-900 mb-2">How it works:</h4>
        <ul className="text-xs text-blue-700 space-y-1">
          <li>• <strong>Collaborative Filtering:</strong> Recommends products based on similar users' preferences</li>
          <li>• <strong>Content-Based Filtering:</strong> Recommends products similar to ones users have interacted with</li>
          <li>• <strong>Hybrid Approach:</strong> Combines both methods for better recommendations</li>
          <li>• <strong>SVD Model:</strong> Uses matrix factorization for dimensionality reduction</li>
        </ul>
      </div>
    </div>
  )
} 