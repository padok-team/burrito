import axios from 'axios';

import { Repositories } from '@/clients/repositories/types.ts';

export const fetchRepositories = async () => {
  const response = await axios.get<Repositories>(
    `${import.meta.env.VITE_API_BASE_URL}/repositories`
  );
  return response.data;
};
