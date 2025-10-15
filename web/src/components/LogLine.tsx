
const LogLine = ({ line } : { line: string }) => {
  return (
    <div className={`text-sm ${line.startsWith('$') ? 'text-green-400' : 'text-slate-300'}`}>
      {line}
    </div>
  )
}

export default LogLine
