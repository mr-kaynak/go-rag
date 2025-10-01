import { useState, useEffect } from "react"
import { Brain, AlertCircle } from "lucide-react"
import { ThemeProvider } from "@/hooks/use-theme"
import { Toaster } from "@/components/ui/toaster"
import { ThemeToggle } from "@/components/theme-toggle"
import { FileUpload } from "@/components/file-upload"
import { ChatInterface } from "@/components/chat-interface"
import { SystemPromptEditor } from "@/components/system-prompt-editor"
import { TokenMetrics } from "@/components/token-metrics"
import { ApiKeysManager } from "@/components/api-keys-manager"
import { DocumentsList } from "@/components/documents-list"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { api } from "@/lib/api"

const DEFAULT_SYSTEM_PROMPT =
  "You are a helpful AI assistant. Answer questions based on the provided context."

function App() {
  const [provider, setProvider] = useState<"openrouter" | "bedrock">("openrouter")
  const [systemPrompt, setSystemPrompt] = useState(DEFAULT_SYSTEM_PROMPT)
  const [inputTokens, setInputTokens] = useState(0)
  const [outputTokens, setOutputTokens] = useState(0)
  const [serverConnected, setServerConnected] = useState<boolean | null>(null)

  useEffect(() => {
    const checkServer = async () => {
      try {
        await api.health()
        setServerConnected(true)

        // Fetch default system prompt from settings
        try {
          const defaultPrompt = await api.getDefaultSystemPrompt()
          if (defaultPrompt && defaultPrompt.prompt) {
            setSystemPrompt(defaultPrompt.prompt)
          }
        } catch (error) {
          console.log("No default system prompt found, using hardcoded default")
        }
      } catch (error) {
        setServerConnected(false)
      }
    }

    checkServer()
  }, [])

  const handleTokenUsage = (input: number, output: number) => {
    setInputTokens(prev => prev + input)
    setOutputTokens(prev => prev + output)
  }

  return (
    <ThemeProvider defaultTheme="dark" storageKey="rag-ui-theme">
      <div className="min-h-screen bg-background">
        <header className="border-b">
          <div className="container mx-auto px-4 py-4 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center">
                <Brain className="h-6 w-6 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-2xl font-bold">RAG System</h1>
                <p className="text-sm text-muted-foreground">
                  Retrieval-Augmented Generation
                </p>
              </div>
            </div>
            <ThemeToggle />
          </div>
        </header>

        <main className="container mx-auto px-4 py-8">
          {serverConnected === false && (
            <Alert variant="destructive" className="mb-6">
              <AlertCircle className="h-4 w-4" />
              <AlertTitle>Server Connection Failed</AlertTitle>
              <AlertDescription>
                Unable to connect to the backend server at http://localhost:3000.
                Please make sure the server is running.
              </AlertDescription>
            </Alert>
          )}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              <Card>
                <CardHeader>
                  <CardTitle>Provider Configuration</CardTitle>
                  <CardDescription>
                    Select your LLM provider
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-2">
                    <Label>Provider</Label>
                    <Tabs value={provider} onValueChange={(v) => setProvider(v as any)}>
                      <TabsList className="grid w-full grid-cols-2">
                        <TabsTrigger value="openrouter">OpenRouter</TabsTrigger>
                        <TabsTrigger value="bedrock">AWS Bedrock</TabsTrigger>
                      </TabsList>
                    </Tabs>
                  </div>
                </CardContent>
              </Card>

              <ChatInterface
                provider={provider}
                systemPrompt={systemPrompt}
                onTokenUsage={handleTokenUsage}
              />
            </div>

            <div className="space-y-6">
              <ApiKeysManager />

              <FileUpload />

              <DocumentsList />

              <SystemPromptEditor
                value={systemPrompt}
                onChange={setSystemPrompt}
                defaultPrompt={DEFAULT_SYSTEM_PROMPT}
              />

              <TokenMetrics inputTokens={inputTokens} outputTokens={outputTokens} />
            </div>
          </div>
        </main>
        <Toaster />
      </div>
    </ThemeProvider>
  )
}

export default App
