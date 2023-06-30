import React, { ReactNode } from 'react';
import { IntlProvider } from 'react-intl';

import { flattenMessages } from 'providers/LocaleProvider/utils.ts';
import enMessages from 'translations/en.json';
import frMessages from 'translations/fr.json';

interface Props {
  children: ReactNode;
}

enum Locales {
  FR = 'fr',
  EN = 'en',
}

const locales = {
  [Locales.FR]: flattenMessages(frMessages),
  [Locales.EN]: flattenMessages(enMessages),
};

const LocaleProvider: React.FC<Props> = ({ children }) => {
  const locale = Locales.FR
  return (
    <IntlProvider locale={locale} messages={locales[locale]}>
      {children}
    </IntlProvider>
  );
};
export default LocaleProvider;
