import React, { useState, useRef } from 'react';
import { twMerge } from 'tailwind-merge';
import {
  useFloating,
  useClick,
  useDismiss,
  useRole,
  useListNavigation,
  useTypeahead,
  useInteractions,
  FloatingFocusManager,
  offset,
  flip,
  size,
  autoUpdate,
  FloatingPortal
} from '@floating-ui/react';

import Dropdown from '@/components/core/Dropdown';

export interface PaginationDropdownProps {
  className?: string;
  variant?: 'light' | 'dark';
  disabled?: boolean;
  selectedPagination: number;
  setSelectedPagination: (pagination: number) => void;
}

const options: Array<{ value: number; label: string }> = [
  { value: 5, label: '5' },
  { value: 10, label: '10' },
  { value: 25, label: '25' },
  { value: 50, label: '50' }
];

const PaginationDropdown: React.FC<PaginationDropdownProps> = ({
  className,
  variant = 'light',
  disabled,
  selectedPagination,
  setSelectedPagination
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const listElementsRef = useRef<Array<HTMLElement | null>>([]);
  const listContentRef = useRef<Array<string | null>>([]);
  const isTypingRef = useRef(false);

  const { refs, floatingStyles, context } = useFloating<HTMLElement>({
    placement: 'bottom-start',
    open: isOpen,
    onOpenChange: setIsOpen,
    whileElementsMounted: autoUpdate,
    middleware: [
      offset(8),
      flip(),
      size({
        apply({ availableHeight, elements }) {
          elements.floating.style.maxHeight = `${availableHeight}px`;
        },
        padding: 8
      })
    ]
  });

  const click = useClick(context, {
    enabled: !disabled,
    event: 'mousedown'
  });
  const listNavigation = useListNavigation(context, {
    enabled: !disabled,
    listRef: listElementsRef,
    activeIndex: activeIndex,
    onNavigate: setActiveIndex
  });
  const typeahead = useTypeahead(context, {
    enabled: !disabled,
    listRef: listContentRef,
    activeIndex: activeIndex,
    onMatch: setActiveIndex,
    onTypingChange(isTyping) {
      isTypingRef.current = isTyping;
    }
  });
  const dismiss = useDismiss(context);
  const role = useRole(context, { role: 'select' });

  const { getReferenceProps, getFloatingProps } = useInteractions([
    click,
    listNavigation,
    typeahead,
    dismiss,
    role
  ]);

  const styles = {
    light: `bg-nuances-white
      text-primary-600
      shadow-light`,
    dark: `bg-nuances-black
      text-nuances-300
      shadow-dark`
  };

  return (
    <>
      <Dropdown
        className={className}
        label={selectedPagination.toString()}
        filled={true}
        disabled={disabled}
        variant={variant}
        ref={refs.setReference}
        {...getReferenceProps()}
      />
      {isOpen && (
        <FloatingPortal>
          <FloatingFocusManager context={context} modal={false}>
            <div
              ref={refs.setFloating}
              style={floatingStyles}
              className={twMerge(
                `flex
                flex-col
                rounded-lg
                outline-hidden
                p-2`,
                styles[variant]
              )}
              {...getFloatingProps()}
            >
              <div className="flex flex-col gap-1 overflow-auto">
                {options.map(({ value, label }) => (
                  <button
                    key={value}
                    className="outline-hidden"
                    onClick={() => {
                      setSelectedPagination(value);
                      setIsOpen(false);
                    }}
                  >
                    {label}
                  </button>
                ))}
              </div>
            </div>
          </FloatingFocusManager>
        </FloatingPortal>
      )}
    </>
  );
};

export default PaginationDropdown;
