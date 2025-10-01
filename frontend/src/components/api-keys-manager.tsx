import { useState, useEffect } from "react"
import { Key, Eye, EyeOff, Save } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useToast } from "@/hooks/use-toast"
import { api } from "@/lib/api"

export function ApiKeysManager() {
  const { toast } = useToast()
  const [openRouterKey, setOpenRouterKey] = useState("")
  const [bedrockKey, setBedrockKey] = useState("")
  const [showOpenRouter, setShowOpenRouter] = useState(false)
  const [showBedrock, setShowBedrock] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    loadKeys()
  }, [])

  const loadKeys = async () => {
    try {
      const keys = await api.getAPIKeys()
      setOpenRouterKey(keys.openrouter || "")
      setBedrockKey(keys.bedrock || "")
    } catch (error) {
      console.error("Failed to load API keys:", error)
    }
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      await api.saveAPIKeys({
        openrouter: openRouterKey,
        bedrock: bedrockKey,
      })
      toast({
        title: "Success",
        description: "API keys saved successfully",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save API keys",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Key className="h-5 w-5" />
          API Keys
        </CardTitle>
        <CardDescription>
          Configure your LLM provider API keys
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="openrouter">OpenRouter API Key</Label>
          <div className="flex gap-2">
            <Input
              id="openrouter"
              type={showOpenRouter ? "text" : "password"}
              value={openRouterKey}
              onChange={(e) => setOpenRouterKey(e.target.value)}
              placeholder="sk-or-..."
            />
            <Button
              variant="outline"
              size="icon"
              onClick={() => setShowOpenRouter(!showOpenRouter)}
            >
              {showOpenRouter ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="bedrock">AWS Bedrock API Key</Label>
          <div className="flex gap-2">
            <Input
              id="bedrock"
              type={showBedrock ? "text" : "password"}
              value={bedrockKey}
              onChange={(e) => setBedrockKey(e.target.value)}
              placeholder="ABSK..."
            />
            <Button
              variant="outline"
              size="icon"
              onClick={() => setShowBedrock(!showBedrock)}
            >
              {showBedrock ? (
                <EyeOff className="h-4 w-4" />
              ) : (
                <Eye className="h-4 w-4" />
              )}
            </Button>
          </div>
        </div>

        <Button onClick={handleSave} disabled={loading} className="gap-2">
          <Save className="h-4 w-4" />
          {loading ? "Saving..." : "Save API Keys"}
        </Button>
      </CardContent>
    </Card>
  )
}
