import axios from "axios";

import { Layers } from "@/clients/layers/types.ts";

export const fetchLayers = async (limit: number, next?: string) => {
  const response = await axios.get<Layers>(
    `${import.meta.env.VITE_API_BASE_URL}/layers`,
    {
      params: {
        limit,
        next,
      },
    }
  );
  return response.data;
};
