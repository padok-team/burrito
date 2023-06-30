import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayerSummaries } from 'client/layers/client.ts';

import { Card, Container } from './Home.style';

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
            <Card key={layerSummary.id}>{layerSummary.name}</Card>
          ))}
      </Container>
    </BaseLayout>
  );
};

export default Home;
