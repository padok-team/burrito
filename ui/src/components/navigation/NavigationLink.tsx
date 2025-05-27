import React from 'react';
import { twMerge } from 'tailwind-merge';

import { Link, LinkProps } from 'react-router-dom';

export interface NavigationLinkProps {
  className?: string;
  to: LinkProps['to'];
  children: React.ReactNode;
  disabled?: boolean;
}

const NavigationLink: React.FC<NavigationLinkProps> = ({
  className,
  to,
  children,
  disabled
}) => {
  return !disabled ? (
    <Link to={to}>
      <div
        className={twMerge(
          `relative
          inline-block
          text-blue-500
          fill-blue-500
          hover:text-blue-400
          hover:fill-blue-400
          duration-500
          hover:duration-500
          after:content-['']
          after:absolute
          after:-bottom-[3px]
          after:left-0
          after:w-full
          after:h-[3px]
          after:bg-blue-500
          after:rounded-xs
          after:scale-x-0
          after:opacity-0
          after:origin-bottom-left
          after:transition-all
          after:duration-500
          hover:after:scale-x-100
          hover:after:opacity-100
          hover:after:rounded-xs
          hover:after:bg-blue-400
          hover:after:transition-all
          hover:after:duration-500`,
          className
        )}
      >
        {children}
      </div>
    </Link>
  ) : (
    <div
      className={twMerge(
        `text-nuances-200
        fill-nuances-200`,
        className
      )}
    >
      {children}
    </div>
  );
};

export default NavigationLink;
