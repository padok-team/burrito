import React, { useState } from 'react';
import {
  useFloating,
  useDismiss,
  useRole,
  useClick,
  useInteractions,
  FloatingFocusManager,
  FloatingOverlay,
  FloatingPortal
} from '@floating-ui/react';
import { useNavigate } from 'react-router-dom';

import LogsButton from '@/components/buttons/LogsButton';
import LogsTerminal from '@/components/tools/LogsTerminal';
import OpenInLogsButton from '@/components/buttons/OpenInLogsButton';

import { Layer } from '@/clients/layers/types';

export interface ModalLogsTerminalProps {
  variant?: 'light' | 'dark';
  layer: Layer;
}

const ModalLogsTerminal: React.FC<ModalLogsTerminalProps> = ({
  variant = 'light',
  layer
}) => {
  const [isOpen, setIsOpen] = useState(false);

  const { refs, context } = useFloating({
    open: isOpen,
    onOpenChange: setIsOpen
  });

  const click = useClick(context);
  const role = useRole(context);
  const dismiss = useDismiss(context, { outsidePressEvent: 'mousedown' });

  const { getReferenceProps, getFloatingProps } = useInteractions([
    click,
    role,
    dismiss
  ]);

  const navigate = useNavigate();

  const handleOpenInLogs = () => {
    navigate(`/logs/${layer.namespace}/${layer.name}/${layer.lastRun.id}`);
  };

  return (
    <>
      <LogsButton
        variant={variant}
        ref={refs.setReference}
        {...getReferenceProps()}
      />
      <FloatingPortal>
        {isOpen && (
          <FloatingOverlay
            className="grid place-items-center z-10 bg-overlay"
            lockScroll
          >
            <FloatingFocusManager context={context}>
              <div
                className="relative"
                ref={refs.setFloating}
                {...getFloatingProps()}
              >
                <LogsTerminal
                  layer={layer}
                  run={layer.lastRun.id}
                  className="h-[70vh] w-[60vw]"
                  variant={variant}
                />
                <OpenInLogsButton
                  className="absolute -top-14 right-0"
                  variant={variant === 'light' ? 'primary' : 'secondary'}
                  onClick={handleOpenInLogs}
                />
              </div>
            </FloatingFocusManager>
          </FloatingOverlay>
        )}
      </FloatingPortal>
    </>
  );
};

export default ModalLogsTerminal;
