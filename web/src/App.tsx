import { Routes, Route } from 'react-router-dom'
import UserLayout from './layouts/UserLayout'
import AdminLayout from './layouts/AdminLayout'
import ProtectedRoute from './components/ProtectedRoute'
import Landing from './pages/Landing'
import AgentMarket from './pages/AgentMarket'
import AgentDetail from './pages/AgentDetail'
import Chat from './pages/Chat'
import Dashboard from './pages/Dashboard'
import Overview from './pages/admin/Overview'
import AgentList from './pages/admin/AgentList'
import KnowledgeBaseList from './pages/admin/KnowledgeBaseList'
import KnowledgeBaseDetail from './pages/admin/KnowledgeBaseDetail'

export default function App() {
  return (
    <Routes>
      <Route element={<UserLayout />}>
        <Route path="/" element={<Landing />} />
        <Route path="/agents" element={<AgentMarket />} />
        <Route path="/agents/:id" element={<AgentDetail />} />
        <Route path="/chat/:agentId" element={<ProtectedRoute><Chat /></ProtectedRoute>} />
        <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      </Route>
      <Route element={<ProtectedRoute requireAdmin><AdminLayout /></ProtectedRoute>}>
        <Route path="/admin" element={<Overview />} />
        <Route path="/admin/agents" element={<AgentList />} />
        <Route path="/admin/knowledge-bases" element={<KnowledgeBaseList />} />
        <Route path="/admin/knowledge-bases/:id" element={<KnowledgeBaseDetail />} />
      </Route>
    </Routes>
  )
}
