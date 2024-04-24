import axios from "axios";

import { Attempts } from "@/clients/runs/types.ts";

export const fetchAttempts = async (namespace: string, layer: string, runId: string) => {
  const response = await axios.get<Attempts>(
    `${import.meta.env.VITE_API_BASE_URL}/run/${namespace}/${layer}/${runId}/attempts`
  );
  return response.data;
};
