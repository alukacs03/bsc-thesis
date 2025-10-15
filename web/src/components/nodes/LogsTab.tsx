import Terminal from "../Terminal"

interface LogsTabProps {
    logs?: string[];
}

const LogsTab = ({ logs }: LogsTabProps) => {
  return (
        <>
            <Terminal
                title="Agent Logs"
                commandHistory={[
                    ...(logs ? logs : ["No logs available."])
                ]}
                refreshButtonAction={() => alert('Refresh logs clicked (would fetch latest logs)')}
            />
        </>
    )
}

export default LogsTab