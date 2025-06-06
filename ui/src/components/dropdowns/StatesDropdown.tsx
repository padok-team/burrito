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
import Checkbox from '@/components/core/Checkbox';

import { LayerState } from '@/clients/layers/types';

export interface StatesDropdownProps {
  className?: string;
  variant?: 'light' | 'dark';
  disabled?: boolean;
  selectedStates: LayerState[];
  setSelectedStates: (states: LayerState[]) => void;
}

const options: Array<{ value: LayerState; label: string }> = [
  { value: 'success', label: 'OK' },
  { value: 'warning', label: 'OutOfSync' },
  { value: 'error', label: 'Error' }
];

const StatesDropdown: React.FC<StatesDropdownProps> = ({
  className,
  variant = 'light',
  disabled,
  selectedStates,
  setSelectedStates
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

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions(
    [click, listNavigation, typeahead, dismiss, role]
  );

  const handleSelect = (state: LayerState) => {
    if (selectedStates.includes(state)) {
      setSelectedStates(selectedStates.filter((s) => s !== state));
    } else {
      setSelectedStates([...selectedStates, state]);
    }
  };

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
        label="States"
        filled={selectedStates.length > 0}
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
              <span className="font-semibold px-2">States</span>
              <hr
                className={`
                  h-px
                  w-auto
                  -mx-2
                  my-2
                  ${
                    variant === 'light'
                      ? 'border-primary-600'
                      : 'border-nuances-300'
                  }
                `}
              />
              <div className="flex flex-col gap-1 px-2 py-0.5 overflow-auto">
                {options.map(({ value, label }, index) => (
                  <Checkbox
                    key={value}
                    role="option"
                    variant={variant}
                    label={label}
                    checked={selectedStates.includes(value)}
                    readOnly
                    tabIndex={activeIndex === index ? 0 : -1}
                    ref={(node) => {
                      listElementsRef.current[index] = node;
                      listContentRef.current[index] = label;
                    }}
                    {...getItemProps({
                      onClick() {
                        handleSelect(value);
                      },
                      onKeyDown(event) {
                        if (event.key === 'Enter') {
                          event.preventDefault();
                          handleSelect(value);
                        }

                        if (event.key === ' ' && !isTypingRef.current) {
                          event.preventDefault();
                          handleSelect(value);
                        }
                      }
                    })}
                  />
                ))}
              </div>
            </div>
          </FloatingFocusManager>
        </FloatingPortal>
      )}
    </>
  );
};

export default StatesDropdown;
