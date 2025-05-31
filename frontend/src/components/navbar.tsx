import { Link, useNavigate, useLocation } from 'react-router'
import { useState, useEffect } from 'react'
import { ShoppingBag, LogOut, Menu } from 'lucide-react'
import { useAuthStore } from '../store/auth'
import { useCartStore } from '../store/cart'
import { useUIStore } from '../store/ui'
import { ThemeToggle } from './theme-toggle'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet'

export function Navbar() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, isAuthenticated, logout } = useAuthStore()
  const { cart, fetchCart } = useCartStore()
  const { addToast } = useUIStore()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  useEffect(() => {
    if (isAuthenticated) {
      fetchCart()
      
      if (location.pathname === '/login' || location.pathname === '/register') {
        navigate('/')
      }
    }
  }, [isAuthenticated, fetchCart, location.pathname, navigate])

  const handleLogout = async () => {
    try {
      await logout()
      addToast({
        type: 'success',
        description: 'Logged out successfully'
      })
      navigate('/')
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to logout'
      })
    }
    setMobileMenuOpen(false)
  }

  const cartItemCount = cart?.total_items || 0

  const navigationItems = [
    { name: 'Home', href: '/', icon: '🏠' },
    { name: 'Products', href: '/products', icon: '📦' },
  ]

  const authenticatedItems = [
    { name: 'Admin', href: '/admin', icon: '⚙️' },
  ]

  return (
    <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between">
        <Link to="/" className="flex items-center space-x-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            🤖
          </div>
          <span className="hidden font-bold sm:inline-block">
            ML E-Commerce
          </span>
        </Link>

        <nav className="hidden md:flex items-center space-x-6 text-sm font-medium">
          {navigationItems.map((item) => (
            <Link
              key={item.name}
              to={item.href}
              className="transition-colors hover:text-foreground/80 text-foreground/60"
            >
              <span className="mr-2">{item.icon}</span>
              {item.name}
            </Link>
          ))}
          {isAuthenticated && authenticatedItems.map((item) => (
            <Link
              key={item.name}
              to={item.href}
              className="transition-colors hover:text-foreground/80 text-foreground/60"
            >
              <span className="mr-2">{item.icon}</span>
              {item.name}
            </Link>
          ))}
        </nav>

        <div className="flex items-center space-x-2">
          <ThemeToggle />
          {isAuthenticated && (
            <Button
              variant="outline"
              size="icon"
              className="relative"
            >
              <ShoppingBag className="h-4 w-4" />
              {cartItemCount > 0 && (
                <Badge
                  variant="destructive"
                  className="absolute -top-2 -right-2 h-5 w-5 rounded-full p-0 text-xs"
                >
                  {cartItemCount > 99 ? '99+' : cartItemCount}
                </Badge>
              )}
            </Button>
          )}

          {isAuthenticated ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>
                      {user?.name?.charAt(0).toUpperCase() || 'U'}
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount>
                <div className="flex items-center justify-start gap-2 p-2">
                  <div className="flex flex-col space-y-1 leading-none">
                    <p className="font-medium">{user?.name}</p>
                    <p className="w-[200px] truncate text-sm text-muted-foreground">
                      {user?.email}
                    </p>
                  </div>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Log out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className="hidden md:flex items-center space-x-2">
              <Button variant="ghost" asChild>
                <Link to="/auth/login">Login</Link>
              </Button>
              <Button asChild>
                <Link to="/auth/register">Sign Up</Link>
              </Button>
            </div>
          )}

          <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
            <SheetTrigger asChild>
              <Button
                variant="ghost"
                className="md:hidden"
                size="icon"
              >
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle Menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="pr-0">
              <div className="px-6">
                <Link
                  to="/"
                  className="flex items-center space-x-2"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <div className="flex h-6 w-6 items-center justify-center rounded-lg bg-primary text-primary-foreground text-sm">
                    🤖
                  </div>
                  <span className="font-bold">ML E-Commerce</span>
                </Link>
                <div className="my-4 h-[calc(100vh-8rem)] pb-10 pl-6">
                  <div className="flex flex-col space-y-3">
                    {navigationItems.map((item) => (
                      <Link
                        key={item.name}
                        to={item.href}
                        onClick={() => setMobileMenuOpen(false)}
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <span className="mr-2">{item.icon}</span>
                        {item.name}
                      </Link>
                    ))}
                    {isAuthenticated && authenticatedItems.map((item) => (
                      <Link
                        key={item.name}
                        to={item.href}
                        onClick={() => setMobileMenuOpen(false)}
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <span className="mr-2">{item.icon}</span>
                        {item.name}
                      </Link>
                    ))}
                    {!isAuthenticated && (
                      <>
                        <Link
                          to="/auth/login"
                          onClick={() => setMobileMenuOpen(false)}
                          className="text-muted-foreground hover:text-foreground"
                        >
                          Login
                        </Link>
                        <Link
                          to="/auth/register"
                          onClick={() => setMobileMenuOpen(false)}
                          className="text-muted-foreground hover:text-foreground"
                        >
                          Sign Up
                        </Link>
                      </>
                    )}
                  </div>
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </header>
  )
}
