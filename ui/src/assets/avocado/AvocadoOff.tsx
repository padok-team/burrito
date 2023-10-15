import { SVGProps } from "react";

const AvocadoOff = (props: SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={240}
    height={160}
    fill="none"
    viewBox="0 0 240 160"
    {...props}
  >
    <path
      stroke="#000"
      strokeWidth={5}
      d="M2.633 75.83v.003a74.864 74.864 0 0 0-.119 4.208v.001c.007 13.222 3.635 26.214 10.528 37.743 6.893 11.527 16.816 21.205 28.815 28.133 11.998 6.927 25.682 10.88 39.758 11.497 14.077.616 28.095-2.123 40.731-7.968l.02-.009.019-.009c19.588-9.468 40.771-15.756 62.638-18.59l5.075-.598h.001c13.04-1.533 25.065-7.45 33.805-16.684 8.744-9.24 13.596-21.168 13.596-33.552 0-12.383-4.852-24.311-13.596-33.55-8.74-9.236-20.766-15.152-33.806-16.685l-2.066-.243-.015-.002-.014-.001c-22.221-2.344-43.791-8.448-63.647-18.01-12.299-6.082-26.033-9.177-39.939-9.007-13.91.17-27.558 3.602-39.685 9.989-12.129 6.387-22.355 15.529-29.72 26.597C7.647 50.163 3.384 62.797 2.633 75.83Z"
    />
    <path
      fill="#E9F4FF"
      stroke="#C2D1DF"
      strokeWidth={10}
      d="m184.117 123.393-.029.003-.028.004c-22.646 2.932-44.598 9.442-64.914 19.257-11.524 5.323-24.335 7.829-37.216 7.265-12.896-.565-25.403-4.187-36.337-10.499-10.933-6.312-19.914-15.095-26.128-25.487-6.21-10.389-9.46-22.055-9.465-33.897 0-1.252.036-2.512.107-3.778.672-11.675 4.492-23.029 11.135-33.013 6.646-9.988 15.912-18.292 26.97-24.116 11.06-5.825 23.54-8.97 36.283-9.126 12.743-.156 25.309 2.683 36.533 8.237l.024.012.024.011c20.637 9.939 43.036 16.278 66.096 18.713l2.037.24c11.366 1.336 21.751 6.484 29.233 14.39 7.474 7.898 11.544 17.999 11.544 28.396 0 10.398-4.07 20.499-11.544 28.396-7.482 7.907-17.867 13.055-29.233 14.391l-.002.001-5.09.6Z"
    />
    <g filter="url(#a)">
      <path
        fill="#C2D1DF"
        fillRule="evenodd"
        d="M80 115c19.33 0 35-15.67 35-35S99.33 45 80 45 45 60.67 45 80s15.67 35 35 35Z"
        clipRule="evenodd"
      />
    </g>
    <defs>
      <filter
        id="a"
        width={70}
        height={70}
        x={45}
        y={45}
        colorInterpolationFilters="sRGB"
        filterUnits="userSpaceOnUse"
      >
        <feFlood floodOpacity={0} result="BackgroundImageFix" />
        <feBlend in="SourceGraphic" in2="BackgroundImageFix" result="shape" />
        <feColorMatrix
          in="SourceAlpha"
          result="hardAlpha"
          values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
        />
        <feOffset dx={10} dy={10} />
        <feComposite in2="hardAlpha" k2={-1} k3={1} operator="arithmetic" />
        <feColorMatrix values="0 0 0 0 0.700972 0 0 0 0 0.760814 0 0 0 0 0.816667 0 0 0 1 0" />
        <feBlend in2="shape" result="effect1_innerShadow_12812_4083" />
      </filter>
    </defs>
  </svg>
);

export default AvocadoOff;
