from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI(title="Adaptive Quiz ML Service", version="1.0.0")

class UserHistoryItem(BaseModel):
    question_id: int
    is_correct: bool
    difficulty: float # 0.0 to 1.0

class RecommendationRequest(BaseModel):
    user_id: int
    quiz_id: int
    history: List[UserHistoryItem]
    prior_ability: Optional[float] = 0.5 # Default to medium if not provided

class RecommendationResponse(BaseModel):
    target_difficulty: float
    predicted_ability: float
    message: str

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "ml-service"}

@app.post("/recommend", response_model=RecommendationResponse)
def recommend_difficulty(request: RecommendationRequest):
    # 1. Use Prior Ability as Anchor
    anchor_ability = request.prior_ability if request.prior_ability is not None else 0.5

    if not request.history:
        # New user/start of quiz -> Start at Anchor
        return {
            "target_difficulty": anchor_ability,
            "predicted_ability": anchor_ability,
            "message": f"Starting at difficulty {anchor_ability}"
        }
    
    # 2. Calculate ability score (simple running average of 'performance')
    # Performance on a question = (1 if correct else 0) * difficulty_weight
    
    # Phantom Prior (Bayesian Smoothing) using Anchor
    # Assume 1 item at anchor difficulty was correct? Or just weight it?
    # Let's say we trust the anchor with weight 2.0
    total_score = anchor_ability * 2.0
    total_weight = 2.0
    
    for item in request.history:
        weight = 1 + item.difficulty # Higher difficulty = higher weight
        if item.is_correct:
            total_score += 1 * item.difficulty # Higher diff correct = more score
        else:
             total_score += 0 # Incorrect adds 0
    
        total_weight += 1 # Normalize weight count? 
        # Actually better heuristic:
        # Score = Sum(Diff if Correct) / TotalQuestions + Phantom
        
    # Re-evaluating Heuristic:
    # We want Ability to track Difficulty.
    # Let's stick to the previous simple logic but inject Prior.
    
    # Previous Logic:
    # total_score = 1.5 (Prior)
    # total_weight = 1.5
    
    # New Logic with Prior:
    total_score = anchor_ability * 2.0
    total_weight = 2.0

    for item in request.history:
        weight = 1.0 + item.difficulty
        if item.is_correct:
             total_score += 1.0 * weight
        # else 0
        total_weight += weight
        
    ability = 0.5
    if total_weight > 0:
        ability = total_score / total_weight
        
    # Normalize ability
    ability = max(0.0, min(1.0, ability))
    
    last_item = request.history[-1]
    
    # Current difficulty pointer
    next_difficulty = ability
    
    # DEBUG LOG
    print(f"DEBUG: LastQ Diff={last_item.difficulty}, Correct={last_item.is_correct}, Ability={ability}")

    if last_item.is_correct:
        # Increase difficulty
        step = 0.1
        next_val = last_item.difficulty + step
        
        # Pull up logic (Damped)
        if ability > next_val:
             # (Target * 3 + Ability) / 4 -> Moderate pull
             next_val = (next_val * 3 + ability) / 4
        
        next_difficulty = min(1.0, next_val)
        msg = "Correct! Increasing difficulty."
    else:
        # Decrease difficulty
        step = 0.05 # Smaller drop step
        next_val = last_item.difficulty - step
        
        # Only pull down if ability is lower, Damped
        if ability < next_val:
             next_val = (next_val * 4 + ability) / 5
             
        # MAX DROP GUARD: Don't drop more than 0.15 from last difficulty
        if (last_item.difficulty - next_val) > 0.15:
            next_val = last_item.difficulty - 0.15

        # MIN FLOOR: 0.2
        next_difficulty = max(0.2, next_val)
        msg = "Incorrect. Decreasing difficulty."

    # Return the recommendation
    return {
        "target_difficulty": round(next_difficulty, 2),
        "predicted_ability": round(ability, 2),
        "message": msg
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5002)
