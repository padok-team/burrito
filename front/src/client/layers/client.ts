import { LayerStatus, LayerSummary } from 'client/layers/type.ts';
import { get } from 'client/client.ts';

export async function fetchLayerSummaries(): Promise<LayerSummary[]> {
  const response = await get<LayerSummary[]>(
    `${import.meta.env.VITE_BACKEND_URL}/layers`,
  );

  return response.data;
}

export async function fetchLayerSummariesMock(): Promise<LayerSummary[]> {
  return [
    {
      id: 'id',
      name: 'development',
      repoURL: 'https://github.com/padok-team/burrito',
      path: 'internal/e2e/testdata/terraform/random-pets',
      branch: 'main',
      status: LayerStatus.PlanNeeded,
    },
  ];
}
