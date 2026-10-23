import axios from 'axios';

import { PullRequests } from '@/clients/pulls/types.ts';

export const fetchPullRequests = async () => {
  const response = await axios.get<PullRequests>(
    `${import.meta.env.VITE_API_BASE_URL}/pulls`
  );
  return response.data;
};
