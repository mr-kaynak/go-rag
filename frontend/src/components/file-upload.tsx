import { useState } from "react"
import { Upload, FileText, Loader2, CheckCircle2, X, AlertCircle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { api, type UploadResponse } from "@/lib/api"

interface UploadedFile {
  id: string
  name: string
  chunkCount: number
}

const MAX_FILE_SIZE = 50 * 1024 * 1024 // 50MB

export function FileUpload() {
  const [uploading, setUploading] = useState(false)
  const [uploadedFiles, setUploadedFiles] = useState<UploadedFile[]>([])
  const [error, setError] = useState<string | null>(null)

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
  }

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Client-side validation
    if (file.size > MAX_FILE_SIZE) {
      setError(`File too large. Maximum file size is ${formatFileSize(MAX_FILE_SIZE)}`)
      e.target.value = ''
      return
    }

    if (file.size === 0) {
      setError('File is empty. Please select a valid file.')
      e.target.value = ''
      return
    }

    setUploading(true)
    setError(null)

    try {
      const result: UploadResponse = await api.upload(file)
      setUploadedFiles(prev => [
        ...prev,
        {
          id: result.document_id,
          name: result.file_name,
          chunkCount: result.chunk_count,
        },
      ])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setUploading(false)
      e.target.value = ''
    }
  }

  const removeFile = (id: string) => {
    setUploadedFiles(prev => prev.filter(f => f.id !== id))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Upload className="h-5 w-5" />
          Document Upload
        </CardTitle>
        <CardDescription>
          Upload documents to build your knowledge base for RAG queries (max {formatFileSize(MAX_FILE_SIZE)})
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="text-xs">
            Supported formats: TXT, MD • Maximum size: {formatFileSize(MAX_FILE_SIZE)} • Files are processed and split into chunks for semantic search
          </AlertDescription>
        </Alert>

        <div className="flex items-center justify-center w-full">
          <label
            htmlFor="file-upload"
            className="flex flex-col items-center justify-center w-full h-32 border-2 border-dashed rounded-lg cursor-pointer hover:bg-accent/50 transition-colors"
          >
            <div className="flex flex-col items-center justify-center pt-5 pb-6">
              {uploading ? (
                <>
                  <Loader2 className="h-8 w-8 mb-2 text-primary animate-spin" />
                  <p className="text-sm text-muted-foreground">Uploading...</p>
                </>
              ) : (
                <>
                  <Upload className="h-8 w-8 mb-2 text-muted-foreground" />
                  <p className="mb-2 text-sm text-muted-foreground">
                    <span className="font-semibold">Click to upload</span> or drag and drop
                  </p>
                  <p className="text-xs text-muted-foreground">TXT or MD files only</p>
                </>
              )}
            </div>
            <input
              id="file-upload"
              type="file"
              className="hidden"
              accept=".txt,.md"
              onChange={handleFileChange}
              disabled={uploading}
            />
          </label>
        </div>

        {error && (
          <div className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
            {error}
          </div>
        )}

        {uploadedFiles.length > 0 && (
          <div className="space-y-2">
            <h4 className="text-sm font-medium flex items-center gap-2">
              <CheckCircle2 className="h-4 w-4 text-green-500" />
              Uploaded Documents ({uploadedFiles.length})
            </h4>
            <div className="space-y-2">
              {uploadedFiles.map(file => (
                <div
                  key={file.id}
                  className="flex items-center justify-between p-3 rounded-lg border bg-card"
                >
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    <FileText className="h-4 w-4 text-muted-foreground flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{file.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {file.chunkCount} chunks
                      </p>
                    </div>
                    <Badge variant="secondary" className="flex-shrink-0">
                      Indexed
                    </Badge>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 flex-shrink-0"
                    onClick={() => removeFile(file.id)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
