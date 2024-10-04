import React from 'react';

interface ProgressBarProps {
  /** 
   * Progress value between 0 and 100 
   */
  value: number;
  label?: string;
  color?: string;
  className?: string;
}

const ProgressBar: React.FC<ProgressBarProps> = ({
  value,
  label,
  color = 'bg-blue-500',
  className = '',
}) => {
  // Ensure the value is between 0 and 100
  const normalizedValue = Math.min(Math.max(value, 0), 100);

  return (
    <div className={`w-full bg-gray-200 rounded-full h-4 ${className}`} aria-label="Progress Bar">
      <div
        className={`${color} h-4 rounded-full transition-width duration-300 ease-in-out`}
        style={{ width: `${normalizedValue}%` }}
        role="progressbar"
        aria-valuenow={normalizedValue}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        {label && (
          <span className="sr-only">
            {label}
          </span>
        )}
      </div>
    </div>
  );
};

export default ProgressBar;
