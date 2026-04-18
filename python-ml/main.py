import os
import uuid
import json
from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
import psycopg2
from psycopg2.extras import execute_values
from langchain_community.document_loaders import PyPDFLoader, TextLoader, Docx2txtLoader
from langchain.schema import Document
from langchain.text_splitter import RecursiveCharacterTextSplitter
import google.generativeai as genai

app = FastAPI(title="NexusIQ ML Service")

# Gemini setup
genai.configure(api_key=os.getenv("GEMINI_API_KEY"))


# ── DB connection ────────────────────────────────────────────────────────────

def get_db():
    return psycopg2.connect(
        host="postgres",
        database=os.getenv("POSTGRES_DB", "nexusiq"),
        user=os.getenv("POSTGRES_USER", "nexusiq"),
        password=os.getenv("POSTGRES_PASSWORD", "nexusiq_secret"),
    )


# ── Request models ───────────────────────────────────────────────────────────

class IngestRequest(BaseModel):
    document_id: str
    tenant_id:   str
    file_path:   str
    source_type: str

class QueryRequest(BaseModel):
    question:  str
    tenant_id: str

class StreamRequest(BaseModel):
    question:  str
    tenant_id: str


# ── Health ───────────────────────────────────────────────────────────────────

@app.get("/health")
def health():
    return {"status": "ok", "service": "python-ml"}


# ── Ingestion pipeline ───────────────────────────────────────────────────────

@app.post("/embed")
def embed_document(req: IngestRequest):
    """
    Full ingestion pipeline:
    load file → split into chunks → embed each chunk → bulk insert into pgvector
    """
    db  = get_db()
    cur = db.cursor()

    try:
        # 1. Mark document as processing
        cur.execute(
            "UPDATE documents SET status = 'processing' WHERE id = %s",
            (req.document_id,)
        )
        db.commit()

        # 2. Load document with the right LangChain loader
        loader = _get_loader(req.file_path, req.source_type)
        pages  = loader.load()

        # 3. Split into chunks (512 tokens, 50 overlap)
        splitter = RecursiveCharacterTextSplitter(
            chunk_size=512,
            chunk_overlap=50,
            separators=["\n\n", "\n", ". ", " ", ""],
        )
        chunks = splitter.split_documents(pages)

        if not chunks:
            raise ValueError("No text extracted from document")

        # 4. Embed each chunk via Gemini gemini-embedding-001 (3072 dims)
        records = []
        for i, chunk in enumerate(chunks):
            text = chunk.page_content.strip()
            if not text:
                continue

            result = genai.embed_content(
                model="models/gemini-embedding-001",
                content=text,
                task_type="retrieval_document",
            )
            embedding = result["embedding"]  # list of 1536 floats

            records.append((
                str(uuid.uuid4()),
                req.document_id,
                req.tenant_id,
                text,
                embedding,
                i,
            ))

        # 5. Bulk insert into pgvector
        execute_values(
            cur,
            """INSERT INTO chunks (id, document_id, tenant_id, content, embedding, chunk_index)
               VALUES %s""",
            records,
            template="(%s, %s, %s, %s, %s::vector, %s)",
        )

        # 6. Mark document as ready
        cur.execute(
            "UPDATE documents SET status = 'ready' WHERE id = %s",
            (req.document_id,)
        )
        db.commit()

        return {"status": "ready", "chunks_stored": len(records)}

    except Exception as e:
        db.rollback()
        cur.execute(
            "UPDATE documents SET status = 'failed' WHERE id = %s",
            (req.document_id,)
        )
        db.commit()
        raise HTTPException(status_code=500, detail=str(e))

    finally:
        cur.close()
        db.close()


# ── Similarity search (no LLM — raw chunks) ─────────────────────────────────

