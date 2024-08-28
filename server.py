from transformers import GPTNeoForCausalLM, GPT2Tokenizer
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

app = FastAPI()

# Load model and tokenizer
model_name = "EleutherAI/gpt-neo-1.3B"
model = GPTNeoForCausalLM.from_pretrained(model_name)
tokenizer = GPT2Tokenizer.from_pretrained(model_name)

class SummaryRequest(BaseModel):
    text: str

class SummaryResponse(BaseModel):
    summary: str

@app.post("/summarize", response_model=SummaryResponse)
async def summarize(request: SummaryRequest):
    try:
        inputs = tokenizer(request.text, return_tensors="pt")
        outputs = model.generate(inputs.input_ids, max_length=1024, num_beams=5, early_stopping=True)
        summary = tokenizer.decode(outputs[0], skip_special_tokens=True)
        return SummaryResponse(summary=summary)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
