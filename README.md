# Enterprise RAG System

A production-ready RAG (Retrieval-Augmented Generation) system with a Go backend and React frontend. Supports multiple LLM providers (OpenRouter, AWS Bedrock, Ollama) with document ingestion, semantic search, and a modern web interface.

## Features

### Backend
- 🚀 **High Performance** - Built with Go and Fiber framework
- 🤖 **Multi-Provider Support** - OpenRouter, AWS Bedrock, and Ollama integration
- 📄 **Document Management** - Upload, process, and delete documents with automatic chunking
- 🔍 **Semantic Search** - Vector similarity search using cosine similarity
- 💾 **Persistent Storage** - BadgerDB for settings/metadata, JSON for vector storage
- 🔐 **Encrypted Settings** - AES-256 encryption for API keys
- 🎯 **Flexible Embeddings** - Support for Ollama, OpenRouter, and Bedrock embeddings
- 📊 **Streaming Support** - Real-time streaming responses for Bedrock
- 🔧 **Configurable** - System prompts, models, and chunking parameters
- 🛡️ **Production-Ready** - Structured logging, error handling, CORS, and graceful shutdown

### Frontend
- ⚛️ **Modern React** - Built with React 19, TypeScript, and Vite
- 🎨 **Beautiful UI** - Tailwind CSS with shadcn/ui components
- 🌓 **Dark/Light Mode** - Theme switching with persistence
- 💬 **Chat Interface** - Real-time chat with context display
- 📁 **File Upload** - Drag-and-drop document upload
- 📋 **Document List** - View and delete uploaded documents
- ⚙️ **Settings Management** - API keys, system prompts, and provider configuration
- 📊 **Token Metrics** - Track input/output token usage

## Quick Start

### Prerequisites

