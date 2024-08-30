import React from "react";
import { twMerge } from "tailwind-merge";

export interface SocialButtonProps {
  className?: string;
  variant?: "light" | "dark";
  Icon: React.FC<React.SVGProps<SVGSVGElement>>;
  onClick?: () => void;
}

const SocialButton: React.FC<SocialButtonProps> = ({
  className,
  variant,
  onClick,
  Icon
}) => {
  return (
    <button
    onClick={onClick}
    className={twMerge(
      `${
        variant === "light"
          ? "hover:bg-primary-300 "
          : "hover:bg-nuances-black"
      }
      rounded-full
      cursor-pointer
      transition-colors
      duration-300`,
      className
    )}
  >
    <Icon className="p-2 fill-blue-500" width={40} height={40} />
  </button>
  );
};

export default SocialButton;
