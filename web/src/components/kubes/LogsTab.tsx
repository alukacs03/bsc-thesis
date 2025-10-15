import Terminal from "../Terminal"

interface LogsTabProps {
    workloads?: { name: string; type: string; namespace: string; replicas: string; status: string; age: string; image: string; cluster: string; }[];
}

const LogsTab = ({ workloads }: LogsTabProps) => {
  return (
    <>
        {(!workloads || workloads.length === 0) && <p className="text-m text-slate-600">No workloads found in this cluster.</p>}
        {(workloads && workloads.length > 0) && (
            <>
                <select style={{ marginBottom: "1rem" }}>
                    {workloads?.map((workload) => (
                        <option key={workload.name} value={workload.name}>
                            {workload.name} ({workload.namespace})
                        </option>
                    ))}
                </select>
                <Terminal
                    title="Pod Logs"
                    commandHistory={[
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: Using the \"epoll\" event method",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: nginx/1.19.0",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: built by gcc 8.3.0 (Debian 8.3.0-6)",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: OS: Linux 5.4.0-104-generic",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: getrlimit(RLIMIT_NOFILE): 1048576:1048576",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: start worker processes",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:00:00 [info] 1#1: start worker process 25",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:05:00 [error] 25#25: *1 open() \"/usr/share/nginx/html/404.html\" failed (2: No such file or directory), client: 192.168.1.1",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 client: 192.168.1.1",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 server: example.com",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 request: \"GET /index.html HTTP/1.1\"",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 host: \"example.com\"",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 referrer: \"-\"",
                        "nginx-ingress-controller-5c689d4b7f-7xk9l - stdout - 2024/10/01 12:10:00 [info] 25#25: *2 client: 192.168.1.1"
                    ]}
                    refreshButtonAction={() => alert('Refresh logs clicked (would fetch latest logs)')}
                />
            </>
        )}
    </>

    )
}

export default LogsTab
