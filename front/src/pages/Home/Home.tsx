import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayerSummaries } from 'client/layers/client.ts';

import { Card, Container, Name, Detail } from './Home.style';
import { generatePath } from 'react-router-dom';
import { PATHS } from 'Router.tsx';

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
            <Card
              key={layerSummary.id}
              to={generatePath(PATHS.LAYER, { id: layerSummary.id })}
            >
              <Name>{layerSummary.name}</Name>
              <Detail>
                <span>URL: </span>
                <span>{layerSummary.repoURL}</span>
              </Detail>
              <Detail>
                <span>Path: </span>
                <span>{layerSummary.path}</span>
              </Detail>
              <Detail>
                <span>Branche: </span>
                <span>{layerSummary.branch}</span>
              </Detail>
            </Card>
          ))}
      </Container>
    </BaseLayout>
  );
};

export default Home;
