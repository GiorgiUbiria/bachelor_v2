import { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router'
import { Search, Filter, X, TrendingUp, Clock, Sparkles } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import apiService from '../services/api'

interface SearchSuggestion {
  query: string
  type: 'product' | 'category' | 'trending'
  count?: number
}

interface SearchBarProps {
  placeholder?: string
  onSearch?: (query: string, filters?: SearchFilters) => void
  showFilters?: boolean
  className?: string
}

interface SearchFilters {
  category?: string
  minPrice?: number
  maxPrice?: number
  sortBy?: string
}

export function SearchBar({ 
  placeholder = "Search products with AI suggestions...", 
  onSearch,
  showFilters = true,
  className = ""
}: SearchBarProps) {
  const navigate = useNavigate()
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [recentSearches, setRecentSearches] = useState<string[]>([])
  const [trendingSearches, setTrendingSearches] = useState<string[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [filters, setFilters] = useState<SearchFilters>({})
  const [showFiltersPanel, setShowFiltersPanel] = useState(false)
  const searchRef = useRef<HTMLDivElement>(null)

  // Load recent searches from localStorage
  useEffect(() => {
    const saved = localStorage.getItem('recent-searches')
    if (saved) {
      try {
        setRecentSearches(JSON.parse(saved))
      } catch (error) {
        console.error('Error loading recent searches:', error)
      }
    }
  }, [])

  // Fetch trending searches and categories
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [trendingResponse, categoriesResponse] = await Promise.all([
          apiService.mlService.search.popular(5).catch(() => ({ data: { searches: [] } })),
          apiService.mlService.search.categories().catch(() => ({ data: { categories: [] } }))
        ])
        
        setTrendingSearches(trendingResponse.data.searches || [])
        setCategories(categoriesResponse.data.categories || [])
      } catch (error) {
        console.error('Error fetching search data:', error)
      }
    }
    
    fetchData()
  }, [])

  // Fetch suggestions based on query
  useEffect(() => {
    const fetchSuggestions = async () => {
      if (query.length < 2) {
        setSuggestions([])
        return
      }

      try {
        const response = await apiService.mlService.search.suggestions(query, 8)
        setSuggestions(response.data.suggestions || [])
      } catch (error) {
        console.error('Error fetching suggestions:', error)
        setSuggestions([])
      }
    }

    const debounceTimer = setTimeout(fetchSuggestions, 300)
    return () => clearTimeout(debounceTimer)
  }, [query])

  // Handle click outside to close suggestions
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setShowSuggestions(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const saveRecentSearch = (searchQuery: string) => {
    if (!searchQuery.trim()) return
    
    const updated = [searchQuery, ...recentSearches.filter(s => s !== searchQuery)].slice(0, 5)
    setRecentSearches(updated)
    localStorage.setItem('recent-searches', JSON.stringify(updated))
  }

  const handleSearch = (searchQuery: string = query) => {
    if (!searchQuery.trim()) return

    saveRecentSearch(searchQuery)
    setShowSuggestions(false)
    
    if (onSearch) {
      onSearch(searchQuery, filters)
    } else {
      // Navigate to products page with search params
      const params = new URLSearchParams()
      params.set('search', searchQuery)
      if (filters.category) params.set('category', filters.category)
      if (filters.minPrice) params.set('minPrice', filters.minPrice.toString())
      if (filters.maxPrice) params.set('maxPrice', filters.maxPrice.toString())
      if (filters.sortBy) params.set('sort', filters.sortBy)
      
      navigate(`/products?${params.toString()}`)
    }
  }

  const handleSuggestionClick = (suggestion: SearchSuggestion) => {
    setQuery(suggestion.query)
    handleSearch(suggestion.query)
  }

  const clearRecentSearches = () => {
    setRecentSearches([])
    localStorage.removeItem('recent-searches')
  }

  const getSuggestionIcon = (type: string) => {
    switch (type) {
      case 'trending':
        return <TrendingUp className="h-3 w-3 text-orange-500" />
      case 'category':
        return <Filter className="h-3 w-3 text-blue-500" />
      default:
        return <Search className="h-3 w-3 text-muted-foreground" />
    }
  }

  return (
    <div ref={searchRef} className={`relative ${className}`}>
      <div className="flex gap-2">
        {/* Main Search Input */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setShowSuggestions(true)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleSearch()
              }
            }}
            placeholder={placeholder}
            className="pl-10 pr-4"
          />
          {query && (
            <button
              onClick={() => setQuery('')}
              className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        {/* Search Button */}
        <Button onClick={() => handleSearch()} disabled={!query.trim()}>
          <Search className="h-4 w-4" />
        </Button>

        {/* Filters Toggle */}
        {showFilters && (
          <Popover open={showFiltersPanel} onOpenChange={setShowFiltersPanel}>
            <PopoverTrigger asChild>
              <Button variant="outline" size="icon">
                <Filter className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80" align="end">
              <div className="space-y-4">
                <h4 className="font-medium">Search Filters</h4>
                
                <div className="space-y-3">
                  <div>
                    <label className="text-sm font-medium">Category</label>
                    <Select 
                      value={filters.category || "all"} 
                      onValueChange={(value) => setFilters(prev => ({ ...prev, category: value === "all" ? undefined : value }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="All Categories" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All Categories</SelectItem>
                        {categories.map((category) => (
                          <SelectItem key={category} value={category}>
                            {category}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="text-sm font-medium">Min Price</label>
                      <Input
                        type="number"
                        placeholder="$0"
                        value={filters.minPrice || ''}
                        onChange={(e) => setFilters(prev => ({ 
                          ...prev, 
                          minPrice: e.target.value ? Number(e.target.value) : undefined 
                        }))}
                      />
                    </div>
                    <div>
                      <label className="text-sm font-medium">Max Price</label>
                      <Input
                        type="number"
                        placeholder="$999"
                        value={filters.maxPrice || ''}
                        onChange={(e) => setFilters(prev => ({ 
                          ...prev, 
                          maxPrice: e.target.value ? Number(e.target.value) : undefined 
                        }))}
                      />
                    </div>
                  </div>

                  <div>
                    <label className="text-sm font-medium">Sort By</label>
                    <Select 
                      value={filters.sortBy || "relevance"} 
                      onValueChange={(value) => setFilters(prev => ({ ...prev, sortBy: value === "relevance" ? undefined : value }))}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Relevance" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="relevance">Relevance</SelectItem>
                        <SelectItem value="price_asc">Price: Low to High</SelectItem>
                        <SelectItem value="price_desc">Price: High to Low</SelectItem>
                        <SelectItem value="rating">Rating</SelectItem>
                        <SelectItem value="newest">Newest</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="flex gap-2">
                  <Button 
                    onClick={() => handleSearch()}
                    className="flex-1"
                  >
                    Apply Filters
                  </Button>
                  <Button 
                    variant="outline"
                    onClick={() => setFilters({})}
                  >
                    Clear
                  </Button>
                </div>
              </div>
            </PopoverContent>
          </Popover>
        )}
      </div>

      {/* Search Suggestions Dropdown */}
      {showSuggestions && (
        <Card className="absolute top-full left-0 right-0 z-50 mt-1 max-h-96 overflow-y-auto">
          <CardContent className="p-0">
            {/* AI Suggestions */}
            {suggestions.length > 0 && (
              <div className="p-3 border-b">
                <div className="flex items-center gap-2 mb-2">
                  <Sparkles className="h-4 w-4 text-primary" />
                  <span className="text-sm font-medium">AI Suggestions</span>
                </div>
                <div className="space-y-1">
                  {suggestions.map((suggestion, index) => (
                    <button
                      key={index}
                      onClick={() => handleSuggestionClick(suggestion)}
                      className="w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-muted rounded-sm"
                    >
                      {getSuggestionIcon(suggestion.type)}
                      <span className="flex-1">{suggestion.query}</span>
                      {suggestion.count && (
                        <Badge variant="secondary" className="text-xs">
                          {suggestion.count}
                        </Badge>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Recent Searches */}
            {recentSearches.length > 0 && (
              <div className="p-3 border-b">
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Recent</span>
                  </div>
                  <button
                    onClick={clearRecentSearches}
                    className="text-xs text-muted-foreground hover:text-foreground"
                  >
                    Clear
                  </button>
                </div>
                <div className="space-y-1">
                  {recentSearches.map((search, index) => (
                    <button
                      key={index}
                      onClick={() => handleSuggestionClick({ query: search, type: 'product' })}
                      className="w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-muted rounded-sm"
                    >
                      <Clock className="h-3 w-3 text-muted-foreground" />
                      <span>{search}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Trending Searches */}
            {trendingSearches.length > 0 && (
              <div className="p-3">
                <div className="flex items-center gap-2 mb-2">
                  <TrendingUp className="h-4 w-4 text-orange-500" />
                  <span className="text-sm font-medium">Trending</span>
                </div>
                <div className="space-y-1">
                  {trendingSearches.map((search, index) => (
                    <button
                      key={index}
                      onClick={() => handleSuggestionClick({ query: search, type: 'trending' })}
                      className="w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-muted rounded-sm"
                    >
                      <TrendingUp className="h-3 w-3 text-orange-500" />
                      <span>{search}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* No suggestions */}
            {suggestions.length === 0 && recentSearches.length === 0 && trendingSearches.length === 0 && (
              <div className="p-6 text-center text-muted-foreground">
                <Search className="h-8 w-8 mx-auto mb-2" />
                <p>Start typing to see suggestions</p>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}