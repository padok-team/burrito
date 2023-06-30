import React from 'react';
import { useQuery } from '@tanstack/react-query';

import BaseLayout from 'layouts/BaseLayout';
import { fetchLayer, fetchMocLayer } from 'client/layers/client.ts';

import { Container } from './Layer.style';
import LayerCard from '../../components/LayerCard';
import { useParams } from 'react-router';

import { Grid } from '@chakra-ui/react';

const Layer: React.FC = () => {
  const { id } = useParams() as any;

  if (id) {
    const query = useQuery({
      queryKey: ['layers/' + id],
      queryFn: () => fetchMocLayer(id),
    });

    return (
      <BaseLayout>
        <Container>
          <Grid
            minH="100vh"
            templateRows="repeat(2, 1fr)"
            templateColumns="repeat(2, 1fr)"
            gap={4}
            p="1%"
            className="main-grid"
          >
            {query.isSuccess && (
              <LayerCard
                id={query.data.address}
                type={query.data.type}
              ></LayerCard>
            )}
          </Grid>
        </Container>
      </BaseLayout>
    );
  } else {
    return (
      <BaseLayout>
        <Container>
          <div> Layer not found</div>
        </Container>
      </BaseLayout>
    );
  }
};

export default Layer;
