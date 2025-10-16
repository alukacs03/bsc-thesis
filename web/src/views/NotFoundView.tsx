import { AlertTriangle } from "lucide-react"

const NotFoundView = () => {
  return (
    <>
        <div className="flex flex-col items-center justify-center pt-20 bg-slate-100 min-h-screen">
            <AlertTriangle width={64} height={64} className="text-red-600 mb-4"/>
            <h1 className="text-4xl font-bold text-slate-800 mb-4">404 - Not Found</h1>
            <p className="text-lg text-slate-600">The page you are looking for does not exist.</p>
            <div className="mt-6">
                <a href="/dashboard" className="text-blue-600 hover:text-blue-800 underline">
                    Go back to Dashboard
                </a>
            </div>
            <div className="text-center">
                <p className="text-sm text-slate-600 mt-10">
                Don't have an account?{' '}
                <a
                  href="/register"
                  className="text-blue-600 hover:text-blue-700 underline"
                >
                  Request Access
                </a>
              </p>
            </div>
        </div>

    </>
  )
}

export default NotFoundView
