from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Optional

router = APIRouter()

class ChatMessage(BaseModel):
    message: str
    session_id: Optional[str] = None
    user_id: Optional[str] = None

class ChatResponse(BaseModel):
    response: str
    intent: Optional[str] = None
    confidence: Optional[float] = None
    session_id: str

@router.post("/message", response_model=ChatResponse)
async def chat_message(chat: ChatMessage):
    """Process chat message and return bot response"""
    return ChatResponse(
        response="Hello! I'm a chatbot assistant. This feature is coming soon!",
        intent="greeting",
        confidence=0.95,
        session_id=chat.session_id or "default_session"
    )

@router.get("/intents")
async def get_available_intents():
    """Get list of available chatbot intents"""
    return {
        "intents": [
            "greeting",
            "product_inquiry", 
            "order_status",
            "return_policy"
        ]
    }

@router.post("/train")
async def train_chatbot():
    """Train/retrain chatbot models"""
    return {
        "message": "Chatbot training endpoint - coming soon"
    } 