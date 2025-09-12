import axios from 'axios';

import { Layers } from '@/clients/layers/types.ts';

export const fetchLayers = async () => {
  const response = await axios.get<Layers>(
    `${import.meta.env.VITE_API_BASE_URL}/layers`
  );
  return response.data;
};

export const syncLayer = async (namespace: string, name: string) => {
  const response = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/sync`
  );
  return response;
};

export const applyLayer = async (namespace: string, name: string) => {
  const response = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/apply`
  );
  return response;
};
