import { useState } from 'react'
import EdgeNodes from './components/EdgeNodes'
import Events from './components/Events'
import Security from './components/Security'
import Alerts from './components/Alerts'

const tabs = ['Edge Nodes', 'Events', 'Security', 'Alerts'] as const
type Tab = (typeof tabs)[number]

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('Edge Nodes')

  return (
    <div className="min-h-screen bg-[#0f172a] text-[#f1f5f9]">
      <header className="border-b border-[#334155] bg-[#1e293b] px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold tracking-tight">Factory Platform</h1>
        <span className="text-sm text-[#94a3b8]">Distributed IoT Dashboard</span>
      </header>

      <nav className="border-b border-[#334155] bg-[#1e293b] px-6 flex gap-1">
        {tabs.map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-3 text-sm font-medium transition-colors relative ${
              activeTab === tab
                ? 'text-[#3b82f6]'
                : 'text-[#94a3b8] hover:text-[#f1f5f9]'
            }`}
          >
            {tab}
            {activeTab === tab && (
              <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-[#3b82f6]" />
            )}
          </button>
        ))}
      </nav>

      <main className="p-6">
        {activeTab === 'Edge Nodes' && <EdgeNodes />}
        {activeTab === 'Events' && <Events />}
        {activeTab === 'Security' && <Security />}
        {activeTab === 'Alerts' && <Alerts />}
      </main>
    </div>
  )
}

export default App
