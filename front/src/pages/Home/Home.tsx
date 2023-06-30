import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayers } from 'client/layers/client.ts';

import { Container } from './Home.style';

interface Props {}

const Home: React.FC<Props> = () => {
  const query = useQuery({ queryKey: ['layers'], queryFn: fetchLayers });
  return (
    <BaseLayout>
      <Container>
        {query.isSuccess &&
          query.data.map((layerSummary) => <div>{layerSummary.name}</div>)}
      </Container>
    </BaseLayout>
  );
};

export default Home;
