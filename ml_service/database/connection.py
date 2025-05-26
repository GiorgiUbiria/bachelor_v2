import os
import asyncpg
from sqlalchemy import create_engine, MetaData
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
import pandas as pd

# Database configuration
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = os.getenv("DB_PORT", "5432")
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASSWORD = os.getenv("DB_PASSWORD", "postgres123")
DB_NAME = os.getenv("DB_NAME", "bachelor_db")

# SQLAlchemy setup
DATABASE_URL = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
engine = create_engine(DATABASE_URL)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
Base = declarative_base()

# AsyncPG connection pool
connection_pool = None

async def init_db():
    """Initialize database connection pool"""
    global connection_pool
    connection_pool = await asyncpg.create_pool(
        host=DB_HOST,
        port=DB_PORT,
        user=DB_USER,
        password=DB_PASSWORD,
        database=DB_NAME,
        min_size=1,
        max_size=10
    )
    print("✅ Database connection pool initialized")

async def get_db_connection():
    """Get database connection from pool"""
    if connection_pool is None:
        await init_db()
    return await connection_pool.acquire()

async def release_db_connection(connection):
    """Release database connection back to pool"""
    await connection_pool.release(connection)

def get_db_session():
    """Get SQLAlchemy database session"""
    db = SessionLocal()
    try:
        return db
    finally:
        db.close()

async def fetch_dataframe(query: str, params: list = None) -> pd.DataFrame:
    """Execute query and return results as pandas DataFrame"""
    connection = await get_db_connection()
    try:
        if params:
            result = await connection.fetch(query, *params)
        else:
            result = await connection.fetch(query)
        
        # Convert to DataFrame
        if result:
            columns = list(result[0].keys())
            data = [list(row.values()) for row in result]
            return pd.DataFrame(data, columns=columns)
        else:
            return pd.DataFrame()
    finally:
        await release_db_connection(connection)

async def execute_query(query: str, params: list = None):
    """Execute query without returning results"""
    connection = await get_db_connection()
    try:
        if params:
            await connection.execute(query, *params)
        else:
            await connection.execute(query)
    finally:
        await release_db_connection(connection)

# Utility functions for ML data retrieval

async def get_user_interactions() -> pd.DataFrame:
    """Get all user interactions for ML training"""
    query = """
    SELECT 
        ui.user_id,
        ui.product_id,
        ui.interaction_type,
        ui.timestamp,
        p.category,
        p.price
    FROM user_interactions ui
    JOIN products p ON ui.product_id = p.id
    ORDER BY ui.timestamp DESC
    """
    return await fetch_dataframe(query)

async def get_products_data() -> pd.DataFrame:
    """Get products data for content-based filtering"""
    query = """
    SELECT 
        id,
        name,
        description,
        price,
        category,
        stock,
        created_at
    FROM products
    """
    return await fetch_dataframe(query)

async def get_user_product_matrix() -> pd.DataFrame:
    """Get user-product interaction matrix"""
    query = """
    SELECT 
        user_id,
        product_id,
        COUNT(*) as interaction_count,
        MAX(CASE WHEN interaction_type = 'purchase' THEN 5
                WHEN interaction_type = 'cart_add' THEN 3
                WHEN interaction_type = 'like' THEN 2
                WHEN interaction_type = 'view' THEN 1
                ELSE 0 END) as max_score
    FROM user_interactions
    GROUP BY user_id, product_id
    """
    return await fetch_dataframe(query)

async def get_search_queries() -> pd.DataFrame:
    """Get search queries for search enhancement"""
    query = """
    SELECT 
        query,
        results_count,
        results_clicked,
        timestamp
    FROM search_queries
    WHERE query IS NOT NULL AND query != ''
    ORDER BY timestamp DESC
    """
    return await fetch_dataframe(query)

async def get_sales_data() -> pd.DataFrame:
    """Get sales data for trend analysis"""
    query = """
    SELECT 
        DATE(o.created_at) as date,
        p.category,
        p.id as product_id,
        p.name as product_name,
        oi.quantity,
        oi.price,
        (oi.quantity * oi.price) as total_amount
    FROM orders o
    JOIN order_items oi ON o.id = oi.order_id
    JOIN products p ON oi.product_id = p.id
    WHERE o.status = 'completed'
    ORDER BY o.created_at DESC
    """
    return await fetch_dataframe(query) 