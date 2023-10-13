import React from "react";

export interface TagProps {
  variant: "success" | "warning" | "error" | "disabled";
}

const Tag: React.FC<TagProps> = ({ variant }) => {
  const variantClasses = {
    success: `bg-status-success-default
    text-nuances-black`,
    warning: `bg-status-warning-default
    text-nuances-black`,
    error: `bg-status-error-default
    text-nuances-white`,
    disabled: `bg-nuances-50
    text-nuances-200`,
  };

  const getContent = () => {
    switch (variant) {
      case "success":
        return "OK";
      case "warning":
        return "OutOfSync";
      case "error":
        return "Error";
      case "disabled":
        return "Disabled";
    }
  };

  return (
    <div
      className={`flex
        px-3 py-1
        items-center
        gap-1
        rounded-full
        text-sm
        font-semibold
        leading-5
        ${variantClasses[variant]}`}
    >
      {getContent()}
    </div>
  );
};

export default Tag;
