import axios from 'axios';

import { Layers, Layer, StateGraph } from '@/clients/layers/types.ts';

export const fetchLayers = async () => {
  const response = await axios.get<Layers>(
    `${import.meta.env.VITE_API_BASE_URL}/layers`
  );
  return response.data;
};

export const fetchLayer = async (namespace: string, name: string) => {
  const response = await axios.get<Layer>(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}`
  );
  return response.data;
}

export const fetchStateGraph = async (namespace: string, name: string) => {
  const response = await axios.get<StateGraph>(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/stategraph`
  );
  return response.data;
}

export const syncLayer = async (namespace: string, name: string) => {
  const response = await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/sync`
  );
  return response;
};
