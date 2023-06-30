import React from 'react';
import { FormattedMessage } from 'react-intl';

import { Container } from './Home.style';

interface Props {}

const Home: React.FC<Props> = () => {
  return (
    <Container>
      <FormattedMessage id="title" />
    </Container>
  );
};

export default Home;
