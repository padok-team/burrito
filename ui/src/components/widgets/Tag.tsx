import React from 'react';

import { LayerState } from '@/clients/layers/types';

export interface TagProps {
  variant: LayerState;
}

const Tag: React.FC<TagProps> = ({ variant }) => {
  const styles: Record<LayerState, string> = {
    success: `bg-status-success-default text-nuances-black`,
    warning: `bg-status-warning-default text-nuances-black`,
    error: `bg-status-error-default text-nuances-white`,
    disabled: `bg-nuances-50 text-nuances-200`,
    deleted: `text-white`
  };

  const getContent = () => {
    switch (variant) {
      case 'success':
        return 'OK';
      case 'warning':
        return 'OutOfSync';
      case 'error':
        return 'Error';
      case 'disabled':
        return 'Disabled';
      case 'deleted':
        return 'Deleted';
    }
  };

  // Use inline style for deleted since purple isn't in the default tailwind config
  const inlineStyle = variant === 'deleted' ? { backgroundColor: '#8B5CF6' } : {};

  return (
    <div
      className={`
        flex
        px-3 py-1
        items-center
        gap-1
        rounded-full
        text-sm
        font-semibold
        leading-5
        ${styles[variant]}
      `}
      style={inlineStyle}
    >
      {getContent()}
    </div>
  );
};

export default Tag;
