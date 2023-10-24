import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createBrowserRouter,
  RouterProvider,
  Navigate,
} from "react-router-dom";

import ThemeProvider from "@/contexts/ThemeContext";

import Layout from "@/layout/Layout";
import Layers from "@/pages/layers/Layers";
import Pulls from "@/pages/pulls/Pulls";

const queryClient = new QueryClient();
const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Navigate to="/layers" />,
      },
      {
        path: "layers",
        element: <Layers />,
      },
      {
        path: "pulls",
        element: <Pulls />,
      },
    ],
  },
  {
    path: "*",
    element: <Navigate to="/" />,
  },
]);

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <RouterProvider router={router} />
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
