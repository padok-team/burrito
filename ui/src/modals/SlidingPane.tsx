import React, { useEffect } from 'react';
import ReactDOM from 'react-dom';
import FocusLock from 'react-focus-lock';

interface SlidingPaneProps {
  isOpen: boolean;
  onClose: () => void;
  children?: React.ReactNode;
  width?: string;
  variant?: 'light' | 'dark';
}

const SlidingPane: React.FC<SlidingPaneProps> = ({
  isOpen,
  onClose,
  children,
  width = 'w-1/3',
  variant = 'light'
}) => {
  // Handle Escape key to close the pane
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  // Prevent background scrolling when pane is open
  useEffect(() => {
    if (isOpen) {
      document.body.classList.add('overflow-hidden');
    } else {
      document.body.classList.remove('overflow-hidden');
    }

    return () => {
      document.body.classList.remove('overflow-hidden');
    };
  }, [isOpen]);

  return ReactDOM.createPortal(
    <>
      {/* Background */}
      <div
        className={`fixed inset-0 flex bg-nuances-400 bg-opacity-50 z-9 duration-300 ease-in-out ${
          isOpen ? 'opacity-100 visible' : 'opacity-0 invisible'
        }`}
        onClick={onClose}
        aria-hidden={!isOpen}
      ></div>

      {/* Sliding Pane */}
      <FocusLock disabled={!isOpen}>
        <div
          className={`fixed top-0 right-0 h-screen z-10 shadow-lg transform transition-transform duration-300 ease-in-out ${
            isOpen ? 'translate-x-0' : 'translate-x-full'
          } ${width} ${variant === 'light' ? 'bg-primary-100' : 'bg-nuances-black'}`}
        >
          {/* Close Button */}
          <button
            aria-label="Close"
            className={`absolute top-4 right-8 text-2xl focus:outline-hidden ${
              variant === 'light' ? 'text-gray-600' : 'text-nuances-50'
            }`}
            onClick={onClose}
          >
            &times;
          </button>
          {/* Content */}
          <div className="p-8 pt-12 overflow-y-auto h-full">{children}</div>
        </div>
      </FocusLock>
    </>,
    document.body
  );
};

export default SlidingPane;
