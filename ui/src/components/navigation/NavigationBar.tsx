import React from "react";
import { twMerge } from "tailwind-merge";

import ProfilePicture from "@/components/misc/ProfilePicture";
import Burrito from "@/assets/illustrations/Burrito";
import LayerGroupIcon from "@/assets/icons/LayerGroupIcon";
import CodeBranchIcon from "@/assets/icons/CodeBranchIcon";

export interface NavigationBarProps {
  className?: string;
  variant?: "light" | "dark";
}

export const NavigationBar: React.FC<NavigationBarProps> = ({
  className,
  variant = "light",
}) => {
  const styles = {
    light: `bg-background-light
      fill-nuances-black`,
    dark: `bg-background-dark
      fill-nuances-50`,
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        justify-between
        items-center
        h-screen
        w-[72px]
        min-w-[72px]
        py-8
        gap-20
        ${styles[variant]}`,
        className
      )}
    >
      <div className="flex flex-col items-center gap-10">
        <Burrito height={40} width={40} />
        <div className="flex flex-col items-center gap-6">
          <LayerGroupIcon />
          <CodeBranchIcon />
        </div>
      </div>
      <ProfilePicture variant={variant} />
    </div>
  );
};

export default NavigationBar;
