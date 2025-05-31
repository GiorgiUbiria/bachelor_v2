import { Outlet } from 'react-router'
import { Navbar } from '../components/navbar'
import { Footer } from '../components/footer'

export default function RootLayout() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />

      <main className="flex-1">
        <Outlet />
      </main>

      <Footer />
    </div>
  )
}