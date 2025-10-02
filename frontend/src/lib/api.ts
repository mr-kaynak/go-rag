const API_BASE_URL = 'http://localhost:3000/api/v1'

export interface UploadResponse {
  document_id: string
  file_name: string
  chunk_count: number
}

export interface ChatRequest {
  message: string
  provider: 'openrouter' | 'bedrock'
  model?: string
  system_prompt?: string
}

export interface TokenMetrics {
  input_tokens: number
  output_tokens: number
  total_tokens: number
}

export interface ChatResponse {
  message: string
  context?: string[]
  token_metrics?: TokenMetrics
}

export interface HealthResponse {
  status: string
  version: string
}

export interface APIKeys {
  openrouter?: string
  bedrock?: string
}

export interface Document {
  id: string
  file_name: string
  file_size: number
  chunk_count: number
  uploaded_at: string
}

export interface SystemPromptRequest {
  name: string
  prompt: string
  default: boolean
}

export const api = {
  async health(): Promise<HealthResponse> {
    const response = await fetch(`${API_BASE_URL}/health`)
    if (!response.ok) throw new Error('Health check failed')
    return response.json()
  },

  async getSystemPrompt(): Promise<{ system_prompt: string }> {
    const response = await fetch(`${API_BASE_URL}/system-prompt`)
    if (!response.ok) throw new Error('Failed to fetch system prompt')
    return response.json()
  },

  async upload(file: File): Promise<UploadResponse> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch(`${API_BASE_URL}/upload`, {
      method: 'POST',
      body: formData,
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Upload failed')
    }

    return response.json()
  },

  async chat(request: ChatRequest): Promise<ChatResponse> {
    const response = await fetch(`${API_BASE_URL}/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Chat request failed')
    }

    return response.json()
  },

  async saveAPIKeys(keys: APIKeys): Promise<{ success: boolean }> {
    const response = await fetch(`${API_BASE_URL}/settings/api-keys`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(keys),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to save API keys')
    }

    return response.json()
  },

  async getAPIKeys(): Promise<APIKeys> {
    const response = await fetch(`${API_BASE_URL}/settings/api-keys`)
    if (!response.ok) throw new Error('Failed to get API keys')
    return response.json()
  },

  async listDocuments(): Promise<Document[]> {
    const response = await fetch(`${API_BASE_URL}/documents`)
    if (!response.ok) throw new Error('Failed to list documents')
    return response.json()
  },

  async deleteDocument(id: string): Promise<{ success: boolean }> {
    const response = await fetch(`${API_BASE_URL}/documents/${id}`, {
      method: 'DELETE',
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to delete document')
    }

    return response.json()
  },

  async saveSystemPrompt(prompt: SystemPromptRequest): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/settings/system-prompts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(prompt),
    })

    if (!response.ok) {
      const error = await response.json()
      throw new Error(error.error || 'Failed to save system prompt')
    }

    return response.json()
  },

  async getDefaultSystemPrompt(): Promise<{ id: string; name: string; prompt: string; default: boolean }> {
    const response = await fetch(`${API_BASE_URL}/settings/system-prompts/default`)
    if (!response.ok) throw new Error('Failed to get default system prompt')
    return response.json()
  },
}
