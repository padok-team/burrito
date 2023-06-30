import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { generatePath } from 'react-router-dom';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayerSummaries } from 'client/layers/client.ts';
import { LayerStatus } from 'client/layers/type.ts';
import { PATHS } from 'Router.tsx';

import {
  Card,
  CheckboxIcon,
  Container,
  Detail,
  Name,
  StatusContainer,
  Status,
  CrossMarkIcon,
  HourglassIcon,
} from './Home.style';

const Home: React.FC = () => {
  const query = useQuery({
    queryKey: ['layers'],
    queryFn: fetchLayerSummaries,
  });

  const getStatus = (status: LayerStatus) => {
    switch (status) {
      case LayerStatus.PlanNeeded:
        return (
          <Status>
            <CheckboxIcon />
            <span>{LayerStatus.PlanNeeded}</span>
          </Status>
        );
      case LayerStatus.Idle:
        return (
          <Status>
            <CheckboxIcon />
            <span>{LayerStatus.Idle}</span>
          </Status>
        );
      case LayerStatus.ApplyNeeded:
        return (
          <Status>
            <HourglassIcon />
            <span>{LayerStatus.ApplyNeeded}</span>
          </Status>
        );
      case LayerStatus.FailureGracePeriod:
        return (
          <Status>
            <CrossMarkIcon />
            <span>{LayerStatus.FailureGracePeriod}</span>
          </Status>
        );
    }
  };

  return (
    <BaseLayout>
      <Container>
        {query.isSuccess &&
          query.data.map((layerSummary) => (
            <Card
              key={layerSummary.id}
              to={generatePath(PATHS.LAYER, {
                name: layerSummary.name,
                namespace: layerSummary.namespace,
              })}
            >
              <div>
                <Name>{layerSummary.name}</Name>
                <Detail>
                  <span>URL: </span>
                  <span>{layerSummary.repoUrl}</span>
                </Detail>
                <Detail>
                  <span>Path: </span>
                  <span>{layerSummary.path}</span>
                </Detail>
              </div>
              <StatusContainer>
                {getStatus(layerSummary.status)}
              </StatusContainer>
            </Card>
          ))}
      </Container>
    </BaseLayout>
  );
};

export default Home;
