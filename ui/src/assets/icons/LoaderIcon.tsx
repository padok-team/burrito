import { SVGProps } from "react";

const LoaderIcon = (props: SVGProps<SVGSVGElement>) => (
  <svg height={24} width={24} viewBox="0 0 24 24" {...props}>
    <path d="M5.474 5.474c-.301-.301-.792-.303-1.068.02a10 10 0 1 0 7.111-3.482c-.425.02-.724.409-.67.83.053.423.438.719.863.704A8.459 8.459 0 1 1 5.5 6.587c.272-.327.275-.812-.026-1.113Z" />
  </svg>
);

export default LoaderIcon;
