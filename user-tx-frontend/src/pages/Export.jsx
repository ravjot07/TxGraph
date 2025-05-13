import React from 'react'

export default function ExportPage() {
  return (
    <div className="space-y-4 max-w-sm mx-auto text-center">
      <h2 className="text-xl font-semibold">Export Graph Data</h2>
      <a
        href="/api/export/json"
        className="block px-4 py-2 bg-yellow-500 text-white rounded"
        download="graph.json"
      >
        Download JSON
      </a>
      <a
        href="/api/export/csv"
        className="block px-4 py-2 bg-yellow-700 text-white rounded"
        download="graph.csv"
      >
        Download CSV
      </a>
    </div>
  )
}
