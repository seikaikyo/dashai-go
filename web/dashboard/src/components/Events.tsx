import { useState, useEffect, useRef, useCallback } from 'react'

interface EventItem {
  id?: string
  timestamp: string
  source: string
  type: string
  severity: string
  data?: Record<string, unknown>
}

interface EventStats {
  total_events: number
  by_type: Record<string, number>
  by_severity: Record<string, number>
}

const severityColor: Record<string, string> = {
  info: 'bg-[#3b82f6] text-white',
  warning: 'bg-[#eab308] text-[#0f172a]',
  critical: 'bg-[#ef4444] text-white',
}

export default function Events() {
  const [events, setEvents] = useState<EventItem[]>([])
  const [stats, setStats] = useState<EventStats | null>(null)
  const [wsStatus, setWsStatus] = useState<'connecting' | 'open' | 'closed'>('connecting')
  const wsRef = useRef<WebSocket | null>(null)

  const fetchStats = useCallback(async () => {
    try {
      const res = await fetch('/events/stats')
      if (res.ok) {
        setStats(await res.json())
      }
    } catch {
      // stats fetch is non-critical
    }
  }, [])

  useEffect(() => {
    fetchStats()
    const id = setInterval(fetchStats, 15000)
    return () => clearInterval(id)
  }, [fetchStats])

  useEffect(() => {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${proto}//${location.host}/events/stream`)
    wsRef.current = ws

    ws.onopen = () => setWsStatus('open')
    ws.onclose = () => setWsStatus('closed')
    ws.onerror = () => setWsStatus('closed')

    ws.onmessage = (msg) => {
      try {
        const evt: EventItem = JSON.parse(msg.data)
        setEvents((prev) => [evt, ...prev].slice(0, 100))
      } catch {
        // skip malformed messages
      }
    }

    return () => {
      ws.close()
    }
  }, [])

  return (
    <div>
      {/* Stats cards */}
      {stats && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-6">
          <StatCard label="Total Events" value={stats.total_events} />
          {Object.entries(stats.by_severity).map(([sev, count]) => (
            <StatCard key={sev} label={`${sev} events`} value={count} />
          ))}
        </div>
      )}

      {/* Connection status */}
      <div className="flex items-center gap-2 mb-4 text-sm text-[#94a3b8]">
        <span
          className={`inline-block h-2 w-2 rounded-full ${
            wsStatus === 'open'
              ? 'bg-[#22c55e]'
              : wsStatus === 'connecting'
                ? 'bg-[#eab308]'
                : 'bg-[#ef4444]'
          }`}
        />
        WebSocket {wsStatus}
        <span className="ml-auto">{events.length} events</span>
      </div>

      {/* Event log */}
      <div className="rounded-lg border border-[#334155] bg-[#1e293b] overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[#334155] text-left text-[#94a3b8]">
              <th className="px-4 py-3 font-medium">Time</th>
              <th className="px-4 py-3 font-medium">Source</th>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Severity</th>
              <th className="px-4 py-3 font-medium">Data</th>
            </tr>
          </thead>
          <tbody>
            {events.length === 0 && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-[#94a3b8]">
                  Waiting for events...
                </td>
              </tr>
            )}
            {events.map((evt, i) => (
              <tr key={i} className="border-b border-[#334155] last:border-0">
                <td className="px-4 py-2 font-mono text-xs whitespace-nowrap">
                  {new Date(evt.timestamp).toLocaleTimeString()}
                </td>
                <td className="px-4 py-2">{evt.source}</td>
                <td className="px-4 py-2">{evt.type}</td>
                <td className="px-4 py-2">
                  <span
                    className={`rounded px-2 py-0.5 text-xs font-medium ${
                      severityColor[evt.severity] ?? 'bg-[#334155] text-[#94a3b8]'
                    }`}
                  >
                    {evt.severity}
                  </span>
                </td>
                <td className="px-4 py-2 text-xs text-[#94a3b8] max-w-xs truncate">
                  {evt.data ? JSON.stringify(evt.data) : '-'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg border border-[#334155] bg-[#1e293b] p-4">
      <p className="text-sm text-[#94a3b8]">{label}</p>
      <p className="text-2xl font-semibold mt-1">{value.toLocaleString()}</p>
    </div>
  )
}
