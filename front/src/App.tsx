import LocaleProvider from 'providers/LocaleProvider';
import { FormattedMessage } from 'react-intl';

function App() {
  return (
    <LocaleProvider>
      <FormattedMessage id="title" />
    </LocaleProvider>
  );
}

export default App;
