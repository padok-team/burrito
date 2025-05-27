import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface RunCardLoaderProps {
  className?: string;
  variant?: 'light' | 'dark';
}

const RunCardLoader: React.FC<RunCardLoaderProps> = ({
  className,
  variant = 'light'
}) => {
  const style = {
    light: `bg-[linear-gradient(270deg,#D8EBFF_0%,#ECF5FF_100%)]
      shadow-light`,
    dark: `bg-[linear-gradient(270deg,#252525_0%,rgba(68,67,67,0.24)_100%)]
      shadow-dark`
  };

  return (
    <div
      className={twMerge(
        `h-[132px]
        rounded-2xl
        animate-pulse
        ${style[variant]}`,
        className
      )}
    ></div>
  );
};

export default RunCardLoader;
