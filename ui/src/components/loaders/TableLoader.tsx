import React from 'react';
import { twMerge } from 'tailwind-merge';

export interface TableLoaderProps {
  className?: string;
  variant?: 'light' | 'dark';
}

const TableLoader: React.FC<TableLoaderProps> = ({
  className,
  variant = 'light'
}) => {
  const styles = {
    light: `bg-[linear-gradient(270deg,#D8EBFF_0%,#ECF5FF_100%)]`,
    dark: `bg-[linear-gradient(270deg,#252525_0%,rgba(68,67,67,0.24)_100%)]`
  };

  return (
    <div
      className={twMerge(
        `h-4
        rounded-full
        animate-pulse
        ${styles[variant]}`,
        className
      )}
    ></div>
  );
};

export default TableLoader;
