import React from "react";
import { twMerge } from "tailwind-merge";

import Sombrero from "@/assets/illustrations/Sombrero";

export interface ProfilePictureProps {
  className?: string;
  variant?: "light" | "dark";
}

const ProfilePicture: React.FC<ProfilePictureProps> = ({
  className,
  variant = "light",
}) => {
  const styles = {
    light: `bg-primary-100
      outline-primary-500
      text-primary-600`,
    dark: `bg-nuances-black
      outline-nuances-300
      text-primary-100`,
  };

  return (
    <div className={twMerge(`relative h-10 w-10`, className)}>
      <Sombrero
        className="absolute -rotate-[15deg] -left-[5px] -top-[23px]"
        height={40}
        width={40}
      />
      {/* <img src={ProfilePicture} className="rounded-full" /> // TODO: Add profile picture */}
      <div
        className={`flex
          justify-center
          items-center
          h-full
          w-full
          outline
          outline-1
          rounded-full
          text-base
          font-semibold
          tracking-[2px]
          pl-[2px]
          ${styles[variant]}`}
      >
        BR
      </div>
    </div>
  );
};

export default ProfilePicture;
