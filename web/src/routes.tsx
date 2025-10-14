import { createBrowserRouter } from "react-router-dom";
import DashboardView from "./views/DashboardView";
import NetworkingView from "./views/NetworkingView";
import NotFoundView from "./views/NotFoundView";
import AppLayout from "./layouts/AppLayout";
import NodesView from "./views/NodesView";
import KubernetesView from "./views/KubernetesView";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppLayout />,
    children: [
      { path: "/", element: <DashboardView /> },
      { path: "/dashboard", element: <DashboardView /> },
      { path: "/nodes", element: <NodesView /> },
      { path: "/kubernetes", element: <KubernetesView /> },
      { path: "/networking", element: <NetworkingView /> },
      { path: "*", element: <NotFoundView /> },
    ]
  },
]);