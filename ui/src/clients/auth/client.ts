export const basicAuth = async (formData: {
  username: string;
  password: string;
}) => {
  const response = await fetch('/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    },
    body: new URLSearchParams({
      username: formData.username,
      password: formData.password
    })
  });

  if (!response.ok) {
    throw new Error('Invalid credentials');
  }

  return null;
};

/**
 * Fetches the supported authentication method from the server.
 * Expects JSON response: { type: 'basic' | 'oauth' }
 */
export async function getAuthType(): Promise<'basic' | 'oauth'> {
  const res = await fetch('/auth/type', { credentials: 'include' });
  if (!res.ok) {
    throw new Error(`Failed to fetch auth type: ${res.status}`);
  }
  const data = (await res.json()) as { type: string };
  return data.type?.toLowerCase() === 'oauth' ? 'oauth' : 'basic';
}

/**
 * Fetches current user info from session
 * Expects JSON response: { id: string, name: string, email: string }
 */
export interface UserInfo {
  id: string;
  name?: string;
  email?: string;
  picture?: string;
}
export async function getUserInfo(): Promise<UserInfo> {
  const res = await fetch('/auth/user', { credentials: 'include' });
  if (!res.ok) {
    throw new Error('Failed to fetch user info');
  }
  return res.json();
}
