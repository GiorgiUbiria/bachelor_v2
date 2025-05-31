package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Email        string    `json:"email" gorm:"unique;not null;index"`
	Name         string    `json:"name" gorm:"not null;index"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	Orders           []Order           `json:"orders,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ShoppingCart     *ShoppingCart     `json:"shopping_cart,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserInteractions []UserInteraction `json:"user_interactions,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Recommendations  []Recommendation  `json:"recommendations,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Favorites        []Favorite        `json:"favorites,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Upvotes          []Upvote          `json:"upvotes,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments         []Comment         `json:"comments,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Product represents a product in the e-commerce platform
type Product struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"not null;index"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null;index"`
	Category    string    `json:"category" gorm:"not null;index"`
	Stock       int       `json:"stock" gorm:"default:0;index"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	OrderItems       []OrderItem       `json:"order_items,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CartItems        []CartItem        `json:"cart_items,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	UserInteractions []UserInteraction `json:"user_interactions,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Recommendations  []Recommendation  `json:"recommendations,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProductViews     []ProductView     `json:"product_views,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Favorites        []Favorite        `json:"favorites,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Upvotes          []Upvote          `json:"upvotes,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Comments         []Comment         `json:"comments,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Discounts        []Discount        `json:"discounts,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Tags             []Tag             `json:"tags,omitempty" gorm:"many2many:product_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Order represents an order placed by a user
type Order struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Total     float64   `json:"total" gorm:"type:decimal(10,2);not null;index"`
	Status    string    `json:"status" gorm:"default:'pending';index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	User       User        `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// OrderItem represents an item within an order
type OrderItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrderID   uuid.UUID `json:"order_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null;check:quantity > 0"`
	Price     float64   `json:"price" gorm:"type:decimal(10,2);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Relationships
	Order   Order   `json:"order" gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ShoppingCart represents a user's shopping cart
type ShoppingCart struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	User      User       `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CartItems []CartItem `json:"cart_items" gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CartID    uuid.UUID `json:"cart_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity  int       `json:"quantity" gorm:"not null;check:quantity > 0"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	Cart    ShoppingCart `json:"cart" gorm:"foreignKey:CartID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product      `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// Add unique constraint for cart_id + product_id combination
func (CartItem) TableName() string {
	return "cart_items"
}

// UserInteraction represents user interactions with products for ML
type UserInteraction struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID       uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	InteractionType string    `json:"interaction_type" gorm:"not null;index"` // 'view', 'cart_add', 'purchase', 'like'
	SessionID       string    `json:"session_id" gorm:"index"`
	CreatedAt       time.Time `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// SearchQuery represents search queries for analytics
type SearchQuery struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID         *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Query          string     `json:"query" gorm:"not null;index"`
	ResultsCount   int        `json:"results_count" gorm:"default:0"`
	ResultsClicked int        `json:"results_clicked" gorm:"default:0"`
	CreatedAt      time.Time  `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// Recommendation represents ML-generated recommendations
type Recommendation struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID     uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	AlgorithmType string    `json:"algorithm_type" gorm:"not null;index"` // 'collaborative', 'content_based', 'hybrid'
	Score         float64   `json:"score" gorm:"type:decimal(5,4);not null;index"`
	CreatedAt     time.Time `json:"created_at" gorm:"index"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// RecommendationFeedback represents user feedback on recommendations
type RecommendationFeedback struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID    uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	FeedbackType string    `json:"feedback_type" gorm:"not null;index"` // 'clicked', 'purchased', 'dismissed'
	CreatedAt    time.Time `json:"created_at" gorm:"index"`             // Changed from Timestamp to CreatedAt for consistency

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// UserSession represents user sessions for analytics
type UserSession struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	SessionID string     `json:"session_id" gorm:"not null;index"`
	IPAddress string     `json:"ip_address" gorm:"type:inet"`
	UserAgent string     `json:"user_agent"`
	StartedAt time.Time  `json:"started_at" gorm:"index"`
	EndedAt   *time.Time `json:"ended_at" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// ProductView represents product view events
type ProductView struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	ProductID uuid.UUID  `json:"product_id" gorm:"type:uuid;not null;index"`
	SessionID string     `json:"session_id" gorm:"index"`
	CreatedAt time.Time  `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency

	// Relationships
	User    *User   `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// SearchAnalytics represents search analytics data
type SearchAnalytics struct {
	ID               uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Query            string    `json:"query" gorm:"not null;index"`
	ResultsCount     int       `json:"results_count" gorm:"default:0"`
	ClickThroughRate float64   `json:"click_through_rate" gorm:"type:decimal(5,4);default:0"`
	CreatedAt        time.Time `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency
}

// MLModelPerformance represents ML model performance metrics
type MLModelPerformance struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ModelType      string    `json:"model_type" gorm:"not null;index"`
	Accuracy       float64   `json:"accuracy" gorm:"type:decimal(5,4)"`
	PrecisionScore float64   `json:"precision_score" gorm:"type:decimal(5,4)"`
	RecallScore    float64   `json:"recall_score" gorm:"type:decimal(5,4)"`
	F1Score        float64   `json:"f1_score" gorm:"type:decimal(5,4)"`
	CreatedAt      time.Time `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency
}

// Favorite represents user's favorite products
type Favorite struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Upvote represents user upvotes on products
type Upvote struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Comment represents user comments on products
type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Rating    int       `json:"rating" gorm:"check:rating >= 1 AND rating <= 5"` // 1-5 star rating
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at" gorm:"index"`

	// Relationships
	User    User    `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Discount represents product discounts
type Discount struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProductID         *uuid.UUID `json:"product_id" gorm:"type:uuid;index"`    // Nullable for category-wide discounts
	Category          *string    `json:"category" gorm:"index"`                // Nullable for product-specific discounts
	DiscountType      string     `json:"discount_type" gorm:"not null;index"`  // 'percentage', 'fixed_amount'
	DiscountValue     float64    `json:"discount_value" gorm:"not null"`       // Percentage (0-100) or fixed amount
	MinOrderAmount    float64    `json:"min_order_amount" gorm:"default:0"`    // Minimum order amount to apply discount
	MaxDiscountAmount float64    `json:"max_discount_amount" gorm:"default:0"` // Maximum discount amount (for percentage)
	StartDate         time.Time  `json:"start_date" gorm:"not null;index"`
	EndDate           time.Time  `json:"end_date" gorm:"not null;index"`
	IsActive          bool       `json:"is_active" gorm:"default:true;index"`
	UsageLimit        int        `json:"usage_limit" gorm:"default:0"` // 0 = unlimited
	UsageCount        int        `json:"usage_count" gorm:"default:0"`
	CreatedAt         time.Time  `json:"created_at" gorm:"index"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"index"`

	// Relationships
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// Tag represents product tags
type Tag struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;max:50"`
	Description string    `json:"description" gorm:"max:255"`
	Color       string    `json:"color" gorm:"max:7"` // Hex color code
	CreatedAt   time.Time `json:"created_at" gorm:"index"`

	// Relationships
	Products []Product `json:"products,omitempty" gorm:"many2many:product_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// ProductTag represents the many-to-many relationship between products and tags
type ProductTag struct {
	ProductID uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index"`
	TagID     uuid.UUID `json:"tag_id" gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Relationships
	Product Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Tag     Tag     `json:"tag" gorm:"foreignKey:TagID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// BeforeCreate hook for User model
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Product model
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
