import React from "react";

export interface NavigationButtonProps {
  variant?: "light" | "dark";
  icon: React.ReactNode;
  onClick?: () => void;
}

const NavigationButton: React.FC<NavigationButtonProps> = ({
  variant,
  icon,
  onClick,
}) => {
  const variantClasses = {
    light: `bg-primary-400
      hover:bg-primary-500
      fill-primary-600`,

    dark: `bg-nuances-black
      hover:bg-nuances-400
      fill-nuances-white`,
  };

  variant = variant ?? "dark";

  return (
    <button
      className={`flex
        justify-center
        items-center
        w-8
        h-8
        flex-shrink
        rounded-lg
        ${variantClasses[variant]}`}
      onClick={onClick}
    >
      {icon}
    </button>
  );
};

export default NavigationButton;
