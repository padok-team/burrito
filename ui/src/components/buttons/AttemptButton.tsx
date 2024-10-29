import React from 'react';
import { twMerge } from 'tailwind-merge';

import TimesIcon from '@/assets/icons/TimesIcon';

export interface AttemptButtonProps {
  className?: string;
  variant?: 'light' | 'dark';
  attempt: number;
  isActive?: boolean;
  onClick?: () => void;
  onClose?: () => void;
}

const AttemptButton: React.FC<AttemptButtonProps> = ({
  className,
  variant = 'light',
  attempt,
  isActive,
  onClick,
  onClose
}) => {
  const styles = {
    base: {
      light: `bg-primary-300
        text-nuances-black
        fill-primary-600`,

      dark: `bg-nuances-300
        text-nuances-400
        fill-nuances-400`
    },

    isActive: {
      light: `bg-primary-500
        text-nuances-black
        fill-primary-600`,

      dark: `bg-nuances-black
        text-nuances-white
        fill-nuances-50`
    }
  };

  const handleClose = (e: React.MouseEvent<SVGSVGElement>) => {
    e.stopPropagation();
    onClose?.();
  };

  return (
    <button
      className={twMerge(
        `flex
        flex-row
        items-center
        gap-2
        p-4
        rounded
        cursor-pointer
        ${styles.base[variant]}`,
        isActive && styles.isActive[variant],
        className
      )}
      onClick={onClick}
    >
      <span className="whitespace-nowrap">Attempt {attempt}</span>
      <TimesIcon className="cursor-pointer" onClick={handleClose} />
    </button>
  );
};

export default AttemptButton;
