Authentication

User authentication and profile management
POST
/auth/login
User login
GET
/auth/profile
Get user profile
PUT
/auth/profile
Update user profile
POST
/auth/register
Register a new user
Products

Product catalog management and search
GET
/products
Get products
POST
/products
Create a new product
GET
/products/recommendations
Get product recommendations
GET
/products/search
Search products
GET
/products/{id}
Get product by ID
PUT
/products/{id}
Update a product
DELETE
/products/{id}
Delete a product
Cart

Shopping cart operations
GET
/cart
Get shopping cart
POST
/cart/add
Add item to cart
DELETE
/cart/clear
Clear shopping cart
PUT
/cart/item/{id}
Update cart item quantity
DELETE
/cart/item/{id}
Remove item from cart
Orders

Order management and tracking
GET
/orders
Get user orders
POST
/orders
Create order from cart
GET
/orders/stats
Get order statistics
GET
/orders/{id}
Get order by ID
PUT
/orders/{id}/cancel
Cancel order
PUT
/orders/{id}/status
Update order status
Analytics

Business analytics and reporting
GET
/analytics/dashboard
Get dashboard analytics
GET
/analytics/export
Export analytics data
GET
/analytics/products
Get product analytics
GET
/analytics/recommendations/metrics
Get recommendation metrics
GET
/analytics/search
Get search analytics
GET
/analytics/trends
Get ML trends
GET
/analytics/user
Get user analytics
ML

