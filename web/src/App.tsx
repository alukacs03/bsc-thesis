import { RouterProvider } from 'react-router-dom'
import { router } from './routes'
import { AuthProvider } from './hooks/useAuth'
import { ToastProvider } from './components/ToastProvider'

function App() {
    return (
        <AuthProvider>
            <ToastProvider />
            <RouterProvider router={router} />
        </AuthProvider>
    )
}

export default App
