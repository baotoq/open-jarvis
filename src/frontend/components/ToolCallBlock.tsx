'use client'
import { FileText, Terminal, ChevronDown, Check, X } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import type { ToolCallPart, ToolResultPart } from '@/lib/api'

interface ToolCallBlockProps {
  call: ToolCallPart
  result?: ToolResultPart
}

export function ToolCallBlock({ call, result }: ToolCallBlockProps) {
  const Icon = call.name === 'shell_run' ? Terminal : FileText
  const hasError = result?.error
  return (
    <Collapsible className="w-full rounded-md border bg-muted/30 text-sm">
      <CollapsibleTrigger className="flex w-full items-center gap-2 px-3 py-2 hover:bg-muted/50">
        <Icon className="h-4 w-4 text-muted-foreground" />
        <Badge variant="outline" className="font-mono text-xs">{call.name}</Badge>
        {result && (hasError
          ? <X className="ml-auto h-4 w-4 text-destructive" />
          : <Check className="ml-auto h-4 w-4 text-green-600" />
        )}
        <ChevronDown className="ml-auto h-4 w-4 text-muted-foreground" />
      </CollapsibleTrigger>
      <CollapsibleContent className="px-3 pb-3">
        <div className="mt-1 font-mono text-xs text-muted-foreground">args: {call.args}</div>
        {result && (
          <div className={`mt-2 whitespace-pre-wrap font-mono text-xs ${hasError ? 'text-destructive' : 'text-foreground'}`}>
            {result.error || result.content}
          </div>
        )}
      </CollapsibleContent>
    </Collapsible>
  )
}
