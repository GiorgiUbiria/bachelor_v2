import { Outlet } from 'react-router'

export default function Root() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold text-gray-900">
                ML E-Commerce
              </h1>
            </div>
            <nav className="flex space-x-8">
              <a href="/" className="text-gray-700 hover:text-gray-900">
                Home
              </a>
              <a href="/products" className="text-gray-700 hover:text-gray-900">
                Products
              </a>
              <a href="/auth/login" className="text-gray-700 hover:text-gray-900">
                Login
              </a>
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