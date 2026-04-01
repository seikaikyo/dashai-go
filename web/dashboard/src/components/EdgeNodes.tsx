import { useState, useEffect, useCallback } from 'react'

interface EdgeNode {
  id: string
  type: string
  location: string
  status: string
  capabilities: string[]
  last_heartbeat: string
}

const statusColor: Record<string, string> = {
  online: 'bg-[#22c55e] text-[#0f172a]',
  degraded: 'bg-[#eab308] text-[#0f172a]',
  offline: 'bg-[#ef4444] text-white',
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const sec = Math.floor(diff / 1000)
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  return `${hr}h ago`
}

export default function EdgeNodes() {
  const [nodes, setNodes] = useState<EdgeNode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchNodes = useCallback(async () => {
    try {
      const res = await fetch('/edge/nodes')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      setNodes(Array.isArray(data) ? data : data.nodes ?? [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchNodes()
    const id = setInterval(fetchNodes, 10000)
    return () => clearInterval(id)
  }, [fetchNodes])

  if (loading) {
    return <p className="text-[#94a3b8]">Loading edge nodes...</p>
  }

  if (error) {
    return <p className="text-[#ef4444]">Error: {error}</p>
  }

  if (nodes.length === 0) {
    return <p className="text-[#94a3b8]">No edge nodes found.</p>
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {nodes.map((node) => (
        <div
          key={node.id}
          className="rounded-lg border border-[#334155] bg-[#1e293b] p-5"
        >
          <div className="flex items-start justify-between mb-3">
            <div>
              <h3 className="font-semibold text-[#f1f5f9]">{node.id}</h3>
              <p className="text-sm text-[#94a3b8]">{node.type}</p>
            </div>
            <span
              className={`rounded px-2 py-0.5 text-xs font-medium ${
                statusColor[node.status] ?? 'bg-[#334155] text-[#94a3b8]'
              }`}
            >
              {node.status}
            </span>
          </div>

          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-[#94a3b8]">Location</span>
              <span>{node.location}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-[#94a3b8]">Heartbeat</span>
              <span>{relativeTime(node.last_heartbeat)}</span>
            </div>
          </div>

          {node.capabilities?.length > 0 && (
            <div className="mt-3 flex flex-wrap gap-1">
              {node.capabilities.map((cap) => (
                <span
                  key={cap}
                  className="rounded bg-[#334155] px-2 py-0.5 text-xs text-[#94a3b8]"
                >
                  {cap}
                </span>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
