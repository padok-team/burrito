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

import Checkbox from "@/components/core/Checkbox";
import AngleDownIcon from "@/assets/icons/AngleDownIcon";

const options = ["OK", "OutOfSync", "Error"];

export interface NewDropdownProps {
  className?: string;
  variant?: "light" | "dark";
  label: string;
  filled?: boolean;
  disabled?: boolean;
}

const NewDropdown: React.FC<NewDropdownProps> = ({
  className,
  variant = "light",
  label,
  filled,
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

  const click = useClick(context, { event: "mousedown" });
  const dismiss = useDismiss(context);
  const role = useRole(context, { role: "listbox" });
  const listNavigation = useListNavigation(context, {
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

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions(
    [dismiss, role, listNavigation, click, typeahead]
  );

  const handleSelect = (index: number) => {
    setSelectedIndices((prev) =>
      prev.includes(index) ? prev.filter((i) => i !== index) : [...prev, index]
    );
  };

  const styles = {
    base: {
      light: `bg-primary-400
        text-primary-600
        fill-primary-600`,

      dark: `bg-nuances-400
        text-nuances-300
        fill-nuances-300`,
    },

    filled: {
      light: `text-nuances-black`,
      dark: `text-nuances-50`,
    },

    disabled: `bg-nuances-50
      text-nuances-200
      fill-nuances-200
      hover:outline-0
      focus:outline-0
      cursor-default`,

    children: {
      light: `bg-nuances-white
        shadow-light`,
      dark: `bg-nuances-black
        shadow-dark`,
    },
  };

  return (
    <>
      <div
        className={twMerge(
          `relative
          flex
          flex-row
          items-center
          justify-center
          h-8
          p-2
          gap-2
          rounded-lg
          text-base
          font-medium
          whitespace-nowrap
          cursor-pointer
          outline-primary-600
          outline-offset-0
          hover:outline
          hover:outline-1
          focus:outline
          focus:outline-2
          ${styles.base[variant]}`,
          className,
          filled && styles.filled[variant],
          disabled && styles.disabled
        )}
        tabIndex={0}
        ref={refs.setReference}
        {...getReferenceProps()}
      >
        {label}
        <AngleDownIcon className="pointer-events-none" />
      </div>
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
                py-2
                ${
                  variant === "light" ? "text-primary-600" : "text-nuances-300"
                }`,
                styles.children[variant]
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
