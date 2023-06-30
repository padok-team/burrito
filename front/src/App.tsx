import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import LocaleProvider from 'providers/LocaleProvider';
import Router from 'Router';

const queryClient = new QueryClient();

function App() {
  return (
    <LocaleProvider>
      <QueryClientProvider client={queryClient}>
        <Router />
      </QueryClientProvider>
    </LocaleProvider>
  );
}

export default App;
