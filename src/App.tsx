import { Navigate, Route, Routes } from 'react-router-dom'
import { isAuthenticated } from './utils/auth'
import ProtectedRoute from './components/ProtectedRoute'
import Layout from './components/Layout'
import LoginPage from './views/pages/LoginPage'
import SignupPage from './views/pages/SignupPage'
import DashboardPage from './views/pages/DashboardPage'
import DepositPage from './views/pages/DepositPage'
import ConvertPage from './views/pages/ConvertPage'
import PayoutPage from './views/pages/PayoutPage'
import TransactionsPage from './views/pages/TransactionsPage'
import TransactionDetailPage from './views/pages/TransactionDetailPage'

export default function App() {
  return (
    <Routes>
      <Route
        path="/login"
        element={isAuthenticated() ? <Navigate to="/" replace /> : <LoginPage />}
      />
      <Route
        path="/signup"
        element={isAuthenticated() ? <Navigate to="/" replace /> : <SignupPage />}
      />

      <Route element={<ProtectedRoute />}>
        <Route element={<Layout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/deposit" element={<DepositPage />} />
          <Route path="/convert" element={<ConvertPage />} />
          <Route path="/payout" element={<PayoutPage />} />
          <Route path="/transactions" element={<TransactionsPage />} />
          <Route path="/transactions/:id" element={<TransactionDetailPage />} />
        </Route>
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
