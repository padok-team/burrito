import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  createBrowserRouter,
  RouterProvider,
  Navigate
} from 'react-router-dom';

import ThemeProvider from '@/contexts/ThemeContext';

import Layout from '@/layout/Layout';
import Layers from '@/pages/Layers';
import Pulls from '@/pages/Pulls';
import Logs from '@/pages/Logs';
import Login from '@/pages/Login';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (
        failureCount,
        error: Error & { response?: { status: number } }
      ) => {
        // Don't retry on 401 errors
        if (error?.response?.status === 401) {
          window.location.href = '/login';
          return false;
        }
        return failureCount < 3;
      }
    },
    mutations: {
      onError: (error: Error & { response?: { status: number } }) => {
        if (error?.response?.status === 401) {
          window.location.href = '/login';
        }
      }
    }
  }
});
const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: <Navigate to="/layers" />
      },
      {
        path: 'layers',
        element: <Layers />
      },
      {
        path: 'pulls',
        element: <Pulls />
      },
      {
        path: 'logs/:namespace?/:layerId?/:runId?',
        element: <Logs />
      }
    ]
  },
  {
    path: '/login',
    element: <Login />
  },
  {
    path: '*',
    element: <Navigate to="/" />
  }
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
