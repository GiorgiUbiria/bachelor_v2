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

// ChatSession represents a chat session with the bot
type ChatSession struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	SessionID string     `json:"session_id" gorm:"uniqueIndex;not null"`
	StartedAt time.Time  `json:"started_at" gorm:"index"`
	EndedAt   *time.Time `json:"ended_at" gorm:"index"`

	// Relationships
	User         *User         `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	ChatMessages []ChatMessage `json:"chat_messages" gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	SessionID  uuid.UUID `json:"session_id" gorm:"type:uuid;not null;index"`
	Message    string    `json:"message" gorm:"not null"`
	SenderType string    `json:"sender_type" gorm:"not null;index"` // 'user', 'bot'
	Intent     string    `json:"intent" gorm:"index"`
	CreatedAt  time.Time `json:"created_at" gorm:"index"` // Changed from Timestamp to CreatedAt for consistency

	// Relationships
	Session ChatSession `json:"session" gorm:"foreignKey:SessionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// ChatbotIntent represents chatbot intents and responses
type ChatbotIntent struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	IntentName      string    `json:"intent_name" gorm:"uniqueIndex;not null"`
	TrainingPhrases []string  `json:"training_phrases" gorm:"type:text[]"`
	Responses       []string  `json:"responses" gorm:"type:text[]"`
	CreatedAt       time.Time `json:"created_at" gorm:"index"`
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
