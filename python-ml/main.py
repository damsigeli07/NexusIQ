from fastapi import FastAPI

app = FastAPI(title="NexusIQ ML Service")

@app.get("/health")
def health():
    return {"status": "ok", "service": "python-ml"}

@app.post("/embed")
def embed_document(payload: dict):
    # Day 2: LangChain + pgvector ingestion pipeline goes here
    return {"message": "embedding pipeline coming Day 2"}

@app.post("/query")
def query(payload: dict):
    # Day 3: RAG query + LLM streaming goes here
    return {"message": "RAG query coming Day 3"}