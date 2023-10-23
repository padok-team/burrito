import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import ThemeProvider from "@/contexts/ThemeContext";

import Layers from "@/pages/layers/Layers";

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <Layers />
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;
