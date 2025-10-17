import type { SVGProps } from 'react';

const SyncIcon = (props: SVGProps<SVGSVGElement>) => (
  <svg height={24} width={24} viewBox="0 0 24 24" {...props}>
    <path
      d="M19.91 15.51h-4.53a1 1 0 0 0 0 2h2.4A8 8 0 0 1 4 12a1 1 0 1 0-2 0 10 10 0 0 0 16.88 7.23V21a1 1 0 0 0 2 0v-4.5a1 1 0 0 0-.97-.99ZM12 2a10 10 0 0 0-6.88 2.77V3a1 1 0 0 0-2 0v4.5a1 1 0 0 0 1 1h4.5a1 1 0 0 0 0-2h-2.4A8 8 0 0 1 20 12a1 1 0 0 0 2 0A10 10 0 0 0 12 2Z"
      transform="scale(-1,1)"
      transform-origin="center"
    />
  </svg>
);

export default SyncIcon;
