import { Link } from 'react-router'

export default function Home() {
  return (
    <div className="space-y-12">
      {/* Hero Section */}
      <section className="bg-gradient-to-r from-blue-600 to-purple-600 text-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold mb-6">
              ML-Powered E-Commerce
            </h1>
            <p className="text-xl md:text-2xl mb-8 text-blue-100">
              Experience the future of online shopping with AI-driven recommendations,
              smart search, and personalized experiences.
            </p>
            <div className="space-x-4">
              <Link 
                to="/products"
                className="bg-white text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition-colors inline-block"
              >
                Shop Now
              </Link>
              <Link 
                to="/admin"
                className="border-2 border-white text-white px-8 py-3 rounded-lg font-semibold hover:bg-white hover:text-blue-600 transition-colors inline-block"
              >
                Admin Panel
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="text-center mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-4">
            Powered by Machine Learning
          </h2>
          <p className="text-lg text-gray-600">
            Our platform demonstrates cutting-edge ML algorithms in everyday e-commerce
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          <div className="text-center p-6 bg-white rounded-lg shadow-md">
            <div className="w-12 h-12 bg-blue-100 rounded-lg mx-auto mb-4 flex items-center justify-center">
              <span className="text-blue-600 text-xl">🎯</span>
            </div>
            <h3 className="text-lg font-semibold mb-2">Smart Recommendations</h3>
            <p className="text-gray-600">
              Personalized product suggestions using collaborative and content-based filtering
            </p>
          </div>

          <div className="text-center p-6 bg-white rounded-lg shadow-md">
            <div className="w-12 h-12 bg-green-100 rounded-lg mx-auto mb-4 flex items-center justify-center">
              <span className="text-green-600 text-xl">🔍</span>
            </div>
            <h3 className="text-lg font-semibold mb-2">Enhanced Search</h3>
            <p className="text-gray-600">
              ML-powered search with relevance ranking and intelligent suggestions
            </p>
          </div>

          <div className="text-center p-6 bg-white rounded-lg shadow-md">
            <div className="w-12 h-12 bg-purple-100 rounded-lg mx-auto mb-4 flex items-center justify-center">
              <span className="text-purple-600 text-xl">📈</span>
            </div>
            <h3 className="text-lg font-semibold mb-2">Trend Analysis</h3>
            <p className="text-gray-600">
              Real-time sales trends and forecasting with time series analysis
            </p>
          </div>

          <div className="text-center p-6 bg-white rounded-lg shadow-md">
            <div className="w-12 h-12 bg-orange-100 rounded-lg mx-auto mb-4 flex items-center justify-center">
              <span className="text-orange-600 text-xl">🤖</span>
            </div>
            <h3 className="text-lg font-semibold mb-2">AI Chatbot</h3>
            <p className="text-gray-600">
              Intelligent customer service with intent recognition and automated responses
            </p>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-gray-100">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <div className="text-center">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Ready to Experience ML-Powered Shopping?
            </h2>
            <p className="text-lg text-gray-600 mb-8">
              Join thousands of users discovering products through intelligent recommendations
            </p>
            <Link 
              to="/auth/register"
              className="bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-colors inline-block"
            >
              Get Started
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}
