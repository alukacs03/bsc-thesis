import { AlertTriangle } from "lucide-react"

const NotFoundView = () => {
  return (
    <>
        <div className="flex flex-col items-center justify-center pt-20 bg-slate-100">
            <AlertTriangle width={64} height={64} className="text-red-600 mb-4"/>
            <h1 className="text-4xl font-bold text-slate-800 mb-4">404 - Not Found</h1>
            <p className="text-lg text-slate-600">The page you are looking for does not exist.</p>
        </div>
    </>
  )
}

export default NotFoundView
