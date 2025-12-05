import React from 'react';
import {
  FloatingFocusManager,
  FloatingOverlay,
  FloatingPortal
} from '@floating-ui/react';

export interface ConfirmationModalProps {
  isOpen: boolean;
  onConfirm: () => void;
  onCancel: () => void;
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'light' | 'dark';
  context: any;
  refs: any;
  getFloatingProps: () => any;
}

const ConfirmationModal: React.FC<ConfirmationModalProps> = ({
  isOpen,
  onConfirm,
  onCancel,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  variant = 'light',
  context,
  refs,
  getFloatingProps
}) => {
  const modalStyles = {
    background: variant === 'light' ? 'bg-nuances-white' : 'bg-nuances-400',
    text: variant === 'light' ? 'text-nuances-black' : 'text-nuances-50',
    subtext: variant === 'light' ? 'text-primary-600' : 'text-nuances-300',
    border: variant === 'light' ? 'border-nuances-200' : 'border-nuances-500'
  };

  return (
    <FloatingPortal>
      {isOpen && (
        <FloatingOverlay className="grid place-items-center z-50 bg-overlay" lockScroll>
          <FloatingFocusManager context={context}>
            <div
              className={`
                relative
                p-6
                rounded-2xl
                shadow-2xl
                max-w-md
                w-full
                mx-4
                ${modalStyles.background}
              `}
              ref={refs.setFloating}
              {...getFloatingProps()}
            >
              {/* Title */}
              <h2 className={`text-xl font-bold mb-2 ${modalStyles.text}`}>
                {title}
              </h2>

              {/* Message */}
              <p className={`text-base mb-6 ${modalStyles.subtext}`}>
                {message}
              </p>

              {/* Warning Icon */}
              <div className="flex justify-center mb-6">
                <svg
                  width="64"
                  height="64"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="text-red-500"
                >
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                  <line x1="12" y1="9" x2="12" y2="13" />
                  <line x1="12" y1="17" x2="12.01" y2="17" />
                </svg>
              </div>

              {/* Buttons */}
              <div className="flex gap-3 justify-end">
                <button
                  onClick={onCancel}
                  className={`
                    px-4
                    py-2
                    rounded-lg
                    font-semibold
                    transition-colors
                    ${
                      variant === 'light'
                        ? 'bg-nuances-100 hover:bg-nuances-200 text-nuances-black'
                        : 'bg-nuances-500 hover:bg-nuances-600 text-nuances-50'
                    }
                  `}
                >
                  {cancelText}
                </button>
                <button
                  onClick={onConfirm}
                  className={`
                    px-4
                    py-2
                    rounded-lg
                    font-semibold
                    transition-colors
                    bg-red-500
                    hover:bg-red-600
                    text-white
                  `}
                >
                  {confirmText}
                </button>
              </div>
            </div>
          </FloatingFocusManager>
        </FloatingOverlay>
      )}
    </FloatingPortal>
  );
};

export default ConfirmationModal;