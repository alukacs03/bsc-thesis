import { createBrowserRouter } from "react-router-dom";
import MainLayout from "./layouts/MainLayout";
import DashboardPage from "./pages/DashboardPage";
import Nodes from "./pages/Nodes";
import Approvals from "./pages/Approvals";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <MainLayout />,
    children: [
        { path: "/", element: <DashboardPage /> },
        { path: "/nodes", element: <Nodes /> },
        { path: "/approvals", element: <Approvals /> },
    ]
  },
]);