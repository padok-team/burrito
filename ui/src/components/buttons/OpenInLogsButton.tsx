import React from 'react';

import Button from '@/components/core/Button';
import ArrowResizeDiagonalIcon from '@/assets/icons/ArrowResizeDiagonalIcon';

export interface OpenInLogsButtonProps {
  className?: string;
  variant?: 'primary' | 'secondary' | 'tertiary';
  onClick?: () => void;
}

const OpenInLogsButton: React.FC<OpenInLogsButtonProps> = ({
  className,
  variant = 'primary',
  onClick
}) => {
  return (
    <Button
      rightIcon={<ArrowResizeDiagonalIcon />}
      className={className}
      variant={variant}
      onClick={onClick}
    >
      Open in logs
    </Button>
  );
};

export default OpenInLogsButton;
