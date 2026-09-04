import axios from 'axios';

import { Plan } from '@/clients/plans/types.ts';

export const fetchPlan = async (
  namespace: string,
  layer: string,
  runId: string,
  attemptId: number
) => {
  if (attemptId === undefined || attemptId === null) {
    throw new Error('attemptId is required to fetch a plan');
  }
  const response = await axios.get<Plan>(
    `${import.meta.env.VITE_API_BASE_URL}/plans/${namespace}/${layer}/${runId}/${attemptId}`
  );
  return response.data;
};
