import React, { useState } from "react";
import { twMerge } from "tailwind-merge";

import AvocadoOn from "@/assets/avocado/AvocadoOn";
import AvocadoOff from "@/assets/avocado/AvocadoOff";
import AvocadoSeed from "@/assets/avocado/AvocadoSeed";

export interface ToggleProps {
  className?: string;
  checked?: boolean;
  defaultChecked?: boolean;
  label?: string;
  labelPlacement?: "left" | "right";
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const Toggle: React.FC<ToggleProps> = ({
  className,
  checked,
  defaultChecked,
  label,
  labelPlacement = "left",
  onChange,
}) => {
  const [internalChecked, setInternalChecked] = useState(
    defaultChecked ?? false
  );

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setInternalChecked(event.target.checked);
  };

  return (
    <label className={twMerge(`flex items-center cursor-pointer`, className)}>
      {labelPlacement === "left" && label && (
        <span className="mr-2">{label}</span>
      )}
      <div className="relative rotate-180 ">
        <AvocadoOn
          className={`
            relative
            top-0
            left-0
            transition-all
            ease-in-out
            duration-500
            ${
              checked ?? internalChecked ? "opacity-100" : "opacity-0 delay-150"
            }
          `}
          height={32}
          width={48}
        />
        <AvocadoOff
          className={`
            absolute
            top-0
            left-0
            transition-all
            ease-in-out
            duration-500
            ${
              checked ?? internalChecked ? "opacity-0 delay-150" : "opacity-100"
            }
          `}
          height={32}
          width={48}
        />
        <AvocadoSeed
          className={`
            absolute
            top-0
            left-0
            transition-all
            duration-500
            ${checked ?? internalChecked ? "rotate-90" : "translate-x-[22px] "}
          `}
          height={32}
          width={32}
        />
        <input
          type="checkbox"
          className="hidden"
          checked={checked ?? internalChecked}
          onChange={onChange ?? handleChange}
        />
      </div>
      {labelPlacement === "right" && label && (
        <span className="ml-2">{label}</span>
      )}
    </label>
  );
};

export default Toggle;
