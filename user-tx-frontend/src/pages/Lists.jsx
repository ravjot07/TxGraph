import { useEffect, useState } from 'react'
import axios from 'axios'

export default function Lists() {
  const [users, setUsers]               = useState([])
  const [txns,  setTxns]                = useState([])

  // Filters
  const [userQuery, setUserQuery]       = useState('')
  const [minAmt, setMinAmt]             = useState('')
  const [maxAmt, setMaxAmt]             = useState('')
  const [currency, setCurrency]         = useState('')
  const [startDate, setStartDate]       = useState('')
  const [endDate, setEndDate]           = useState('')
  const [descQuery, setDescQuery]       = useState('')
  const [deviceQuery, setDeviceQuery]   = useState('')

  // Pagination
  const USERS_PER_PAGE = 5
  const TXNS_PER_PAGE  = 5
  const [userPage, setUserPage] = useState(1)
  const [txnPage,  setTxnPage]  = useState(1)

  useEffect(() => {
    axios.get('/api/users').then(res => setUsers(res.data))
    axios.get('/api/transactions').then(res => setTxns(res.data))
  }, [])

  // Filter logic
  const filteredUsers = users.filter(u => {
    const q = userQuery.toLowerCase()
    return (
      u.name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q) ||
      u.phone.toLowerCase().includes(q)
    )
  })

  const filteredTxns = txns.filter(t => {
    if (minAmt && t.amount < parseFloat(minAmt)) return false
    if (maxAmt && t.amount > parseFloat(maxAmt)) return false
    if (currency && t.currency !== currency)       return false
    const ts = new Date(t.timestamp)
    if (startDate && ts < new Date(startDate))     return false
    if (endDate   && ts > new Date(endDate))       return false
    if (descQuery && !t.description.toLowerCase().includes(descQuery.toLowerCase())) return false
    if (deviceQuery && (!t.deviceId || !t.deviceId.toLowerCase().includes(deviceQuery.toLowerCase()))) return false
    return true
  })

  const currencies = Array.from(new Set(txns.map(t => t.currency))).sort()

  // Pagination slices
  const userPageCount = Math.ceil(filteredUsers.length / USERS_PER_PAGE)
  const txnPageCount  = Math.ceil(filteredTxns.length  / TXNS_PER_PAGE)
  const displayUsers  = filteredUsers.slice((userPage-1)*USERS_PER_PAGE, userPage*USERS_PER_PAGE)
  const displayTxns   = filteredTxns .slice((txnPage-1)*TXNS_PER_PAGE,  txnPage*TXNS_PER_PAGE)

  return (
    <div className="container mx-auto px-6 py-12 space-y-12">
      {/* Users Section */}
      <section className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-semibold mb-4">Users</h2>
        <input
          type="text"
          placeholder="Search users by name, email or phone..."
          value={userQuery}
          onChange={e => { setUserQuery(e.target.value); setUserPage(1) }}
          className="w-full mb-4 border border-gray-300 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <div className="overflow-x-auto">
          <table className="w-full table-auto divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                {['ID','Name','Email','Phone'].map(h => (
                  <th key={h} className="px-4 py-2 text-left text-sm font-medium text-gray-700">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {displayUsers.length > 0 ? displayUsers.map(u => (
                <tr key={u.id}>
                  <td className="px-4 py-2 text-sm text-gray-600">{u.id}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{u.name}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{u.email}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{u.phone}</td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-center text-gray-500">
                    No users match your search.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        {/* User Pagination */}
        {userPageCount > 1 && (
          <div className="flex justify-center items-center mt-4 space-x-2">
            <button
              onClick={() => setUserPage(p => Math.max(p-1, 1))}
              disabled={userPage === 1}
              className="px-3 py-1 bg-gray-200 rounded disabled:opacity-50"
            >
              Prev
            </button>
            {Array.from({ length: userPageCount }, (_, i) => (
              <button
                key={i+1}
                onClick={() => setUserPage(i+1)}
                className={`px-3 py-1 rounded ${
                  userPage === i+1 ? 'bg-blue-600 text-white' : 'bg-gray-200'
                }`}
              >
                {i+1}
              </button>
            ))}
            <button
              onClick={() => setUserPage(p => Math.min(p+1, userPageCount))}
              disabled={userPage === userPageCount}
              className="px-3 py-1 bg-gray-200 rounded disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </section>

      {/* Transactions Section */}
      <section className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-semibold mb-4">Transactions</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4 mb-6">
          {[
            { label: 'Amount ≥',    type: 'number', value: minAmt,     setter: setMinAmt,    attrs: { step: '0.01' } },
            { label: 'Amount ≤',    type: 'number', value: maxAmt,     setter: setMaxAmt,    attrs: { step: '0.01' } },
            { label: 'Currency',    type: 'select', value: currency,   setter: setCurrency },
            { label: 'From Date',   type: 'date',   value: startDate,  setter: setStartDate },
            { label: 'To Date',     type: 'date',   value: endDate,    setter: setEndDate },
            { label: 'Device ID',   type: 'text',   value: deviceQuery,setter: setDeviceQuery },
            { label: 'Description', type: 'text',   value: descQuery,  setter: setDescQuery },
          ].map(({ label, type, value, setter, attrs }) => (
            <div key={label}>
              <label className="block mb-1 text-sm font-medium text-gray-700">
                {label}
              </label>
              {type === 'select' ? (
                <select
                  value={value}
                  onChange={e => { setter(e.target.value); setTxnPage(1) }}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500"
                >
                  <option value="">All</option>
                  {currencies.map(c => (
                    <option key={c} value={c}>{c}</option>
                  ))}
                </select>
              ) : (
                <input
                  type={type}
                  value={value}
                  onChange={e => { setter(e.target.value); setTxnPage(1) }}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500"
                  {...attrs}
                />
              )}
            </div>
          ))}
        </div>

        <div className="overflow-x-auto">
          <table className="w-full table-auto divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                {['ID','From','To','Amount','Currency','Timestamp','Device ID','Description']
                  .map(h => (
                    <th key={h} className="px-4 py-2 text-left text-sm font-medium text-gray-700">
                      {h}
                    </th>
                  ))
                }
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {displayTxns.length > 0 ? displayTxns.map(t => (
                <tr key={t.id}>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.id}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.fromUserId}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.toUserId}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.amount}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.currency}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">
                    {new Date(t.timestamp).toLocaleString()}
                  </td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.deviceId || '—'}</td>
                  <td className="px-4 py-2 text-sm text-gray-600">{t.description}</td>
                </tr>
              )) : (
                <tr>
                  <td colSpan={8} className="px-4 py-6 text-center text-gray-500">
                    No transactions match your filters.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Transaction Pagination */}
        {txnPageCount > 1 && (
          <div className="flex justify-center items-center mt-4 space-x-2">
            <button
              onClick={() => setTxnPage(p => Math.max(p-1, 1))}
              disabled={txnPage === 1}
              className="px-3 py-1 bg-gray-200 rounded disabled:opacity-50"
            >
              Prev
            </button>
            {Array.from({ length: txnPageCount }, (_, i) => (
              <button
                key={i+1}
                onClick={() => setTxnPage(i+1)}
                className={`px-3 py-1 rounded ${
                  txnPage === i+1 ? 'bg-blue-600 text-white' : 'bg-gray-200'
                }`}
              >
                {i+1}
              </button>
            ))}
            <button
              onClick={() => setTxnPage(p => Math.min(p+1, txnPageCount))}
              disabled={txnPage === txnPageCount}
              className="px-3 py-1 bg-gray-200 rounded disabled:opacity-50"
            >
              Next
            </button>
          </div>
        )}
      </section>
    </div>
  )
}
