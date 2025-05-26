import React from 'react';

import Button from '@/components/core/Button';

export interface SSOButtonProps {
  className?: string;
  isLoading?: boolean;
  onClick?: () => void;
}

const SSOButton: React.FC<SSOButtonProps> = ({
  className,
  isLoading,
  onClick
}) => {
  return (
    <Button
      className={className}
      variant="secondary"
      isLoading={isLoading}
      onClick={onClick}
    >
      <div className="flex items-center gap-4 justify-center">
        <span>Login with SSO</span>
      </div>
    </Button>
  );
};

export default SSOButton;