Machine Learning services and model management
POST
/ml/auto-tagging/auto-tag
Auto-tag products
GET
/ml/auto-tagging/insights
Get tagging insights
GET
/ml/auto-tagging/suggest/{id}
Suggest product tags
POST
/ml/initialize-services
Initialize ML services
GET
/ml/sentiment/category/{category}
Analyze category sentiment
GET
/ml/sentiment/insights
Get sentiment insights
GET
/ml/sentiment/product/{id}
Analyze product sentiment
GET
/ml/smart-discounts/insights
Get discount insights
GET
/ml/smart-discounts/suggest/category/{category}
Suggest category discounts
GET
/ml/smart-discounts/suggest/product/{id}
Suggest product discount
Comments
POST
/comments
Add product comment
PUT
/comments/{comment_id}
Update comment
DELETE
/comments/{comment_id}
Delete comment
GET
/comments/{product_id}
Get product comments
Discounts
POST
/discounts
Create discount
GET
/discounts/active
Get active discounts
Favorites
GET
/favorites
Get user favorites
POST
/favorites
Add product to favorites
DELETE
/favorites/{product_id}
Remove product from favorites
Tags
GET
/tags
Get all tags
POST
/tags
Create tag
POST
/tags/products
Add tag to product
GET
/tags/products/{product_id}
Get product tags
Upvotes
POST
/upvotes
Upvote product
GET
/upvotes/{product_id}
Get product upvotes
DELETE
/upvotes/{product_id}
Remove upvote
Models
handlers.AddCommentRequest{
content*	[...]
product_id*	[...]
rating*	[...]
}
handlers.AddFavoriteRequest{
product_id*	[...]
}
handlers.AddProductTagRequest{
product_id*	[...]
tag_id*	[...]
}
handlers.AddToCartRequest{
product_id*	[...]
quantity*	[...]
}
handlers.AddUpvoteRequest{
product_id*	[...]
}
handlers.AuthResponse{
token	string
example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
user	models.User{
comments	[...]
created_at	[...]
email	[...]
favorites	[...]
id	[...]
name	[...]
orders	[...]
recommendations	[...]
shopping_cart	models.ShoppingCart{...}
updated_at	[...]
upvotes	[...]
user_interactions	[...]
}
}
handlers.CreateDiscountRequest{
category	[...]
discount_type*	[...]
discount_value*	[...]
end_date*	[...]
max_discount_amount	[...]
min_order_amount	[...]
product_id	[...]
start_date*	[...]
usage_limit	[...]
}
handlers.CreateOrderRequest{
cart_item_ids	[...]
payment_method*	[...]
shipping_address*	[...]
}
handlers.CreateProductRequest{
category*	[...]
description*	[...]
image_url	[...]
name*	[...]
price*	[...]
stock*	[...]
}
handlers.CreateTagRequest{
color	[...]
description	[...]
name*	[...]
}
handlers.LoginRequest{
email*	[...]
password*	[...]
}
handlers.RegisterRequest{
email*	string
maxLength: 255
example: user@example.com
name*	string
maxLength: 100
minLength: 2
example: John Doe
password*	string
maxLength: 128
minLength: 8
example: password123
}
handlers.StandardErrorResponse{
error	string
example: Invalid email or password
success	boolean
example: false
}
handlers.UpdateCartItemRequest{
quantity*	integer
maximum: 100
minimum: 0
example: 3
}
handlers.UpdateCommentRequest{
content	string
maxLength: 1000
minLength: 1
example: Updated comment
rating	integer
maximum: 5
minimum: 1
example: 4
}
handlers.UpdateOrderStatusRequest{
status*	string
example: processingEnum:
[ pending, processing, shipped, delivered, cancelled ]
}
handlers.UpdateProductRequest{
category	string
maxLength: 100
minLength: 1
example: Electronics
description	string
maxLength: 1000
minLength: 1
example: Latest iPhone with A17 Pro chip
image_url	string
example: https://example.com/image.jpg
name	string
maxLength: 255
minLength: 1
example: iPhone 15 Pro
price	number
minimum: 0.01
example: 999.99
stock	integer
minimum: 0
example: 50
}
handlers.UpdateProfileRequest{
name	string
maxLength: 100
minLength: 2
example: John Doe
phone	string
maxLength: 20
example: +1234567890
}
handlers.UserProfileResponse{
comments	[models.Comment{
content	string
created_at	string
id	string
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	string
rating	integer

1-5 star rating
updated_at	string
user	{
description:	

Relationships
}
user_id	string
}]
created_at	string
email	string
favorites	[models.Favorite{
created_at	string
id	string
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	string
user	{
description:	

Relationships
}
user_id	string
}]
id	string
name	string
orders	[

Relationships
models.Order{
created_at	string
id	string
order_items	[http://localhost:8081/swagger/doc.json#/definitions/models.OrderItem]
status	string
total	number
updated_at	string
user	{
description:	

Relationships
}
user_id	string
}]
recent_activity	handlers.UserRecentActivity{
recent_interactions	[models.UserInteraction{
created_at	[...]
id	[...]
interaction_type	[...]
product	models.Product{...}
product_id	[...]
session_id	[...]
user	{...}
user_id	[...]
}]
recent_orders	[models.Order{
created_at	[...]
id	[...]
order_items	[...]
status	[...]
total	[...]
updated_at	[...]
user	{...}
user_id	[...]
}]
recent_searches	[models.SearchQuery{
created_at	[...]
id	[...]
query	[...]
results_clicked	[...]
results_count	[...]
user	{...}
user_id	[...]
}]
recent_views	[models.ProductView{
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
session_id	[...]
user	{...}
user_id	[...]
}]
}
recommendations	[models.Recommendation{
algorithm_type	[...]
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
score	[...]
user	{...}
user_id	[...]
}]
shopping_cart	models.ShoppingCart{
cart_items	[...]
created_at	[...]
id	[...]
updated_at	[...]
user	{...}
user_id	[...]
}
statistics	handlers.UserProfileStatistics{
average_order_value	[...]
cart_items_count	[...]
cart_total_value	[...]
favorite_category	[...]
recommendations_clicked	[...]
recommendations_received	[...]
total_interactions	[...]
total_orders	[...]
total_spent	[...]
}
updated_at	string
upvotes	[models.Upvote{
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
user	{...}
user_id	[...]
}]
user_interactions	[models.UserInteraction{
created_at	[...]
id	[...]
interaction_type	[...]
product	models.Product{...}
product_id	[...]
session_id	[...]
user	{...}
user_id	[...]
}]
}
handlers.UserProfileStatistics{
average_order_value	[...]
cart_items_count	[...]
cart_total_value	[...]
favorite_category	[...]
recommendations_clicked	[...]
recommendations_received	[...]
total_interactions	[...]
total_orders	[...]
total_spent	[...]
}
handlers.UserRecentActivity{
recent_interactions	[...]
recent_orders	[...]
recent_searches	[...]
recent_views	[...]
}
models.CartItem{
cart	{...}
cart_id	[...]
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
quantity	[...]
updated_at	[...]
}
models.Comment{
content	[...]
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
rating	[...]
updated_at	[...]
user	{...}
user_id	[...]
}
models.Discount{
category	[...]
created_at	[...]
discount_type	[...]
discount_value	[...]
end_date	[...]
id	[...]
is_active	[...]
max_discount_amount	[...]
min_order_amount	[...]
product	{...}
product_id	[...]
start_date	[...]
updated_at	[...]
usage_count	[...]
usage_limit	[...]
}
models.Favorite{
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
user	{...}
user_id	[...]
}
models.Order{
created_at	[...]
id	[...]
order_items	[...]
status	[...]
total	[...]
updated_at	[...]
user	{...}
user_id	[...]
}
models.OrderItem{
created_at	[...]
id	[...]
order	{...}
order_id	[...]
price	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
quantity	[...]
}
models.Product{
cart_items	[...]
category	[...]
comments	[...]
created_at	[...]
description	[...]
discounts	[...]
favorites	[...]
id	[...]
image_url	[...]
name	[...]
order_items	[...]
price	[...]
product_views	[...]
recommendations	[...]
stock	[...]
tags	[...]
updated_at	[...]
upvotes	[...]
user_interactions	[...]
}
models.ProductView{
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
session_id	[...]
user	{...}
user_id	[...]
}
models.Recommendation{
algorithm_type	[...]
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
score	[...]
user	{...}
user_id	[...]
}
models.SearchQuery{
created_at	[...]
id	[...]
query	[...]
results_clicked	[...]
results_count	[...]
user	{...}
user_id	[...]
}
models.ShoppingCart{
cart_items	[...]
created_at	[...]
id	[...]
updated_at	[...]
user	{...}
user_id	[...]
}
models.Tag{
color	[...]
created_at	[...]
description	[...]
id	[...]
name	[...]
products	[...]
}
models.Upvote{
created_at	[...]
id	[...]
product	http://localhost:8081/swagger/doc.json#/definitions/models.Product
product_id	[...]
user	{...}
user_id	[...]
}
models.User{
comments	[...]
created_at	[...]
email	[...]
favorites	[...]
id	[...]
name	[...]
orders	[...]
recommendations	[...]
shopping_cart	models.ShoppingCart{...}
updated_at	[...]
upvotes	[...]
user_interactions	[...]
}
models.UserInteraction{
created_at	[...]
id	[...]
interaction_type	[...]
product	models.Product{...}
product_id	[...]
session_id	[...]
user	{...}
user_id	[...]
}
