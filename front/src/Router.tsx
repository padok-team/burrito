import React from 'react';
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';

import Home from 'pages/Home';
import Layer from 'pages/Layer';

export const PATHS = {
  HOME: '/home',
  LAYER: '/layers/:id'
};
const Router: React.FC = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path={PATHS.HOME} element={<Home />} />
        <Route path={PATHS.LAYER} element={<Layer />} />
        <Route path="*" element={<Navigate replace to="/" />} />
      </Routes>
    </BrowserRouter>
  );
};

export default Router;
