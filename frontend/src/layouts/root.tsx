import { Outlet, Link } from 'react-router'
import { useState, useEffect } from 'react'
import { useAuthStore } from '../store/auth'
import { useCartStore } from '../store/cart'
import ShoppingCart from '../components/ShoppingCart'

export default function Root() {
  const { isAuthenticated, user, logout } = useAuthStore()
  const { cart, fetchCart } = useCartStore()
  const [isCartOpen, setIsCartOpen] = useState(false)

  useEffect(() => {
    if (isAuthenticated) {
      fetchCart()
    }
  }, [isAuthenticated, fetchCart])

  const handleLogout = () => {
    logout()
    setIsCartOpen(false)
  }

  const cartItemCount = cart?.total_items || 0

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
                  
                  {/* Cart Icon */}
                  <button
                    onClick={() => setIsCartOpen(true)}
                    className="relative text-gray-700 hover:text-gray-900 p-2"
                  >
                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4m0 0L7 13m0 0l-1.5 6M7 13l-1.5 6m0 0h9m-9 0h9" />
                    </svg>
                    {cartItemCount > 0 && (
                      <span className="absolute -top-1 -right-1 bg-blue-600 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center">
                        {cartItemCount > 99 ? '99+' : cartItemCount}
                      </span>
                    )}
                  </button>
                  
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

      {/* Shopping Cart Sidebar */}
      <ShoppingCart isOpen={isCartOpen} onClose={() => setIsCartOpen(false)} />

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