import React from "react";

import SyncIcon from "@/assets/icons/SyncIcon";

const Running: React.FC = () => {
  return (
    <div className={`flex items-center gap-2 text-blue-500 fill-blue-500`}>
      <span className="text-sm font-semibold">Running</span>
      <SyncIcon className="animate-spin-slow" height={16} width={16} />
    </div>
  );
};

export default Running;
