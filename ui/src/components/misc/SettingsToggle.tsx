import React, { useContext } from 'react';
import { twMerge } from 'tailwind-merge';
import { useQuery } from '@tanstack/react-query';

import { ThemeContext } from '@/contexts/ThemeContext';
import { getUserInfo, UserInfo } from '@/clients/auth/client';

import Box from '@/components/core/Box';
import Toggle from '@/components/core/Toggle';
import Button from '@/components/core/Button';

export interface SettingsToggleProps {
  className?: string;
  variant?: 'light' | 'dark';
}

const SettingsToggle: React.FC<SettingsToggleProps> = ({
  className,
  variant = 'light'
}) => {
  const { theme, setTheme } = useContext(ThemeContext);
  // Load current user info
  const { data: user, error } = useQuery<UserInfo, Error>({
    queryKey: ['userInfo'],
    queryFn: getUserInfo,
    retry: false
  });
  return (
    <Box
      variant={variant}
      className={twMerge(
        'flex flex-col justify-center p-4 gap-4 bottom-0',
        className
      )}
    >
      {user && (
        <div
          className={`text-base font-bold text-center ${variant === 'light' ? 'text-nuances-black' : 'text-nuances-50'}`}
        >
          {user.name ?? user.email}
        </div>
      )}

      <Toggle
        className={`
          text-base
          font-normal
          ${variant === 'light' ? 'text-nuances-black' : 'text-nuances-50'}
        `}
        checked={theme === 'dark'}
        onChange={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
        label={`${variant === 'dark' ? 'Disable' : 'Enable'} Dark Mode`}
      />
      {user && !error && (
        <Button
          theme={theme}
          variant={'secondary'}
          onClick={async () => {
            await fetch('/auth/logout', {
              method: 'POST',
              credentials: 'include'
            });
            window.location.href = '/login';
          }}
        >
          Logout
        </Button>
      )}
    </Box>
  );
};

export default SettingsToggle;
