import React, { useContext } from 'react';
import { twMerge } from 'tailwind-merge';

import { ThemeContext } from '@/contexts/ThemeContext';

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
  return (
    <Box
      variant={variant}
      className={twMerge('flex flex-col justify-center p-4 gap-4 bottom-0', className)}
    >
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
      <Button
        theme={theme}
        variant={'secondary'}
        onClick={async () => {
          await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
          window.location.href = '/login';
        }}
      >
        Logout
      </Button>
    </Box>
  );
};

export default SettingsToggle;
