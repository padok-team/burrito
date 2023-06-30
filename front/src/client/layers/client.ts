import { LayerSummary } from 'client/layers/type.ts';
import { get } from 'client/client.ts';

export async function fetchLayerSummaries(): Promise<LayerSummary[]> {
  const response = await get<LayerSummary[]>(
    `${import.meta.env.VITE_BACKEND_URL}/layers`,
  );

  return response.data;
}

import { Layer } from 'client/layers/type.ts';

export async function fetchLayer(id: string): Promise<Layer> {
  const response = await get<Layer>(
    `${import.meta.env.VITE_BACKEND_URL}/layers`+id,
  );

  return response.data;
}

export async function fetchMocLayer(id: string): Promise<Layer> {
  const response : Layer = {
    address: "aws_terraform_ressource.toto",
    type: "aws_terraform_ressource",
    status: "green",
    depends_on: ["lucas","Jerem","Thibaut"]
  }

  return response;
}

