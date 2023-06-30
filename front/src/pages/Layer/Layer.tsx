import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayer } from 'client/layers/client.ts';

import { Container } from './Layer.style';
import ResourceCard from '../../components/ResourceCard';
import { useParams } from 'react-router';

import { Grid } from '@chakra-ui/react';

import { Name } from './Layer.style';

const Layer: React.FC = () => {
  const { namespace, name } = useParams() as any;

  if (!namespace || !name) {
    return (
      <BaseLayout>
        <Container>
          <div> Layer not found</div>
        </Container>
      </BaseLayout>
    );
  }

  const queryLayer = useQuery({
    queryKey: ['layer/namepace/' + namespace + '/name/' + name],
    queryFn: () => fetchLayer(name, namespace),
  });

  if (!queryLayer.isSuccess) {
    return null;
  }

  return (
    <BaseLayout>
      <Container>
        <div>
          <div>
            <Name>Name: </Name>
            <span>{queryLayer.data.name}</span>
          </div>
          <div>
            <Name>Namespace: </Name>
            <span>{queryLayer.data.namespace}</span>
          </div>
          <div>
            <Name>Repo URL: </Name>
            <span>{queryLayer.data.repoUrl}</span>
          </div>
          <div>
            <Name>Nombres de resources: </Name>
            <span>{queryLayer.data.resources.length}</span>
          </div>
        </div>
        <Grid
          templateRows="repeat(2, 1fr)"
          templateColumns="repeat(4, 1fr)"
          gap={4}
          className="main-grid"
        >
          {queryLayer.data.resources.map((resource) => (
            <ResourceCard
              key={resource.address}
              resource={resource}
            ></ResourceCard>
          ))}
        </Grid>
      </Container>
    </BaseLayout>
  );
};

export default Layer;
