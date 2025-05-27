import { Outlet, Link } from 'react-router'
import { useAuthStore } from '../store/auth'

export default function Root() {
  const { isAuthenticated, user, logout } = useAuthStore()

  const handleLogout = () => {
    logout()
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Link to="/" className="text-xl font-bold text-gray-900 hover:text-blue-600">
                🤖 ML E-Commerce
              </Link>
            </div>
            <nav className="flex items-center space-x-8">
              <Link to="/" className="text-gray-700 hover:text-gray-900">
                Home
              </Link>
              <Link to="/products" className="text-gray-700 hover:text-gray-900">
                Products
              </Link>
              {isAuthenticated ? (
                <>
                  <Link to="/admin" className="text-gray-700 hover:text-gray-900">
                    Admin
                  </Link>
                  <div className="flex items-center space-x-4">
                    <span className="text-sm text-gray-600">
                      Welcome, {user?.name}
                    </span>
                    <button
                      onClick={handleLogout}
                      className="text-gray-700 hover:text-gray-900"
                    >
                      Logout
                    </button>
                  </div>
                </>
              ) : (
                <div className="flex items-center space-x-4">
                  <Link to="/auth/login" className="text-gray-700 hover:text-gray-900">
                    Login
                  </Link>
                  <Link 
                    to="/auth/register" 
                    className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                  >
                    Sign Up
                  </Link>
                </div>
              )}
            </nav>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main>
        <Outlet />
      </main>

      {/* Footer */}
      <footer className="bg-gray-800 text-white mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="text-center">
            <p>&copy; 2024 ML E-Commerce Platform. Built for Bachelor's Project.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}