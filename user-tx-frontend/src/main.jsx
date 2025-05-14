import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'

import App from './App'
import Home from './pages/Home'
import AddUser from './pages/AddUser'
import AddTransaction from './pages/AddTransaction'
import Lists from './pages/Lists'
import GraphView from './pages/GraphView'
import Analytics from './pages/Analytics'
import ExportPage from './pages/Export'
import TransactionClusters from './pages/TransactionClusters'
import './index.css'


ReactDOM.createRoot(document.getElementById('root')).render(
  <BrowserRouter>
    <Routes>
      <Route path="/" element={<App />}>
        <Route index                          element={<Home />} />
        <Route path="add-user"               element={<AddUser />} />
        <Route path="add-transaction"        element={<AddTransaction />} />
        <Route path="lists"                  element={<Lists />} />
        <Route path="graph"                  element={<GraphView />} />
        <Route path="analytics"              element={<Analytics />} />
        <Route path="transaction-clusters"   element={<TransactionClusters />} />
        <Route path="export"                 element={<ExportPage />} />
      </Route>
    </Routes>
  </BrowserRouter>
)
