# NexusIQ — AI Knowledge Operations Platform

> Upload your company's documents. Let your team chat with them in real time.

A multi-tenant RAG (Retrieval-Augmented Generation) platform where businesses upload internal documents (SOPs, wikis, contracts, manuals) and employees query them via natural language chat — like ChatGPT for your company's brain.

Built with a Go microservice backbone, a Python ML pipeline, and a pgvector similarity search engine.

---

## Tech Stack

| Layer | Tech |
|---|---|
| API Gateway | Go 1.23 + Gin |
| ML Service | Python 3.12 + FastAPI + LangChain |
| Vector DB | PostgreSQL 16 + pgvector |
| Embeddings + LLM | Google Gemini API (text-embedding-004) |
| Cache + Rate Limiting | Redis 7 |
| Frontend | React + TypeScript + Tailwind CSS *(Day 4)* |
| Containerisation | Docker + Docker Compose |
| Reverse Proxy | Nginx |

## Features

- **Multi-tenant isolation** — every document, chunk, and chat belongs to a tenant; zero data leakage between organisations
- **Document ingestion pipeline** — upload PDF, DOCX, or TXT → auto-chunked → Gemini embeddings → stored in pgvector
- **RAG query engine** — cosine similarity search retrieves the top-5 most relevant chunks for any question
- **Streaming chat** — LLM responses stream token-by-token via WebSocket *(Day 3)*
- **Analytics dashboard** — see what questions are being asked and where knowledge gaps exist *(Day 4)*
- **JWT authentication** — secure login with role-based access (admin / member)

---

## Local Setup

### Prerequisites
- Docker Desktop
- A free [Google AI Studio](https://aistudio.google.com/) API key (Gemini)

### Run

```bash
git clone https://github.com/damsigeli07/NexusIQ.git
cd NexusIQ

cp .env.example .env
# Open .env and fill in your GEMINI_API_KEY

docker-compose up --build
```

### Verify

```bash
curl http://localhost:8080/api/health
# → {"status":"ok","service":"go-api"}

curl http://localhost:8000/health
# → {"status":"ok","service":"python-ml"}
```

---

## API Reference

| Method | Route | Auth | Description |
|---|---|---|---|
| GET | `/api/health` | — | Health check |
| POST | `/api/auth/register` | — | Create account (requires tenant_slug) |
| POST | `/api/auth/login` | — | Returns JWT token |
| GET | `/api/documents` | JWT | List all documents for your tenant |
| POST | `/api/documents` | JWT | Upload PDF / DOCX / TXT |
| DELETE | `/api/documents/:id` | JWT | Delete a document |
| GET | `/api/chat` | JWT | WebSocket — streaming RAG chat |
| GET | `/api/analytics` | JWT | Usage analytics |

---

## Project Status

- [x] Day 1 — Docker Compose infra, PostgreSQL schema with pgvector, Go skeleton, Redis
- [x] Day 2 — JWT auth, multi-tenant document upload, LangChain chunking, Gemini embeddings, pgvector storage
- [ ] Day 3 — WebSocket streaming, LLM response generation
- [ ] Day 4 — React + TypeScript frontend
- [ ] Day 5 — Cloud deployment + live demo

---