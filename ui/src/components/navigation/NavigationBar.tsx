import React from "react";
import { twMerge } from "tailwind-merge";

import Burrito from "@/assets/illustrations/Burrito";
import Sombrero from "@/assets/illustrations/Sombrero";
import LayerGroupIcon from "@/assets/icons/LayerGroupIcon";
import CodeBranchIcon from "@/assets/icons/CodeBranchIcon";

export interface NavigationBarProps {
  className?: string;
  variant?: "light" | "dark";
}

export const NavigationBar: React.FC<NavigationBarProps> = ({
  className,
  variant,
}) => {
  const styles = {
    light: `bg-background-light
      fill-nuances-black`,
    dark: `bg-background-dark
      fill-nuances-50`,
  };

  variant = variant ?? "light";

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        relative
        items-center
        h-screen
        w-[72px]
        py-8
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
    </div>
  );
};

export default NavigationBar;
