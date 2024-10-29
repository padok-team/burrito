import React, { useState, useEffect, createContext } from 'react';

interface ThemeContextProps {
  theme: 'light' | 'dark';
  setTheme: (theme: 'light' | 'dark') => void;
}

export const ThemeContext = createContext<ThemeContextProps>({
  theme: 'light',
  setTheme: () => {}
});

const getInitialTheme = () => {
  const theme = localStorage.getItem('theme');
  if (theme === 'light' || theme === 'dark') {
    return theme;
  }

  const userMedia = matchMedia('(prefers-color-scheme: dark)');
  if (userMedia.matches) {
    localStorage.setItem('theme', 'dark');
    return 'dark';
  } else {
    localStorage.setItem('theme', 'light');
    return 'light';
  }
};

interface ThemeProviderProps {
  children: React.ReactNode;
}

const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [theme, setTheme] = useState<'light' | 'dark'>(getInitialTheme());

  useEffect(() => {
    localStorage.setItem('theme', theme);
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};

export default ThemeProvider;
