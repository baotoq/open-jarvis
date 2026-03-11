'use client'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { getConfig, updateConfig, type ModelConfig } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export default function SettingsPage() {
  const [form, setForm] = useState<ModelConfig>({
    baseURL: '',
    name: '',
    apiKey: '',
    systemPrompt: '',
  })
  const [status, setStatus] = useState<'idle' | 'loading' | 'saving' | 'saved' | 'error'>('loading')

  useEffect(() => {
    getConfig()
      .then(cfg => { setForm(cfg); setStatus('idle') })
      .catch(() => setStatus('error'))
  }, [])

  const handleChange = (field: keyof ModelConfig) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm(prev => ({ ...prev, [field]: e.target.value }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setStatus('saving')
    try {
      await updateConfig(form)
      setStatus('saved')
      setTimeout(() => setStatus('idle'), 2000)
    } catch {
      setStatus('error')
    }
  }

  return (
    <div className="max-w-lg mx-auto p-8">
      <div className="flex items-center gap-4 mb-6">
        <Link href="/" className="text-sm text-muted-foreground hover:text-foreground">← Back</Link>
        <h1 className="text-xl font-semibold">Model Settings</h1>
      </div>
      {status === 'loading' && <p className="text-muted-foreground">Loading...</p>}
      {status !== 'loading' && (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Base URL</label>
            <Input value={form.baseURL} onChange={handleChange('baseURL')} placeholder="http://localhost:11434/v1" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Model Name</label>
            <Input value={form.name} onChange={handleChange('name')} placeholder="llama3.2" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">API Key</label>
            <Input value={form.apiKey} onChange={handleChange('apiKey')} type="password" placeholder="sk-... (optional for Ollama)" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">System Prompt</label>
            <textarea
              value={form.systemPrompt}
              onChange={handleChange('systemPrompt')}
              rows={4}
              className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
              placeholder="You are Jarvis, a personal AI assistant..."
            />
          </div>
          <div className="flex items-center gap-3">
            <Button type="submit" disabled={status === 'saving'}>
              {status === 'saving' ? 'Saving...' : 'Save'}
            </Button>
            {status === 'saved' && <span className="text-sm text-green-600">Saved</span>}
            {status === 'error' && <span className="text-sm text-destructive">Error saving</span>}
          </div>
        </form>
      )}
    </div>
  )
}
