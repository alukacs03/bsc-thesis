export default function Logo() {
  return (
    <div className="flex items-center space-x-3">
      <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
        <div className="w-4 h-4 bg-white rounded-sm"></div>
      </div>
      <div>
        <h1 className="text-xl text-slate-800">Gluon</h1>
        <p className="text-sm text-slate-600">Infrastructure Hub</p>
      </div>
    </div>
  );
}
