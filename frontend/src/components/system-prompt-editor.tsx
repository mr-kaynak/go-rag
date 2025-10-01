import { useState } from "react"
import { Settings, Eye, EyeOff, Save } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { useToast } from "@/hooks/use-toast"
import { api } from "@/lib/api"

interface SystemPromptEditorProps {
  value: string
  onChange: (value: string) => void
  defaultPrompt?: string
}

export function SystemPromptEditor({ value, onChange, defaultPrompt }: SystemPromptEditorProps) {
  const { toast } = useToast()
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [saving, setSaving] = useState(false)

  const handleReset = () => {
    if (defaultPrompt) {
      onChange(defaultPrompt)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.saveSystemPrompt({
        name: "Default",
        prompt: value,
        default: true,
      })
      toast({
        title: "Success",
        description: "System prompt saved successfully",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save system prompt",
        variant: "destructive",
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            <CardTitle>System Prompt</CardTitle>
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setIsCollapsed(!isCollapsed)}
          >
            {isCollapsed ? (
              <Eye className="h-4 w-4" />
            ) : (
              <EyeOff className="h-4 w-4" />
            )}
          </Button>
        </div>
        {!isCollapsed && (
          <CardDescription>
            Customize the AI assistant's behavior and instructions
          </CardDescription>
        )}
      </CardHeader>
      {!isCollapsed && (
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="system-prompt">Prompt Template</Label>
            <Textarea
              id="system-prompt"
              value={value}
              onChange={e => onChange(e.target.value)}
              placeholder="Enter system prompt..."
              className="min-h-[120px] font-mono text-sm"
            />
          </div>
          <div className="flex justify-between items-center">
            <span className="text-xs text-muted-foreground">{value.length} characters</span>
            <div className="flex gap-2">
              {defaultPrompt && value !== defaultPrompt && (
                <Button variant="outline" size="sm" onClick={handleReset}>
                  Reset to default
                </Button>
              )}
              <Button size="sm" onClick={handleSave} disabled={saving} className="gap-2">
                <Save className="h-3 w-3" />
                {saving ? "Saving..." : "Save"}
              </Button>
            </div>
          </div>
        </CardContent>
      )}
    </Card>
  )
}
