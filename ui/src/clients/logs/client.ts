import axios from "axios";

import { Logs } from "@/clients/logs/types.ts";

export const fetchLogs = async (namespace: string, layer: string, runId: string, attemptId: number | null) => {
  const response = await axios.get<Logs>(
    `${import.meta.env.VITE_API_BASE_URL}/logs/${namespace}/${layer}/${runId}/${attemptId}`
  );
  return response.data;
};
