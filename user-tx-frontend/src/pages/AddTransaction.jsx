import { useState, useEffect } from 'react'
import axios from 'axios'
import { useNavigate } from 'react-router-dom'

export default function AddTransaction() {
  const [users, setUsers]             = useState([])
  const [fromId, setFromId]           = useState('')
  const [toId, setToId]               = useState('')
  const [amount, setAmount]           = useState('')
  const [currency, setCurrency]       = useState('USD')
  const [timestamp, setTimestamp]     = useState('')
  const [description, setDescription] = useState('')
  const [deviceId, setDeviceId]       = useState('')
  const [error, setError]             = useState(null)
  const [success, setSuccess]         = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    axios.get('/api/users')
      .then(res => setUsers(res.data))
      .catch(err => {
        console.error('Failed to load users:', err)
        setError('Error loading users.')
      })

    setTimestamp(new Date().toISOString().slice(0,16))
  }, [])

  const handleSubmit = async e => {
    e.preventDefault()
    setError(null)

    try {
      await axios.post('/api/transactions', {
        fromUserId:   Number(fromId),
        toUserId:     Number(toId),
        amount:       Number(amount),
        currency,
        timestamp,
        description,
        deviceId
      })

      setSuccess(true)
      // reset form
      setFromId('')
      setToId('')
      setAmount('')
      setCurrency('USD')
      setDescription('')
      setDeviceId('')
      setTimestamp(new Date().toISOString().slice(0,16))

      setTimeout(() => navigate('/lists'), 1000)
    } catch (err) {
      console.error('Transaction creation error:', err)
      setError(err.response?.data?.message || 'Failed to create transaction.')
    }
  }

  return (
    <div className="container mx-auto px-6 py-8">
      <div className="max-w-md mx-auto bg-white shadow-lg rounded-lg p-6">
        <h2 className="text-2xl font-semibold text-gray-800 mb-6">
          Add Transaction
        </h2>

        {error && (
          <div className="bg-red-50 text-red-700 border border-red-200 px-4 py-2 rounded mb-4">
            {error}
          </div>
        )}
        {success && (
          <div className="bg-green-50 text-green-700 border border-green-200 px-4 py-2 rounded mb-4">
            Transaction created! Redirecting…
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-gray-700 mb-1">From User</label>
            <select
              value={fromId}
              onChange={e => setFromId(e.target.value)}
              required
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            >
              <option value="">Select a user</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>
                  {u.name} (#{u.id})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-gray-700 mb-1">To User</label>
            <select
              value={toId}
              onChange={e => setToId(e.target.value)}
              required
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            >
              <option value="">Select a user</option>
              {users.map(u => (
                <option key={u.id} value={u.id}>
                  {u.name} (#{u.id})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="block text-gray-700 mb-1">Amount</label>
            <input
              type="number"
              step="0.01"
              value={amount}
              onChange={e => setAmount(e.target.value)}
              required
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            />
          </div>

          <div>
            <label className="block text-gray-700 mb-1">Currency</label>
            <input
              value={currency}
              onChange={e => setCurrency(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            />
          </div>

          <div>
            <label className="block text-gray-700 mb-1">Timestamp</label>
            <input
              type="datetime-local"
              value={timestamp}
              onChange={e => setTimestamp(e.target.value)}
              required
              min={new Date().toISOString().slice(0,16)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            />
          </div>

          <div>
            <label className="block text-gray-700 mb-1">Description</label>
            <input
              value={description}
              onChange={e => setDescription(e.target.value)}
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            />
          </div>

          <div>
            <label className="block text-gray-700 mb-1">Device ID</label>
            <input
              value={deviceId}
              onChange={e => setDeviceId(e.target.value)}
              placeholder="e.g., device-123"
              className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring focus:border-blue-300"
            />
          </div>

          <button
            type="submit"
            className="w-full bg-green-600 hover:bg-green-700 text-white font-medium rounded px-3 py-2 transition"
          >
            Create Transaction
          </button>
        </form>
      </div>
    </div>
  )
}
