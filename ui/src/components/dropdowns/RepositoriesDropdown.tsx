import React, { useState, useRef } from 'react';
import { twMerge } from 'tailwind-merge';
import {
  useFloating,
  useClick,
  useDismiss,
  useRole,
  useListNavigation,
  useInteractions,
  FloatingFocusManager,
  offset,
  flip,
  size,
  autoUpdate,
  FloatingPortal
} from '@floating-ui/react';
import { useQuery } from '@tanstack/react-query';

import { fetchRepositories } from '@/clients/repositories/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';

import Dropdown from '@/components/core/Dropdown';
import Input from '@/components/core/Input';
import Checkbox from '@/components/core/Checkbox';

export interface RepositoriesDropdownProps {
  className?: string;
  variant?: 'light' | 'dark';
  disabled?: boolean;
  selectedRepositories: string[];
  setSelectedRepositories: (repositories: string[]) => void;
}

const RepositoriesDropdown: React.FC<RepositoriesDropdownProps> = ({
  className,
  variant = 'light',
  disabled,
  selectedRepositories,
  setSelectedRepositories
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState<string>('');
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const listElementsRef = useRef<Array<HTMLElement | null>>([]);

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
  const dismiss = useDismiss(context);
  const role = useRole(context, { role: 'combobox' });

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions(
    [click, listNavigation, dismiss, role]
  );

  const repositoriesQuery = useQuery({
    queryKey: reactQueryKeys.repositories,
    queryFn: fetchRepositories,
    select: (data) => ({
      ...data,
      results: data.results.filter((r) =>
        r.name.toLowerCase().includes(search.toLowerCase())
      )
    })
  });

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
    setActiveIndex(null);
  };

  const handleSelect = (repository: string) => {
    if (selectedRepositories.includes(repository)) {
      setSelectedRepositories(
        selectedRepositories.filter((r) => r !== repository)
      );
    } else {
      setSelectedRepositories([...selectedRepositories, repository]);
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
        label="Repositories"
        filled={selectedRepositories.length > 0}
        disabled={disabled}
        variant={variant}
        ref={refs.setReference}
        {...getReferenceProps()}
      />
      {isOpen && (
        <FloatingPortal>
          <FloatingFocusManager
            context={context}
            modal={false}
            initialFocus={-1}
          >
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
              <span className="font-semibold px-2">Repositories</span>
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
              <Input
                variant={variant}
                placeholder="Search repositories"
                value={search}
                onChange={handleSearch}
              />
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
                {repositoriesQuery.isLoading && <span>Loading...</span>}
                {repositoriesQuery.isError && <span>An error occurred.</span>}
                {repositoriesQuery.isSuccess &&
                  (repositoriesQuery.data.results.length !== 0 ? (
                    repositoriesQuery.data.results.map((repository, index) => (
                      <Checkbox
                        key={repository.name}
                        role="option"
                        variant={variant}
                        label={repository.name}
                        checked={selectedRepositories.includes(repository.name)}
                        readOnly
                        tabIndex={activeIndex === index ? 0 : -1}
                        ref={(node) => {
                          listElementsRef.current[index] = node;
                        }}
                        {...getItemProps({
                          onClick() {
                            handleSelect(repository.name);
                          },
                          onKeyDown(event) {
                            if (event.key === 'Enter' || event.key === ' ') {
                              event.preventDefault();
                              handleSelect(repository.name);
                            }
                          }
                        })}
                      />
                    ))
                  ) : (
                    <span>No repositories found.</span>
                  ))}
              </div>
            </div>
          </FloatingFocusManager>
        </FloatingPortal>
      )}
    </>
  );
};

export default RepositoriesDropdown;
