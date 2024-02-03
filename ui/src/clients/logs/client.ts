import axios from "axios";

import { Logs } from "@/clients/logs/types.ts";

export const fetchLogs = async (runId: string, attemptId: number | null) => {
  const response = await axios.get<Logs>(
    `${import.meta.env.VITE_API_BASE_URL}/logs/${runId}/${attemptId}`
  );
  return response.data;
};