- **Go 1.25+** (for backend)
- **Node.js 18+** (for frontend)
- **Ollama** (optional, for local embeddings) - [Install Ollama](https://ollama.ai)
- **API Keys** (at least one):
  - OpenRouter API key, or
  - AWS Bedrock API key

### Backend Setup

1. Clone the repository:
```bash
git clone https://github.com/mrkaynak/rag.git
cd rag
```

2. Install Go dependencies:
```bash
go mod download
```

3. Create `.env` file:
```bash
cp .env.example .env
```

4. Configure your environment variables (see [Configuration](#configuration) section)

5. If using Ollama for embeddings, pull the embedding model:
```bash
ollama pull all-minilm:33m
```

6. Run the backend server:
```bash
go run cmd/server/main.go
```

The backend will start on `http://localhost:3000`

### Frontend Setup

1. Navigate to frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm run dev
```

The frontend will start on `http://localhost:5173`

4. Build for production:
```bash
npm run build
```

### Quick Start with Ollama (No API Keys Required)

For a completely local setup without any API keys:

1. Install Ollama: https://ollama.ai
2. Pull models:
```bash
ollama pull all-minilm:33m  # For embeddings
```
3. Set in `.env`:
```env
EMBEDDING_PROVIDER=ollama
EMBEDDING_MODEL=all-minilm:33m
OLLAMA_BASE_URL=http://localhost:11434
```
4. For LLM, you'll still need OpenRouter or Bedrock API key

## API Documentation

### Health & System

#### Health Check
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

#### Get System Prompt
```bash
GET /api/v1/system-prompt
```

### Document Management

#### Upload Document
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

#### List Documents
```bash
GET /api/v1/documents
```

#### Delete Document
```bash
DELETE /api/v1/documents/:id
```

### Chat

#### Chat (Non-streaming)
```bash
POST /api/v1/chat
Content-Type: application/json

{
  "message": "What is the main topic?",
  "provider": "openrouter",
  "model": "anthropic/claude-3.5-sonnet",
  "system_prompt": "Custom prompt (optional)"
}
```

**Response:**
```json
{
  "message": "Based on the context...",
  "context": ["chunk1", "chunk2"]
}
```

#### Chat Stream (SSE)
```bash
POST /api/v1/chat/stream
Content-Type: application/json

{
  "message": "What is RAG?",
  "provider": "bedrock"
}
```

**SSE Events:**
- `context` - Retrieved document chunks
- `chunk` - Streaming text chunk
- `done` - Stream completed
- `error` - Error occurred

### Settings

#### API Keys
```bash
# Save API keys (encrypted)
POST /api/v1/settings/api-keys
{
  "openrouter": "sk-...",
  "bedrock": "aws-..."
}

# Get API keys (masked)
GET /api/v1/settings/api-keys
```

#### Models
```bash
# Save model configuration
POST /api/v1/settings/models
{
  "provider": "openrouter",
  "model_id": "anthropic/claude-3.5-sonnet",
  "display_name": "Claude 3.5 Sonnet"
}

# List models
GET /api/v1/settings/models

# Delete model
DELETE /api/v1/settings/models/:id
```

#### System Prompts
```bash
# Save system prompt
POST /api/v1/settings/system-prompts
{
  "name": "Default",
  "prompt": "You are a helpful assistant...",
  "default": true
}

# List system prompts
GET /api/v1/settings/system-prompts

# Get default system prompt
GET /api/v1/settings/system-prompts/default

# Delete system prompt
DELETE /api/v1/settings/system-prompts/:id
```

## Architecture

### Project Structure

```
rag/
├── cmd/server/              # Application entrypoint
│   └── main.go
├── frontend/                # React frontend (React 19 + TypeScript)
│   ├── src/
│   │   ├── components/      # UI components (shadcn/ui)
│   │   │   ├── ui/          # Base UI components
│   │   │   ├── chat-interface.tsx
│   │   │   ├── file-upload.tsx
│   │   │   ├── documents-list.tsx
│   │   │   ├── api-keys-manager.tsx
│   │   │   ├── system-prompt-editor.tsx
│   │   │   └── token-metrics.tsx
│   │   ├── lib/            # API client & utilities
│   │   │   ├── api.ts      # Backend API client
│   │   │   └── utils.ts    # Helper functions
│   │   ├── hooks/          # React hooks
│   │   │   ├── use-theme.tsx
│   │   │   └── use-toast.ts
│   │   ├── App.tsx         # Main application
│   │   └── main.tsx        # Entry point
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── internal/
│   ├── config/              # Environment configuration
│   │   └── config.go        # Config loading & validation
│   ├── handler/             # HTTP request handlers
│   │   ├── chat.go          # Chat & streaming endpoints
│   │   ├── upload.go        # Document upload & management
│   │   ├── settings.go      # Settings API
│   │   └── health.go        # Health check
│   ├── middleware/          # HTTP middleware
│   │   ├── cors.go          # CORS configuration
│   │   ├── logger.go        # Request logging
│   │   └── recovery.go      # Panic recovery
│   ├── models/              # Data structures
│   │   └── models.go        # Document, Chunk, Request/Response types
│   └── service/
│       ├── document/
│       │   ├── document.go  # Document processing & chunking
│       │   └── metadata.go  # Document metadata store (BadgerDB)
│       ├── embeddings/
│       │   └── embeddings.go # Multi-provider embeddings
│       ├── llm/
│       │   ├── openrouter.go # OpenRouter client
│       │   └── bedrock.go    # AWS Bedrock client (with streaming)
│       ├── settings/
│       │   ├── settings.go   # Settings store (BadgerDB, encrypted)
│       │   └── seed.go       # Initial data seeding
│       └── vector/
│           └── vector.go     # Vector similarity search (JSON)
├── pkg/
│   └── errors/              # Custom error types
├── data/                    # Persistent data (auto-created)
│   ├── uploads/             # Uploaded documents
│   ├── vectors/             # Vector embeddings (JSON)
│   └── badger/              # BadgerDB files
├── .env.example             # Example environment configuration
├── .env                     # Your configuration (gitignored)
├── go.mod                   # Go dependencies
└── README.md                # This file
```

### How It Works

```
┌──────────────┐
│   Upload     │
│  Document    │
└──────┬───────┘
       │
       ▼
┌──────────────┐      ┌──────────────┐
│   Document   │─────▶│    Chunk     │
│   Service    │      │  (overlap)   │
└──────────────┘      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │  Embeddings  │◀───Ollama/OpenRouter/Bedrock
                      │   Service    │
                      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │   Vector     │
                      │    Store     │
                      └──────┬───────┘
                             │
        ┌────────────────────┴────────────────────┐
        │                                         │
        ▼                                         ▼
┌──────────────┐                          ┌──────────────┐
│    Query     │                          │   Metadata   │
│  (search)    │                          │    Store     │
└──────┬───────┘                          └──────────────┘
       │
       ▼
┌──────────────┐      ┌──────────────┐
│  Retrieve    │─────▶│     LLM      │◀───OpenRouter/Bedrock
│   Context    │      │  (augment)   │
└──────────────┘      └──────┬───────┘
                             │
                             ▼
                      ┌──────────────┐
                      │   Response   │
                      └──────────────┘
```

**Flow:**

1. **Document Upload**: Documents are uploaded and split into overlapping chunks (configurable size)
2. **Embedding Generation**: Each chunk is converted to a vector embedding via Ollama/OpenRouter/Bedrock
3. **Vector Storage**: Embeddings are stored in-memory with JSON persistence for fast similarity search
4. **Metadata Storage**: Document metadata (filename, size, chunk count) stored in BadgerDB
5. **Query Processing**: User questions are embedded and similar chunks retrieved using cosine similarity
6. **Context Augmentation**: Top K most relevant chunks are added to the LLM prompt as context
7. **Response Generation**: LLM generates answer based on retrieved context + user question
8. **Streaming**: For Bedrock, responses can be streamed in real-time via SSE

### Technology Stack

**Backend:**
- **Go 1.25** - High-performance compiled language
- **Fiber v2** - Express-inspired web framework
- **BadgerDB v4** - Embedded key-value store for settings & metadata
- **Zap** - Structured, leveled logging
- **AES-256-GCM** - Encryption for sensitive API keys

**Frontend:**
- **React 19** - Latest React with concurrent features
- **TypeScript 5.8** - Type-safe development
- **Vite 7** - Fast build tool and dev server
- **Tailwind CSS 3.4** - Utility-first CSS framework
- **shadcn/ui** - High-quality component library
- **Radix UI** - Accessible primitives

**Storage:**
- **BadgerDB**: Settings, API keys (encrypted), metadata
- **JSON**: Vector embeddings (in-memory + file persistence)
- **File System**: Uploaded documents

## Configuration

### Environment Variables

All configuration via `.env` file:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| **Server** |
| `PORT` | Server port | `3000` | No |
| `ENV` | Environment (development/production) | `development` | No |
| **OpenRouter** |
| `OPENROUTER_API_KEY` | OpenRouter API key | - | Yes* |
| `OPENROUTER_MODEL` | Default model | `anthropic/claude-3.5-sonnet` | No |
| **AWS Bedrock** |
| `BEDROCK_API_KEY` | AWS Bedrock API key | - | Yes* |
| `BEDROCK_REGION` | AWS region | `eu-north-1` | No |
| `BEDROCK_MODEL_ID` | Model ID | `openai.gpt-oss-20b-1:0` | No |
| **Ollama** |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://localhost:11434` | No |
| **Embeddings** |
| `EMBEDDING_PROVIDER` | Provider: `ollama`, `openrouter`, `bedrock` | `ollama` | No |
| `EMBEDDING_MODEL` | Model name | `all-minilm:33m` | No |
| `EMBEDDING_DIMENSIONS` | Vector dimensions | `384` | No |
| **Storage** |
| `UPLOAD_DIR` | Upload directory | `./data/uploads` | No |
| `VECTOR_STORE_PATH` | Vector store path | `./data/vectors` | No |
| `BADGER_DB_PATH` | BadgerDB path | `./data/badger` | No |
| **Encryption** |
| `ENCRYPTION_KEY` | 32-byte AES-256 key | - | Recommended |
| **RAG** |
| `MAX_CONTEXT_CHUNKS` | Max chunks in context | `5` | No |
| `CHUNK_SIZE` | Characters per chunk | `1000` | No |
| `CHUNK_OVERLAP` | Overlap between chunks | `200` | No |
| `SYSTEM_PROMPT` | Default system prompt | Built-in | No |

\* At least one LLM provider (OpenRouter or Bedrock) is required

### Example Configurations

**Ollama (Local, No API Keys for Embeddings):**
```env
PORT=3000
ENV=development

# Ollama for embeddings (local, no API key needed)
EMBEDDING_PROVIDER=ollama
EMBEDDING_MODEL=all-minilm:33m
EMBEDDING_DIMENSIONS=384
OLLAMA_BASE_URL=http://localhost:11434

# OpenRouter for LLM (API key required)
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=anthropic/claude-3.5-sonnet

# Storage
UPLOAD_DIR=./data/uploads
VECTOR_STORE_PATH=./data/vectors
BADGER_DB_PATH=./data/badger

# Encryption (generate a secure 32-byte key)
ENCRYPTION_KEY=your-32-byte-encryption-key-here!!

# RAG settings
MAX_CONTEXT_CHUNKS=5
CHUNK_SIZE=1000
CHUNK_OVERLAP=200
```

**OpenRouter (Cloud):**
```env
EMBEDDING_PROVIDER=openrouter
EMBEDDING_MODEL=openai/text-embedding-3-small
EMBEDDING_DIMENSIONS=1536
OPENROUTER_API_KEY=sk-or-v1-...
OPENROUTER_MODEL=anthropic/claude-3.5-sonnet
```

**AWS Bedrock:**
```env
EMBEDDING_PROVIDER=bedrock
EMBEDDING_MODEL=amazon.titan-embed-text-v1
BEDROCK_API_KEY=your_aws_key
BEDROCK_REGION=us-east-1
BEDROCK_MODEL_ID=anthropic.claude-3-sonnet-20240229-v1:0
```

## Production Deployment

### Building

**Backend:**
```bash
# Build binary
go build -o rag-server cmd/server/main.go

# Run
./rag-server
```

**Frontend:**
```bash
cd frontend
npm run build
# Build output in frontend/dist/
```

### Docker Deployment

Create `Dockerfile`:
```dockerfile
# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rag-server cmd/server/main.go

# Stage 3: Production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/rag-server .
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/frontend/dist
COPY .env.example .env

# Create data directories
RUN mkdir -p data/uploads data/vectors data/badger

EXPOSE 3000
CMD ["./rag-server"]
```

Build and run:
```bash
docker build -t rag-system .
docker run -p 3000:3000 --env-file .env rag-system
```

### Docker Compose

Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  rag-server:
    build: .
    ports:
      - "3000:3000"
    env_file:
      - .env
    volumes:
      - ./data:/app/data
    restart: unless-stopped

  # Optional: Ollama for local embeddings
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    restart: unless-stopped

volumes:
  ollama-data:
```

Run:
```bash
docker-compose up -d
```

## Performance Considerations

- **Vector Store**: Currently uses in-memory storage with JSON persistence. For production at scale (100k+ documents), consider:
  - PostgreSQL with pgvector
  - Qdrant
  - Weaviate
  - Pinecone

- **Concurrency**: The vector store uses read-write locks (`sync.RWMutex`) for thread-safe operations

- **File Size**: Current implementation loads entire documents into memory. For large files (>10MB):
  - Implement streaming file processing
  - Add file size limits
  - Process in batches

- **Chunking**: Current implementation uses character-based chunking. Consider:
  - Sentence-aware chunking
  - Paragraph-based chunking
  - Token-based chunking for better LLM compatibility

- **Rate Limiting**: Add rate limiting middleware for production:
```go
import "github.com/gofiber/fiber/v2/middleware/limiter"

app.Use(limiter.New(limiter.Config{
    Max:        100,
    Expiration: 1 * time.Minute,
}))
```

## Security Best Practices

- ✅ **HTTPS**: Always use HTTPS in production (reverse proxy with nginx/Caddy)
- ✅ **API Keys**: Rotate API keys regularly
- ✅ **Encryption**: Use a strong 32-byte `ENCRYPTION_KEY` for AES-256
- ✅ **Authentication**: Implement authentication middleware for production
- ✅ **File Upload**: Validate file types and sizes:
  ```go
  if file.Size > 10*1024*1024 { // 10MB limit
      return errors.BadRequest("file too large")
  }
  ```
- ✅ **CORS**: Configure CORS for specific origins in production
- ✅ **Rate Limiting**: Prevent abuse with rate limiting
- ✅ **Input Validation**: Sanitize all user inputs
- ✅ **Monitoring**: Log all API access and monitor for anomalies
- ✅ **Environment Variables**: Never commit `.env` to version control

## Troubleshooting

### Backend Issues

**Port already in use:**
```bash
# Change PORT in .env
PORT=3001
```

**BadgerDB errors:**
```bash
# Delete corrupted database
rm -rf data/badger
# Restart server (will recreate)
```

**Ollama connection failed:**
```bash
# Check if Ollama is running
ollama list

# Start Ollama
ollama serve
```

### Frontend Issues

**API connection failed:**
- Check backend is running on port 3000
- Verify CORS settings in backend
- Check browser console for errors

**Build errors:**
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and effective Go guidelines
- Use TypeScript strict mode for frontend
- Write meaningful commit messages
- Add tests for new features
- Update documentation

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: [GitHub Issues](https://github.com/mrkaynak/rag/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mrkaynak/rag/discussions)

## Roadmap

- [ ] PostgreSQL with pgvector support
- [ ] OpenAI embeddings support
- [ ] PDF document support
- [ ] Multi-language support
- [ ] Conversation history
- [ ] User authentication
- [ ] Admin dashboard
- [ ] Kubernetes deployment guide
- [ ] Prometheus metrics
- [ ] Unit and integration tests

## Acknowledgments

- [Fiber](https://gofiber.io/) - Web framework
- [BadgerDB](https://github.com/dgraph-io/badger) - Embedded database
- [shadcn/ui](https://ui.shadcn.com/) - UI components
- [Ollama](https://ollama.ai/) - Local LLM inference
- [OpenRouter](https://openrouter.ai/) - LLM API aggregator
- [AWS Bedrock](https://aws.amazon.com/bedrock/) - Managed LLM service
