import React, { useContext, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';

import { ThemeContext } from '@/contexts/ThemeContext';
import { basicAuth, getAuthStatus } from '@/clients/auth/client';

import Input from '@/components/core/Input';
import Button from '@/components/core/Button';

import Burrito from '@/assets/illustrations/Burrito';
import EyeSlashIcon from '@/assets/icons/EyeSlashIcon';
import CoverLight from '@/assets/covers/cover-light.png';
import CoverDark from '@/assets/covers/cover-dark.png';
import SSOButton from '@/components/buttons/SSOButton';

const Login: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const navigate = useNavigate();
  
  // Form state
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  // Check if user is already authenticated
  const { data: isAuthenticated, isSuccess } = useQuery({
    queryKey: ['auth'],
    queryFn: getAuthStatus,
    retry: false, 
    refetchOnWindowFocus: false
  });

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: basicAuth,
    onSuccess: () => {
      navigate('/layers', { replace: true });
    },
    onError: (error: Error) => {
      setError(error.message);
    },
  });

  // Redirect to /layers if already authenticated
  if (isSuccess && isAuthenticated) {
    navigate('/layers', { replace: true });
    return null;
  }

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    loginMutation.mutate({ username, password });
  };

  return (
    <div className="flex h-screen">
      <div
        className={`
          flex
          p-16
          w-[500px]
          min-w-[500px]
          overflow-auto
          ${theme === 'light' ? 'bg-nuances-white' : 'bg-nuances-black'}
        `}
      >
        <div className="flex flex-col items-center justify-center gap-10 w-[300px] m-auto">
          <div className="flex flex-col items-start gap-2 w-full">
            <Burrito height={64} width={64} />
            <span
              className={`
                text-3xl
                font-extrabold
                ${
                  theme === 'light'
                    ? 'text-nuances-black'
                    : 'text-nuances-white'
                }
              `}
            >
              Welcome to Burrito
            </span>
          </div>
          <div className="flex flex-col items-center justify-center gap-8 w-full">
            <form onSubmit={handleLogin} className="flex flex-col items-center justify-center gap-8 w-full">
              <Input
                variant={theme}
                placeholder="Your username"
                label="Username"
                type="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
              <Input
                variant={theme}
                placeholder="Your password"
                label="Password"
                rightIcon={<EyeSlashIcon />}
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              {error && (
                <div className={`text-sm ${theme === 'light' ? 'text-red-600' : 'text-red-400'}`}>
                  {error}
                </div>
              )}
              <Button
                className="w-full"
                variant={theme === 'light' ? 'primary' : 'secondary'}
                type="submit"
                disabled={loginMutation.isPending}
              >
                {loginMutation.isPending ? 'Logging in...' : 'Login'}
              </Button>
              <div
                className={`
                  flex
                  flex-row
                  items-center
                  w-full
                  gap-4
                  ${
                    theme === 'light'
                      ? 'text-nuances-black border-nuances-black'
                      : 'text-nuances-white border-nuances-white'
                  }
                `}
              >
                <hr className="w-full" />
                <span>OR</span>
                <hr className="w-full" />
              </div>
              <div className="flex flex-col items-center justify-center gap-4 w-full">
                <SSOButton
                className="w-full"
                onClick={() => (document.location.href = '/auth/login')}
                />
              </div>
            </form>
          </div>
          <div
            className={`
              flex
              flex-row
              items-center
              justify-center
              gap-1
              w-full
              p-6
              rounded-lg
              ${
                theme === 'light'
                  ? 'bg-primary-400 text-nuances-black'
                  : 'bg-nuances-400 text-nuances-50'
              }
            `}
          >
            <span className="text-base font-normal">
              Don't have an account ?
            </span>
            <span className="text-base font-semibold">Sign up</span>
          </div>
        </div>
      </div>
      <div
        className={`
          relative
          flex
          flex-col
          overflow-hidden
          w-[calc(100%-500px)]
          pt-20
          px-16
          gap-6
          ${
            theme === 'light'
              ? 'bg-background-login-light'
              : 'bg-background-login-dark'
          }
        `}
      >
        <div
          className={`
            flex
            flex-col
            ${theme === 'light' ? 'text-nuances-black' : 'text-nuances-white'}
          `}
        >
          <span className="text-5xl font-extrabold">Burrito is a TACoS</span>
          <span className="text-base font-medium">
            Monitor the status of your layers and their impacts on your project.
          </span>
        </div>
        <img
          className={`
            absolute
            right-0
            bottom-0
            -rotate-12
            translate-y-20
            translate-x-28
            rounded-lg
            ${theme === 'light' ? 'shadow-light' : 'shadow-dark'}
          `}
          src={theme === 'light' ? CoverLight : CoverDark}
        />
      </div>
    </div>
  );
};

export default Login;
