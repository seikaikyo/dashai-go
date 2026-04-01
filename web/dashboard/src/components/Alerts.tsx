import { useState, useEffect, useCallback } from 'react'

interface Alert {
  id: string
  severity: string
  type: string
  message: string
  node_id: string
  device_ip: string
  technique: string
  acknowledged: boolean
}

const severityColor: Record<string, string> = {
  info: 'bg-[#3b82f6] text-white',
  warning: 'bg-[#eab308] text-[#0f172a]',
  critical: 'bg-[#ef4444] text-white',
  high: 'bg-[#f97316] text-white',
  medium: 'bg-[#eab308] text-[#0f172a]',
  low: 'bg-[#3b82f6] text-white',
}

const severityOptions = ['all', 'critical', 'high', 'warning', 'medium', 'low', 'info']

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState('all')
  const [acking, setAcking] = useState<string | null>(null)

  const fetchAlerts = useCallback(async () => {
    try {
      const res = await fetch('/security/alerts')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      setAlerts(Array.isArray(data) ? data : data.alerts ?? [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchAlerts()
  }, [fetchAlerts])

  async function handleAck(id: string) {
    setAcking(id)
    try {
      const res = await fetch(`/security/alerts/${id}/ack`, { method: 'POST' })
      if (res.ok) {
        setAlerts((prev) =>
          prev.map((a) => (a.id === id ? { ...a, acknowledged: true } : a))
        )
      }
    } catch {
      // ack failed silently
    } finally {
      setAcking(null)
    }
  }

  const filtered = filter === 'all' ? alerts : alerts.filter((a) => a.severity === filter)

  if (loading) return <p className="text-[#94a3b8]">Loading alerts...</p>
  if (error) return <p className="text-[#ef4444]">Error: {error}</p>

  return (
    <div>
      {/* Filter */}
      <div className="flex items-center gap-3 mb-4">
        <label className="text-sm text-[#94a3b8]">Filter by severity:</label>
        <select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="rounded border border-[#334155] bg-[#1e293b] px-3 py-1.5 text-sm text-[#f1f5f9] focus:outline-none focus:border-[#3b82f6]"
        >
          {severityOptions.map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
        <span className="text-sm text-[#94a3b8] ml-auto">
          {filtered.length} alert{filtered.length !== 1 ? 's' : ''}
        </span>
      </div>

      {/* Alerts table */}
      <div className="rounded-lg border border-[#334155] bg-[#1e293b] overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[#334155] text-left text-[#94a3b8]">
              <th className="px-4 py-3 font-medium">Severity</th>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Message</th>
              <th className="px-4 py-3 font-medium">Node</th>
              <th className="px-4 py-3 font-medium">Device IP</th>
              <th className="px-4 py-3 font-medium">Technique</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Action</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={8} className="px-4 py-8 text-center text-[#94a3b8]">
                  No alerts found.
                </td>
              </tr>
            )}
            {filtered.map((alert) => (
              <tr key={alert.id} className="border-b border-[#334155] last:border-0">
                <td className="px-4 py-3">
                  <span
                    className={`rounded px-2 py-0.5 text-xs font-medium ${
                      severityColor[alert.severity] ?? 'bg-[#334155] text-[#94a3b8]'
                    }`}
                  >
                    {alert.severity}
                  </span>
                </td>
                <td className="px-4 py-3">{alert.type}</td>
                <td className="px-4 py-3 max-w-xs truncate">{alert.message}</td>
                <td className="px-4 py-3 font-mono text-xs">{alert.node_id}</td>
                <td className="px-4 py-3 font-mono text-xs">{alert.device_ip}</td>
                <td className="px-4 py-3 text-xs">{alert.technique}</td>
                <td className="px-4 py-3">
                  {alert.acknowledged ? (
                    <span className="text-[#22c55e] text-xs">ACK</span>
                  ) : (
                    <span className="text-[#eab308] text-xs">UNACK</span>
                  )}
                </td>
                <td className="px-4 py-3">
                  {!alert.acknowledged && (
                    <button
                      onClick={() => handleAck(alert.id)}
                      disabled={acking === alert.id}
                      className="rounded bg-[#334155] px-3 py-1 text-xs text-[#f1f5f9] hover:bg-[#475569] disabled:opacity-50 transition-colors"
                    >
                      {acking === alert.id ? '...' : 'Acknowledge'}
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
