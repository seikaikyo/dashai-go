import { useState, useEffect } from 'react'

interface NodeDetail {
  node_id: string
  location: string
  devices: number
  critical_vulns: number
  last_scan: string
  status: string
}

interface SecurityDashboard {
  summary: {
    nodes: number
    total_devices: number
    ot_devices: number
    it_devices: number
    critical_alerts: number
  }
  compliance: {
    iec_62443: number
    nist_csf: number
    iso_27001: number
    semi_e187: number
  }
  nodes_detail: NodeDetail[]
}

const statusColor: Record<string, string> = {
  secure: 'text-[#22c55e]',
  warning: 'text-[#eab308]',
  critical: 'text-[#ef4444]',
}

export default function Security() {
  const [data, setData] = useState<SecurityDashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch('/security/dashboard')
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        setData(await res.json())
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) return <p className="text-[#94a3b8]">Loading security data...</p>
  if (error) return <p className="text-[#ef4444]">Error: {error}</p>
  if (!data) return null

  const { summary, compliance, nodes_detail } = data

  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <SummaryCard label="Nodes" value={summary.nodes} />
        <SummaryCard label="Total Devices" value={summary.total_devices} />
        <SummaryCard label="OT Devices" value={summary.ot_devices} />
        <SummaryCard label="IT Devices" value={summary.it_devices} />
        <SummaryCard
          label="Critical Alerts"
          value={summary.critical_alerts}
          danger={summary.critical_alerts > 0}
        />
      </div>

      {/* Compliance scores */}
      <div className="rounded-lg border border-[#334155] bg-[#1e293b] p-5">
        <h2 className="font-semibold mb-4">Compliance Scores</h2>
        <div className="space-y-4">
          <ComplianceBar label="IEC 62443" value={compliance.iec_62443} />
          <ComplianceBar label="NIST CSF" value={compliance.nist_csf} />
          <ComplianceBar label="ISO 27001" value={compliance.iso_27001} />
          <ComplianceBar label="SEMI E187" value={compliance.semi_e187} />
        </div>
      </div>

      {/* Nodes detail table */}
      <div className="rounded-lg border border-[#334155] bg-[#1e293b] overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[#334155] text-left text-[#94a3b8]">
              <th className="px-4 py-3 font-medium">Node</th>
              <th className="px-4 py-3 font-medium">Location</th>
              <th className="px-4 py-3 font-medium">Devices</th>
              <th className="px-4 py-3 font-medium">Critical Vulns</th>
              <th className="px-4 py-3 font-medium">Last Scan</th>
              <th className="px-4 py-3 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            {nodes_detail.map((node) => (
              <tr key={node.node_id} className="border-b border-[#334155] last:border-0">
                <td className="px-4 py-3 font-mono text-xs">{node.node_id}</td>
                <td className="px-4 py-3">{node.location}</td>
                <td className="px-4 py-3">{node.devices}</td>
                <td className="px-4 py-3">
                  <span className={node.critical_vulns > 0 ? 'text-[#ef4444] font-medium' : ''}>
                    {node.critical_vulns}
                  </span>
                </td>
                <td className="px-4 py-3 text-xs text-[#94a3b8]">
                  {new Date(node.last_scan).toLocaleString()}
                </td>
                <td className="px-4 py-3">
                  <span className={statusColor[node.status] ?? 'text-[#94a3b8]'}>
                    {node.status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function SummaryCard({
  label,
  value,
  danger = false,
}: {
  label: string
  value: number
  danger?: boolean
}) {
  return (
    <div className="rounded-lg border border-[#334155] bg-[#1e293b] p-4">
      <p className="text-sm text-[#94a3b8]">{label}</p>
      <p className={`text-2xl font-semibold mt-1 ${danger ? 'text-[#ef4444]' : ''}`}>
        {value}
      </p>
    </div>
  )
}

function ComplianceBar({ label, value }: { label: string; value: number }) {
  const color =
    value >= 80 ? 'bg-[#22c55e]' : value >= 60 ? 'bg-[#eab308]' : 'bg-[#ef4444]'

  return (
    <div>
      <div className="flex justify-between text-sm mb-1">
        <span className="text-[#94a3b8]">{label}</span>
        <span className="font-medium">{value}%</span>
      </div>
      <div className="h-2 rounded-full bg-[#334155]">
        <div
          className={`h-2 rounded-full ${color} transition-all`}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  )
}
