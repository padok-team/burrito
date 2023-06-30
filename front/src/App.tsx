import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ChakraProvider } from '@chakra-ui/react';

import LocaleProvider from 'providers/LocaleProvider';
import Router from 'Router';

const queryClient = new QueryClient();

function App() {
  return (
    <ChakraProvider>
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <Router />
        </QueryClientProvider>
      </LocaleProvider>
    </ChakraProvider>
  );
}

export default App;
