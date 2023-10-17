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
      focus:outline-none
      fill-primary-600`,

    dark: `bg-nuances-black
      hover:bg-nuances-400
      focus:outline-none
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
      tabIndex={0}
      onClick={onClick}
    >
      {icon}
    </button>
  );
};

export default NavigationButton;
