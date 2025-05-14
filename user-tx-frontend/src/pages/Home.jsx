import { Link } from 'react-router-dom'

const FEATURES = [
  { title: 'Add a User',       to: '/add-user',       color: 'bg-green-500' },
  { title: 'Add a Transaction', to: '/add-transaction', color: 'bg-blue-500' },
  { title: 'View Lists',       to: '/lists',          color: 'bg-indigo-500' },
  { title: 'Graph View',       to: '/graph',          color: 'bg-purple-500' },
  { title: 'Shortest Path',    to: '/analytics',      color: 'bg-teal-500' },
  { title: 'Export JSON/CSV',  to: '/export',         color: 'bg-gray-700' },
]

export default function Home() {
  return (
    <div className="container mx-auto px-6 py-12">
      <div className="text-center mb-12">
        <h1 className="text-4xl font-extrabold text-gray-800 mb-4">
          User &amp; Transaction Dashboard
        </h1>
        <p className="text-lg text-gray-600 max-w-2xl mx-auto">
          Manage users, track transactions, and explore relationships visually—all in one place.
        </p>
      </div>

      <section className="grid gap-6 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 mb-12">
        {FEATURES.map(({ title, to, color }) => (
          <Link to={to} key={title}>
            <div className="h-full flex flex-col justify-between p-6 bg-white rounded-lg shadow hover:shadow-lg transition-shadow duration-200">
              <div>
                <h3 className="text-xl font-semibold text-gray-800 mb-2">
                  {title}
                </h3>
                <p className="text-gray-500 text-sm">
                  {descriptionFor(title)}
                </p>
              </div>
              <span
                className={`${color} inline-block mt-4 px-4 py-2 text-white font-medium rounded-full text-sm`}
              >
                Go
              </span>
            </div>
          </Link>
        ))}
      </section>

      <section className="flex flex-wrap justify-center gap-4">
        <Link
          to="/add-user"
          className="px-6 py-3 bg-green-600 text-white rounded-lg shadow hover:bg-green-700 transition-colors"
        >
          + New User
        </Link>
        <Link
          to="/add-transaction"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg shadow hover:bg-blue-700 transition-colors"
        >
          + New Transaction
        </Link>
      </section>
    </div>
  )
}

function descriptionFor(title) {
  switch (title) {
    case 'Add a User':
      return 'Quickly register a new user into the graph.'
    case 'Add a Transaction':
      return 'Record money movement between users.'
    case 'View Lists':
      return 'Browse all users and transactions in a searchable table.'
    case 'Graph View':
      return 'See relationships laid out in an interactive network.'
    case 'Shortest Path':
      return 'Find connection chains between any two users.'
    case 'Export JSON/CSV':
      return 'Download the full graph for offline analysis.'
    default:
      return ''
  }
}
