import React from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import Home from 'pages/Home';

export const PATHS = {
  HOME: '/home',
};
const Router: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path={PATHS.HOME} element={<Home />} />
        <Route path="*" element={<Navigate replace to="/" />} />
      </Routes>
    </BrowserRouter>
  );
};

export default Router;