import LocaleProvider from 'providers/LocaleProvider';

import Router from 'Router';

function App() {
  return (
    <LocaleProvider>
      <Router />
    </LocaleProvider>
  );
}

export default App;
