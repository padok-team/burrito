import React from 'react';
import { twMerge } from 'tailwind-merge';
import { Tooltip } from 'react-tooltip';
export interface GenericIconButtonProps {
  className?: string;
  variant?: 'light' | 'dark';
  disabled?: boolean;
  tooltip?: string;
  width?: number;
  height?: number;
  onClick?: () => void;
  Icon: React.FC<React.SVGProps<SVGSVGElement>>;
}

const GenericIconButton: React.FC<GenericIconButtonProps> = ({
  className,
  variant,
  disabled,
  tooltip,
  onClick,
  width = 40,
  height = 40,
  Icon
}) => {
  const hoverClass = !disabled
    ? variant === 'light'
      ? 'hover:bg-primary-300'
      : 'hover:bg-nuances-black'
    : '';
  return (
    <div style={{ width: `${width}px`, height: `${height}px` }}>
      <Tooltip
        opacity={1}
        id="generic-button-tooltip"
        variant={variant === 'light' ? 'dark' : 'light'}
      />
      <button
        onClick={disabled ? undefined : onClick}
        disabled={disabled}
        className={twMerge(
          `${hoverClass}
          disabled:opacity-50 
          disabled:cursor-default
          rounded-full
          cursor-pointer
          transition-colors
          duration-300`,
          className
        )}
      >
        <Icon
          data-tooltip-id="generic-button-tooltip"
          data-tooltip-content={tooltip}
          className="p-2 fill-blue-500"
          width={width}
          height={height}
        />
      </button>
    </div>
  );
};

export default GenericIconButton;
