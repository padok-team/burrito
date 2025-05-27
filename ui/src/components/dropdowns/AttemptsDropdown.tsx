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
import { useQuery } from '@tanstack/react-query';

import { fetchAttempts } from '@/clients/runs/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';

import Dropdown from '@/components/core/Dropdown';
import Checkbox from '@/components/core/Checkbox';

export interface AttemptsDropdownProps {
  className?: string;
  variant?: 'light' | 'dark';
  disabled?: boolean;
  runId: string;
  namespace: string;
  layer: string;
  selectedAttempts: number[];
  setSelectedAttempts: (attempts: number[]) => void;
}

const AttemptsDropdown: React.FC<AttemptsDropdownProps> = ({
  className,
  variant = 'light',
  disabled,
  namespace,
  layer,
  runId,
  selectedAttempts,
  setSelectedAttempts
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

  const attemptsQuery = useQuery({
    queryKey: reactQueryKeys.attempts(namespace, layer, runId),
    queryFn: () => fetchAttempts(namespace, layer, runId)
  });

  const handleSelect = (attempt: number) => {
    if (selectedAttempts.includes(attempt)) {
      setSelectedAttempts(selectedAttempts.filter((a) => a !== attempt));
    } else {
      setSelectedAttempts([...selectedAttempts, attempt]);
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
        label="Attempts"
        filled={selectedAttempts.length > 0}
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
                p-2
                z-20`,
                styles[variant]
              )}
              {...getFloatingProps()}
            >
              <span className="font-semibold px-2">Attempts</span>
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
                {attemptsQuery.isLoading && <span>Loading...</span>}
                {attemptsQuery.isError && <span>An error occurred.</span>}
                {attemptsQuery.isSuccess &&
                  (attemptsQuery.data.count !== 0 ? (
                    Array.from({ length: attemptsQuery.data.count }).map(
                      (_, index) => (
                        <Checkbox
                          key={index}
                          role="option"
                          variant={variant}
                          label={`Attempt ${index + 1}`}
                          checked={selectedAttempts.includes(index)}
                          readOnly
                          tabIndex={activeIndex === index ? 0 : -1}
                          ref={(node) => {
                            listElementsRef.current[index] = node;
                            listContentRef.current[index] = `Attempt ${
                              index + 1
                            }`;
                          }}
                          {...getItemProps({
                            onClick() {
                              handleSelect(index);
                            },
                            onKeyDown(event) {
                              if (event.key === 'Enter') {
                                event.preventDefault();
                                handleSelect(index);
                              }

                              if (event.key === ' ' && !isTypingRef.current) {
                                event.preventDefault();
                                handleSelect(index);
                              }
                            }
                          })}
                        />
                      )
                    )
                  ) : (
                    <span>No attempts found.</span>
                  ))}
              </div>
            </div>
          </FloatingFocusManager>
        </FloatingPortal>
      )}
    </>
  );
};

export default AttemptsDropdown;
