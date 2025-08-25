import axios from 'axios';

import { Layers, LayerStatusCounts } from '@/clients/layers/types.ts';

export const fetchLayers = async () => {
  const response = await axios.get<Layers>(
    `${import.meta.env.VITE_API_BASE_URL}/layers`
  );
  return response.data;
};

export const fetchLayersStatus = async () => {
  const response = await axios.get<LayerStatusCounts>(
    `${import.meta.env.VITE_API_BASE_URL}/layers/status`
  );
  return response.data;
};

export const syncLayer = async (namespace: string, name: string) => {
  const response = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/sync`
  );
  return response;
};
