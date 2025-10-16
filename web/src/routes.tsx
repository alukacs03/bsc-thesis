import { createBrowserRouter } from "react-router-dom";
import DashboardView from "./views/DashboardView";
import NetworkingView from "./views/NetworkingView";
import NotFoundView from "./views/NotFoundView";
import AppLayout from "./layouts/AppLayout";
import NodesView from "./views/NodesView";
import KubernetesView from "./views/KubernetesView";
import NodeManagementView from "./views/NodeManagementView";
import LoginView from "./views/LoginView";
import RegisterView from "./views/RegisterView";
import ApprovalsView from "./views/ApprovalsView";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardView /> },
      { path: "/", element: <DashboardView /> },
      { path: "/dashboard", element: <DashboardView /> },
      { path: "/nodes", element: <NodesView /> },
      { path: "/nodes/:nodeId", element: <NodeManagementView /> },
      { path: "/kubernetes", element: <KubernetesView /> },
      { path: "/networking", element: <NetworkingView /> },
      { path: "/approvals", element: <ApprovalsView /> },
    ]
  },
  { path: "/login", element: <LoginView /> },
  { path: "/register", element: <RegisterView /> },
  { path: "*", element: <NotFoundView /> },
]);