import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface BoxProps {
  className?: string;
  variant?: 'light' | 'dark';
  children: React.ReactNode;
}

const Box: React.FC<BoxProps> = ({
  className,
  variant = 'light',
  children
}) => {
  const styles = {
    light: `bg-nuances-white
      shadow-light`,
    dark: `bg-nuances-black
      shadow-dark`
  };

  return (
    <div className={twMerge(`flex rounded-lg ${styles[variant]}`, className)}>
      {children}
    </div>
  );
};

export default Box;
