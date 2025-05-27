import React, { useState, useRef } from 'react';
import { twMerge } from 'tailwind-merge';

import ThemeToggle from '@/components/misc/ThemeToggle';

import Sombrero from '@/assets/illustrations/Sombrero';

export interface ProfilePictureProps {
  className?: string;
  variant?: 'light' | 'dark';
}

const ProfilePicture: React.FC<ProfilePictureProps> = ({
  className,
  variant = 'light'
}) => {
  const [open, setOpen] = useState(false);
  const pictureRef = useRef<HTMLDivElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    if (pictureRef.current) {
      if (event.target === pictureRef.current) {
        setOpen(!open);
      }
    }
  };

  const handleBlur = (event: React.FocusEvent<HTMLDivElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget)) {
      setOpen(false);
    }
  };

  const styles = {
    light: `bg-primary-100
      outline-primary-500
      text-primary-600`,
    dark: `bg-nuances-black
      outline-nuances-300
      text-primary-100`
  };

  return (
    <div
      className={twMerge('relative flex items-center', className)}
      tabIndex={0}
      onBlur={handleBlur}
    >
      <Sombrero
        className={`
          absolute
          z-10
          -rotate-15
          -left-[5px]
          -top-[23px]
          pointer-events-none
        `}
        height={40}
        width={40}
      />
      {/* <img src={ProfilePicture} className="rounded-full" /> // TODO: Add profile picture */}
      <div
        className={`
          flex
          justify-center
          items-center
          h-10
          w-10
          outline
          outline-1
          rounded-full
          cursor-pointer
          text-base
          font-semibold
          tracking-[2px]
          pl-[2px]
          ${styles[variant]}
        `}
        onClick={handleClick}
        ref={pictureRef}
      >
        BR
      </div>
      {open && (
        <ThemeToggle
          className="absolute left-12 w-max z-10"
          variant={variant}
        />
      )}
    </div>
  );
};

export default ProfilePicture;
