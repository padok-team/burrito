import axios from 'axios';

export const getAuthStatus = async (): Promise<boolean> => {
  const response = await axios.get(
    `/auth/`
  );
  return response.status === 200;
};

export const basicAuth =  async (formData: { username: string; password: string }) => {
  const response = await fetch('/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: new URLSearchParams({
      username: formData.username,
      password: formData.password,
    }),
  });
  
  if (!response.ok) {
    throw new Error('Invalid credentials');
  }
  
  return null;
}
