import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import Navbar from './components/Navbar'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
import PostDetail from './pages/PostDetail'
import CreatePost from './pages/CreatePost'
import EditPost from './pages/EditPost'
import Profile from './pages/Profile'
import AdminPanel from './pages/AdminPanel'
import Tools from './pages/Tools'

function PrivateRoute({ children }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="loading">Loading…</div>
  return user ? children : <Navigate to="/login" replace />
}

function AdminRoute({ children }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="loading">Loading…</div>
  if (!user) return <Navigate to="/login" replace />
  if (user.role !== 'admin') return <Navigate to="/" replace />
  return children
}

export default function App() {
  return (
    <>
      <Navbar />
      <Routes>
        <Route path="/"            element={<Home />} />
        <Route path="/login"       element={<Login />} />
        <Route path="/register"    element={<Register />} />
        <Route path="/posts/:id"   element={<PostDetail />} />
        <Route path="/tools"       element={<Tools />} />
        <Route path="/submit"      element={<PrivateRoute><CreatePost /></PrivateRoute>} />
        <Route path="/posts/:id/edit" element={<PrivateRoute><EditPost /></PrivateRoute>} />
        <Route path="/user/:id"    element={<Profile />} />
        <Route path="/profile"     element={<PrivateRoute><Profile /></PrivateRoute>} />
        <Route path="/admin"       element={<AdminRoute><AdminPanel /></AdminRoute>} />
      </Routes>
    </>
  )
}
