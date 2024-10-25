import React, { useContext } from 'react';
import { Outlet } from 'react-router-dom';

import { ThemeContext } from '@/contexts/ThemeContext';

import NavigationBar from '@/components/navigation/NavigationBar';

const Layout: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  return (
    <div
      className={`
        flex
        ${theme === 'light' ? 'bg-primary-100' : 'bg-nuances-black'}
      `}
    >
      <NavigationBar variant={theme} />
      <Outlet />
    </div>
  );
};

export default Layout;