@app.post("/query")
def query_rag(req: QueryRequest):
    """
    Embed question → cosine similarity search → return top-5 chunks.
    Used for testing retrieval quality without LLM generation.
    """
    db  = get_db()
    cur = db.cursor()

    try:
        result = genai.embed_content(
            model="models/gemini-embedding-001",
            content=req.question,
            task_type="retrieval_query",
        )
        q_embedding = result["embedding"]

        cur.execute(
            """SELECT id, content, document_id,
                      1 - (embedding <=> %s::vector) AS similarity
               FROM chunks
               WHERE tenant_id = %s
               ORDER BY embedding <=> %s::vector
               LIMIT 5""",
            (q_embedding, req.tenant_id, q_embedding)
        )
        rows = cur.fetchall()

        return {
            "question": req.question,
            "chunks": [
                {
                    "chunk_id":    str(r[0]),
                    "content":     r[1],
                    "document_id": str(r[2]),
                    "similarity":  round(float(r[3]), 4),
                }
                for r in rows
            ],
        }

    finally:
        cur.close()
        db.close()


# ── Streaming RAG endpoint (used by Go WebSocket handler) ────────────────────

@app.post("/stream")
def stream_rag(req: StreamRequest):
    """
    Full RAG + LLM pipeline with SSE streaming:
    embed question → pgvector similarity search → build prompt → stream Gemini response
    """
    return StreamingResponse(
        _generate_stream(req.question, req.tenant_id),
        media_type="text/event-stream",
        headers={
            "Cache-Control":    "no-cache",
            "X-Accel-Buffering": "no",   # disables nginx buffering so tokens flow immediately
        },
    )


def _generate_stream(question: str, tenant_id: str):
    db  = get_db()
    cur = db.cursor()

    try:
        # Step 1 — embed the question
        q_result = genai.embed_content(
            model="models/gemini-embedding-001",
            content=question,
            task_type="retrieval_query",
        )
        q_embedding = q_result["embedding"]

        # Step 2 — retrieve top-5 relevant chunks for this tenant only
        cur.execute(
            """SELECT content, document_id,
                      1 - (embedding <=> %s::vector) AS similarity
               FROM chunks
               WHERE tenant_id = %s
               ORDER BY embedding <=> %s::vector
               LIMIT 5""",
            (q_embedding, tenant_id, q_embedding),
        )
        rows = cur.fetchall()

        if not rows:
            yield "data: " + json.dumps({"token": "I could not find relevant information in your documents."}) + "\n\n"
            yield "data: [DONE]\n\n"
            return

        # Step 3 — build grounded prompt
        context_text = "\n\n---\n\n".join([row[0] for row in rows])
        sources      = list(set(str(row[1]) for row in rows))

        prompt = f"""You are a helpful AI assistant for a company knowledge base.
Answer the user's question based ONLY on the provided context below.
If the context does not contain enough information, say so clearly.
Always be concise and accurate.

CONTEXT:
{context_text}

QUESTION:
{question}

ANSWER:"""

        # Step 4 — send sources immediately so the UI can render them
        yield "data: " + json.dumps({"sources": ",".join(sources)}) + "\n\n"

        # Step 5 — stream Gemini response token by token using a supported free-tier model
        model    = genai.GenerativeModel("models/gemini-2.5-flash-lite")
        response = model.generate_content(prompt, stream=True)

        for chunk in response:
            if chunk.text:
                yield "data: " + json.dumps({"token": chunk.text}) + "\n\n"

        yield "data: [DONE]\n\n"

    except Exception as e:
        yield "data: " + json.dumps({"token": f"Error: {str(e)}"}) + "\n\n"
        yield "data: [DONE]\n\n"

    finally:
        cur.close()
        db.close()


# ── Helpers ──────────────────────────────────────────────────────────────────

class SimpleTextLoader:
    def __init__(self, file_path: str):
        self.file_path = file_path

    def load(self):
        encoding = _detect_text_encoding(self.file_path)
        with open(self.file_path, "r", encoding=encoding, errors="replace") as f:
            return [Document(page_content=f.read(), metadata={"source": self.file_path})]


def _detect_text_encoding(file_path: str) -> str:
    with open(file_path, "rb") as f:
        header = f.read(4)
    if header.startswith(b"\xff\xfe") or header.startswith(b"\xfe\xff"):
        return "utf-16"
    if header.startswith(b"\xef\xbb\xbf"):
        return "utf-8-sig"
    return "utf-8"


def _get_loader(file_path: str, source_type: str):
    if source_type == "pdf":
        return PyPDFLoader(file_path)
    elif source_type == "docx":
        return Docx2txtLoader(file_path)
    else:
        return SimpleTextLoader(file_path)
