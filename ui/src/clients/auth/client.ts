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
 * Fetches the authentication methods enabled on the server. Both can be
 * enabled at once (e.g. OIDC with a basic-auth admin account kept as a
 * fallback).
 * Expects JSON response: { oidc: boolean, basicAuth: boolean }
 */
export interface AuthMethods {
  oidc: boolean;
  basicAuth: boolean;
}

export async function getAuthMethods(): Promise<AuthMethods> {
  const res = await fetch('/auth/type', { credentials: 'include' });
  if (!res.ok) {
    throw new Error(`Failed to fetch auth methods: ${res.status}`);
  }
  return res.json();
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
