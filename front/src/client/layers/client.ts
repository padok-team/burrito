import { LayerSummary } from 'client/layers/type.ts';
import { get } from 'client/client.ts';

export async function fetchLayerSummaries(): Promise<LayerSummary[]> {
  const response = await get<LayerSummary[]>(
    `${import.meta.env.VITE_BACKEND_URL}/layers`,
  );

  return response.data;
}
