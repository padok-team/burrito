import axios from 'axios';

export const getAuthStatus = async (): Promise<boolean> => {
  const response = await axios.get(
    `/auth/`
  );
  return response.status === 200;
};
