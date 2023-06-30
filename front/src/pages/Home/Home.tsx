import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayerSummaries } from 'client/layers/client.ts';

import { Container } from './Home.style';

const Home: React.FC = () => {
  const query = useQuery({
    queryKey: ['layers'],
    queryFn: fetchLayerSummaries,
  });
  return (
    <BaseLayout>
      <Container>
        {query.isSuccess &&
          query.data.map((layerSummary) => (
            <div key={layerSummary.id}>{layerSummary.name}</div>
          ))}
      </Container>
    </BaseLayout>
  );
};

export default Home;
