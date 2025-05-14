import { Outlet, NavLink } from 'react-router-dom'

const NAV_ITEMS = [
  { label: 'Home',                  to: '/' },
  { label: 'Add User',              to: '/add-user' },
  { label: 'Add Transaction',       to: '/add-transaction' },
  { label: 'View Lists',            to: '/lists' },
  { label: 'Graph View',            to: '/graph' },
  { label: 'Shortest Path',         to: '/analytics' },
  { label: 'Transaction Clusters',  to: '/transaction-clusters' },
  { label: 'Export JSON/CSV',       to: '/export' },
]

export default function App() {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <header className="bg-white shadow">
        <div className="container mx-auto px-6 py-4 flex flex-col sm:flex-row items-center justify-between">
          <div className="flex items-center gap-3 mb-2 sm:mb-0">
            <img src="src/assets/logo.png" alt="Logo" className="w-12 h-12" />
            <h1 className="text-2xl font-bold text-blue-600">
              User & Transaction Dashboard
            </h1>
          </div>
          <nav className="flex flex-wrap gap-2">
            {NAV_ITEMS.map(({ label, to }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `px-3 py-2 rounded-md text-sm font-medium ${
                    isActive
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:bg-gray-100'
                  }`
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>

      <main className="flex-1 container mx-auto px-6 py-8">
        <Outlet />
      </main>

      <footer className="bg-white border-t">
        <div className="container mx-auto px-6 py-4 text-center text-sm text-gray-500">
          Flagright Intern Assignment
        </div>
      </footer>
    </div>
  )
}
