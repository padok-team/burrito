import { LayerSummary } from 'client/layers/type.ts';
import { get } from 'client/client.ts';

export async function fetchLayerSummaries(): Promise<LayerSummary[]> {
  const response = await get<LayerSummary[]>(
    `${import.meta.env.VITE_BACKEND_URL}/layers`,
  );

  return response.data;
}

import { Layer, Resource, LayerStatus } from 'client/layers/type.ts';

export async function fetchLayer(
  name: string,
  namespace: string,
): Promise<Layer> {
  const response = await get<Layer>(
    `${
      import.meta.env.VITE_BACKEND_URL
    }/layer?name=${name}&namespace=${namespace}`,
  );

  return response.data;
}

export async function fetchMocLayer(
  namespace: string,
  name: string,
): Promise<Layer> {
  const resource1: Resource = {
    address: 'aws_terraform_ressource.toto',
    type: 'aws_terraform_ressource',
    status: 'green',
    depends_on: ['lucas', 'Jerem', 'Thibaut'],
  };
  const resource2: Resource = {
    address: 'aws_terraform_ressource.tata',
    type: 'aws_terraform_ressource',
    status: 'blue',
    depends_on: [],
  };
  const layer: Layer = {
    id: 'id_layer_1',
    name: 'layer_name',
    namespace: 'mynamespace',
    repoUrl: 'layer_url',
    path: 'layer_path',
    branch: 'branch',
    status: LayerStatus.Idle,
    lastPlanCommit: 'plan_commit',
    lastApplyCommit: 'apply_commit',
    lastRelevantCommit: 'relevant_commit',
    resources: [resource1, resource2],
  };
  return layer;
}
