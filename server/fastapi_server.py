from fastapi import FastAPI, HTTPException, Request
import openai

# Initialize FastAPI app
app = FastAPI()

# Configure OpenAI API key (replace "your_openai_api_key_here" with your actual API key)
openai.api_key = "your_openai_api_key_here"

@app.post("/generate-text/")
async def generate_text(request: Request):
    data = await request.json()
    prompt = data.get("prompt")
    max_tokens = data.get("max_tokens", 150)

    if not prompt:
        raise HTTPException(status_code=400, detail="Prompt is required")

    try:
        response = openai.Completion.create(
            engine="davinci",  # You can use other engines like "curie", "babbage", etc.
            prompt=prompt,
            max_tokens=max_tokens
        )
        return response.choices[0].text.strip()
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)