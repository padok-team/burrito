import React from 'react';

import { Container } from './Home.style';
import BaseLayout from 'layouts/BaseLayout';

interface Props {}

const Home: React.FC<Props> = () => {
  return (
    <BaseLayout>
      <Container></Container>
    </BaseLayout>
  );
};

export default Home;
