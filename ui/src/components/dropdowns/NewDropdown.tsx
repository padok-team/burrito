import React, { useState, useRef } from "react";
import { twMerge } from "tailwind-merge";
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
  autoUpdate,
  FloatingPortal,
} from "@floating-ui/react";

import Dropdown from "@/components/core/Dropdown";
import Checkbox from "@/components/core/Checkbox";

const options = ["OK", "OutOfSync", "Error"];

export interface NewDropdownProps {
  className?: string;
  variant?: "light" | "dark";
  disabled?: boolean;
}

const NewDropdown: React.FC<NewDropdownProps> = ({
  className,
  variant = "light",
  disabled,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndices, setSelectedIndices] = useState<Array<number>>([]);
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const listElementsRef = useRef<Array<HTMLElement | null>>([]);
  const listContentRef = useRef<Array<string | null>>([]);
  const isTypingRef = useRef(false);

  const { refs, floatingStyles, context } = useFloating<HTMLElement>({
    placement: "bottom-start",
    open: isOpen,
    onOpenChange: setIsOpen,
    whileElementsMounted: autoUpdate,
    middleware: [offset(8)],
  });

  const click = useClick(context, {
    enabled: !disabled,
    event: "mousedown",
  });
  const listNavigation = useListNavigation(context, {
    enabled: !disabled,
    listRef: listElementsRef,
    activeIndex,
    onNavigate: setActiveIndex,
  });
  const typeahead = useTypeahead(context, {
    enabled: isOpen,
    listRef: listContentRef,
    activeIndex,
    onMatch: setActiveIndex,
    onTypingChange: (typing) => {
      isTypingRef.current = typing;
    },
  });
  const dismiss = useDismiss(context);
  const role = useRole(context, { role: "listbox" });

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions(
    [click, listNavigation, typeahead, dismiss, role]
  );

  const handleSelect = (index: number) => {
    setSelectedIndices((prev) =>
      prev.includes(index) ? prev.filter((i) => i !== index) : [...prev, index]
    );
  };

  const styles = {
    light: `bg-nuances-white
      text-primary-600
      shadow-light`,
    dark: `bg-nuances-black
      text-nuances-300
      shadow-dark`,
  };

  return (
    <>
      <Dropdown
        className={className}
        label="State"
        filled={selectedIndices.length > 0}
        disabled={disabled}
        variant={variant}
        forwardRef={refs.setReference}
        {...getReferenceProps()}
      />
      {isOpen && (
        <FloatingPortal>
          <FloatingFocusManager context={context} modal={false}>
            <div
              ref={refs.setFloating}
              style={floatingStyles}
              className={twMerge(
                `overflow-y-auto
                rounded-lg
                outline-none
                px-4
                py-2`,
                styles[variant]
              )}
              {...getFloatingProps()}
            >
              <span className="font-semibold">State</span>
              <hr
                className={`
                  h-[1px]
                  w-auto
                  -mx-4
                  my-2
                  ${
                    variant === "light"
                      ? "border-primary-600"
                      : "border-nuances-300"
                  }
                `}
              />
              <div className="flex flex-col gap-1">
                {options.map((value, i) => (
                  <Checkbox
                    key={value}
                    role="option"
                    variant={variant}
                    label={value}
                    checked={selectedIndices.includes(i)}
                    readOnly
                    tabIndex={activeIndex === i ? 0 : -1}
                    forwardedRef={(node) => {
                      listElementsRef.current[i] = node;
                      listContentRef.current[i] = value;
                    }}
                    {...getItemProps({
                      onClick() {
                        handleSelect(i);
                      },
                      onKeyDown(event) {
                        if (event.key === "Enter") {
                          event.preventDefault();
                          handleSelect(i);
                        }

                        if (event.key === " " && !isTypingRef.current) {
                          event.preventDefault();
                          handleSelect(i);
                        }
                      },
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

export default NewDropdown;
