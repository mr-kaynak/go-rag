# Enterprise RAG System

A lightweight, production-ready RAG (Retrieval-Augmented Generation) chat API built with Go and Fiber. Supports OpenRouter and AWS Bedrock for LLM inference with easy document ingestion and semantic search.

## Features

- 🚀 **Lightweight & Fast** - Built with Go Fiber for high performance
- 🤖 **Multi-Provider Support** - OpenRouter and AWS Bedrock integration
- 📄 **Easy Document Upload** - Support for text files with automatic chunking
- 🔍 **Semantic Search** - Vector similarity search with cosine similarity
- 🛡️ **Enterprise-Ready** - Structured logging, error handling, and graceful shutdown
- 🔧 **Simple Configuration** - Environment-based configuration
- 📦 **In-Memory Vector Store** - Lightweight vector storage with persistence

## Quick Start

### Prerequisites

- Go 1.25 or higher
- OpenRouter API key and/or AWS Bedrock API key

### Installation

1. Clone the repository:
```bash
git clone https://github.com/mrkaynak/rag.git
cd rag
```

2. Install dependencies:
```bash
go mod download
```

3. Create `.env` file:
```bash
cp .env.example .env
```

4. Configure your environment variables:
```env
# Server
PORT=3000

# OpenRouter (optional if using Bedrock)
OPENROUTER_API_KEY=your_key_here
OPENROUTER_MODEL=anthropic/claude-3.5-sonnet

# AWS Bedrock (optional if using OpenRouter)
BEDROCK_API_KEY=your_key_here
BEDROCK_REGION=eu-north-1
BEDROCK_MODEL_ID=openai.gpt-oss-20b-1:0

# Embeddings (choose provider: openrouter or bedrock)
EMBEDDING_PROVIDER=openrouter
EMBEDDING_MODEL=openai/text-embedding-3-small
# For Bedrock use: amazon.titan-embed-text-v1 or cohere.embed-english-v3
```

5. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:3000`

## API Documentation

### Health Check

```bash
GET /api/v1/health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Upload Document

Upload a document to be indexed for RAG.

```bash
POST /api/v1/upload
Content-Type: multipart/form-data

file: @document.txt
```

**Response:**
```json
{
  "document_id": "550e8400-e29b-41d4-a716-446655440000",
  "file_name": "document.txt",
  "chunk_count": 15
}
```

**Example:**
```bash
curl -X POST http://localhost:3000/api/v1/upload \
  -F "file=@knowledge.txt"
```

### Chat with RAG

Send a chat message that will be augmented with relevant context from uploaded documents.

```bash
POST /api/v1/chat
Content-Type: application/json

{
  "message": "What is the main topic of the document?",
  "provider": "openrouter",
  "model": "anthropic/claude-3.5-sonnet"
}
```

**Request Parameters:**
- `message` (string, required) - Your question or message
- `provider` (string, required) - Either "openrouter" or "bedrock"
- `model` (string, optional) - Override default model for this request

**Response:**
```json
{
  "message": "Based on the provided context, the main topic...",
  "context": [
    "Relevant chunk 1...",
    "Relevant chunk 2..."
  ]
}
```

**Examples:**

Using OpenRouter:
```bash
curl -X POST http://localhost:3000/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is RAG?",
    "provider": "openrouter"
  }'
```

Using AWS Bedrock:
```bash
curl -X POST http://localhost:3000/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What is RAG?",
    "provider": "bedrock",
    "model": "openai.gpt-oss-20b-1:0"
  }'
```

## Architecture

```
rag/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   └── service/
│       ├── document/    # Document processing
│       ├── embeddings/  # Embedding generation
│       ├── llm/         # LLM clients (OpenRouter, Bedrock)
│       └── vector/      # Vector store
└── pkg/
    └── errors/          # Custom error types
```

## Configuration

All configuration is done via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3000` |
| `ENV` | Environment (development/production) | `development` |
| `OPENROUTER_API_KEY` | OpenRouter API key | - |
| `OPENROUTER_MODEL` | Default OpenRouter model | `anthropic/claude-3.5-sonnet` |
| `BEDROCK_API_KEY` | AWS Bedrock API key | - |
| `BEDROCK_REGION` | AWS Bedrock region | `eu-north-1` |
| `BEDROCK_MODEL_ID` | Default Bedrock model ID | `openai.gpt-oss-20b-1:0` |
| `EMBEDDING_PROVIDER` | Embedding provider (openrouter/bedrock) | `openrouter` |
| `EMBEDDING_MODEL` | Embedding model | `openai/text-embedding-3-small` |
| `EMBEDDING_DIMENSIONS` | Embedding dimensions | `1536` |
| `MAX_CONTEXT_CHUNKS` | Max chunks for context | `5` |
| `CHUNK_SIZE` | Document chunk size | `1000` |
| `CHUNK_OVERLAP` | Chunk overlap size | `200` |

## How It Works

1. **Document Upload**: Documents are uploaded and split into chunks with overlap
2. **Embedding Generation**: Each chunk is converted to a vector embedding
3. **Vector Storage**: Embeddings are stored in an in-memory vector store with file persistence
4. **Query Processing**: User questions are embedded and similar chunks are retrieved
5. **RAG Response**: Retrieved context is provided to the LLM for accurate responses

## Production Deployment

### Building

```bash
go build -o rag-server cmd/server/main.go
```

### Running

```bash
./rag-server
```

### Docker (Optional)

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o rag-server cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/rag-server .
COPY .env .
EXPOSE 3000
CMD ["./rag-server"]
```

## Performance Considerations

- **Vector Store**: Currently uses in-memory storage with JSON persistence. For production at scale, consider PostgreSQL with pgvector or Qdrant
- **Concurrency**: The vector store uses read-write locks for thread-safe operations
- **File Size**: Suitable for small to medium document collections. For large-scale, implement chunked processing
- **Rate Limiting**: Consider adding rate limiting middleware for production use

## Security Best Practices

- Always use HTTPS in production
- Rotate API keys regularly
- Implement rate limiting
- Add authentication/authorization middleware
- Validate file types and sizes on upload
- Monitor and log all API access

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For issues and questions, please open an issue on GitHub.
