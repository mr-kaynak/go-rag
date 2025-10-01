import { Activity, ArrowDown, ArrowUp } from "lucide-react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

interface TokenMetricsProps {
  inputTokens: number
  outputTokens: number
}

export function TokenMetrics({ inputTokens, outputTokens }: TokenMetricsProps) {
  const totalTokens = inputTokens + outputTokens

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat().format(num)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="h-5 w-5" />
          Token Usage
        </CardTitle>
        <CardDescription>
          Track your API token consumption
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-center justify-between p-3 rounded-lg bg-muted">
            <div className="flex items-center gap-2">
              <ArrowUp className="h-4 w-4 text-blue-500" />
              <span className="text-sm font-medium">Input Tokens</span>
            </div>
            <Badge variant="secondary">{formatNumber(inputTokens)}</Badge>
          </div>

          <div className="flex items-center justify-between p-3 rounded-lg bg-muted">
            <div className="flex items-center gap-2">
              <ArrowDown className="h-4 w-4 text-green-500" />
              <span className="text-sm font-medium">Output Tokens</span>
            </div>
            <Badge variant="secondary">{formatNumber(outputTokens)}</Badge>
          </div>

          <div className="pt-2 border-t">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold">Total Tokens</span>
              <Badge>{formatNumber(totalTokens)}</Badge>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
