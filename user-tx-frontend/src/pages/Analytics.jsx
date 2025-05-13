import React, { useState, useEffect, useRef } from 'react'
import axios from 'axios'
import cytoscape from 'cytoscape'

export default function Analytics() {
  const [users, setUsers]     = useState([])
  const [fromID, setFromID]   = useState('')
  const [toID, setToID]       = useState('')
  const [error, setError]     = useState(null)
  const cyRef                 = useRef(null)
  const [cy, setCy]           = useState(null)

  useEffect(() => {
    if (cyRef.current && !cy) {
      const instance = cytoscape({
        container: cyRef.current,
        elements: [],
        style: [
          {
            selector: 'node',
            style: {
              label: 'data(label)',
              'background-color': '#6FB1FC',
              shape: 'ellipse',
            },
          },
          {
            selector: 'edge',
            style: {
              label: 'data(label)',
              'font-size': '10px',
              'text-rotation': 'autorotate',
              'text-margin-y': '-4px',
              'text-background-color': '#ffffff',
              'text-background-opacity': 0.8,
              'text-background-padding': '2px',
              'line-color': '#555',
              width: 2,
            },
          },
          {
            selector: '.highlight',
            style: {
              'line-color': 'red',
              width: 4,
            },
          },
        ],
        layout: { name: 'breadthfirst' },
      })
      setCy(instance)
    }
  }, [cyRef, cy])

  useEffect(() => {
    axios
      .get('/api/users')
      .then(res => setUsers(res.data))
      .catch(err => {
        console.error('Failed to load users:', err)
        setError('Error loading user list.')
      })
  }, [])

  const handlePath = async e => {
    e.preventDefault()
    setError(null)

    if (!fromID || !toID) {
      setError('Please select both users.')
      return
    }

    try {
      const { data } = await axios.get(
        `/api/analytics/shortest-path/users/${fromID}/${toID}`
      )
      const path = data.path

      if (!path || path.length === 0) {
        setError('No path found between those users.')
        cy.elements().remove()
        return
      }

      const nodeElems = path.map(n => {
        const prefix = n.type[0].toLowerCase()
        const id     = `${prefix}${n.id}`
        const label  = n.type === 'User' ? n.name : `Txn #${n.id}`
        return { data: { id, label } }
      })

      const edgeElems = []
      for (let i = 1; i < path.length; i++) {
        const prev = path[i - 1]
        const curr = path[i]
        const pid  = `${prev.type[0].toLowerCase()}${prev.id}`
        const cid  = `${curr.type[0].toLowerCase()}${curr.id}`
        edgeElems.push({
          data: { id: `e_${pid}_${cid}`, source: pid, target: cid, label: 'path' },
          classes: 'highlight',
        })
      }

      cy.elements().remove()
      cy.add([...nodeElems, ...edgeElems])
      cy.layout({ name: 'breadthfirst' }).run()
      cy.fit()
    } catch (err) {
      console.error('Analytics error:', err)
      setError(
        err.response?.data === 'no path found\n'
          ? 'No path found between those users.'
          : 'Error computing path.'
      )
      cy.elements().remove()
    }
  }

  return (
    <div className="container mx-auto px-6 py-12 max-w-lg space-y-6">
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-semibold mb-4 text-center">
          Shortest Path Analytics
        </h2>
        {error && (
          <div className="bg-red-50 text-red-700 px-4 py-2 rounded mb-4">
            {error}
          </div>
        )}
        <form onSubmit={handlePath} className="space-y-4">
          <div>
            <label className="block mb-1 text-sm font-medium text-gray-700">
              From User
            </label>
            <select
              value={fromID}
              onChange={e => setFromID(e.target.value)}
              required
              className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-purple-500"
            >
              <option value="">Select user…</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>
                  {u.name} (#{u.id})
                </option>
              ))}
            </select>
          </div>
          <div>
            <label className="block mb-1 text-sm font-medium text-gray-700">
              To User
            </label>
            <select
              value={toID}
              onChange={e => setToID(e.target.value)}
              required
              className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-purple-500"
            >
              <option value="">Select user…</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>
                  {u.name} (#{u.id})
                </option>
              ))}
            </select>
          </div>
          <button
            type="submit"
            className="w-full bg-purple-600 hover:bg-purple-700 text-white font-medium py-2 rounded-lg transition"
          >
            Compute Path
          </button>
        </form>
      </div>

      <div
        ref={cyRef}
        className="bg-white rounded-lg shadow h-[400px] border border-gray-200"
      />
    </div>
  )
}
