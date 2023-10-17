import React from "react";

export interface NavigationButtonProps {
  variant?: "light" | "dark";
  icon: React.ReactNode;
  onClick?: () => void;
}

const NavigationButton: React.FC<NavigationButtonProps> = ({
  variant = "dark",
  icon,
  onClick,
}) => {
  const styles = {
    light: `bg-primary-400
      hover:bg-primary-500
      fill-primary-600`,

    dark: `bg-nuances-black
      hover:bg-nuances-400
      fill-nuances-white`,
  };

  return (
    <button
      className={`flex
        justify-center
        items-center
        w-8
        h-8
        flex-shrink
        rounded-lg
        ${styles[variant]}`}
      onClick={onClick}
    >
      {icon}
    </button>
  );
};

export default NavigationButton;
