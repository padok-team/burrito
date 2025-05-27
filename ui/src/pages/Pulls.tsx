import React, { useContext } from 'react';

import { ThemeContext } from '@/contexts/ThemeContext';

import TrafficCone from '@/components/temp/TrafficCone';

const Pulls: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  return (
    <div className="flex grow items-center justify-center p-8">
      <div className="flex items-center gap-8">
        <TrafficCone height={80} width={80} />
        <span
          className={`
            text-5xl
            font-bold
            ${theme === 'light' ? 'text-nuances-black' : 'text-nuances-50'}
          `}
        >
          Pull requests page
        </span>
        <TrafficCone height={80} width={80} />
      </div>
    </div>
  );
};

export default Pulls;
