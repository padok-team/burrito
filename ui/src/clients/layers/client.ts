import axios from "axios";

import { Layers } from "@/clients/layers/types.ts";

export const fetchLayers = async () => {
  const response = await axios.get<Layers>(
    `${import.meta.env.VITE_API_BASE_URL}/layers`
  );
  return response.data;
};

export const syncLayer = async (namespace: string, name: string) => {
  await axios.post(
    `${import.meta.env.VITE_API_BASE_URL}/layers/${namespace}/${name}/sync`
  );
}
