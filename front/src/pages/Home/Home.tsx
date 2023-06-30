import React from 'react';
import { FormattedMessage } from 'react-intl';

import { Container } from './Home.style';
import BaseLayout from 'layouts/BaseLayout';

interface Props {}

const Home: React.FC<Props> = () => {
  return (
    <BaseLayout>
      <Container>
        <FormattedMessage id="title" />
      </Container>
    </BaseLayout>
  );
};

export default Home;
