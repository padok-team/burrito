import React from 'react';
import { Handle, Position } from 'reactflow';

type ResourceNodeData = {
  id: string;
  type: string;
  name: string;
  count: number;
  provider: string;
  module: string;
};

type ResourceNodeProps = {
  data: ResourceNodeData;
};

const ResourceNode: React.FC<ResourceNodeProps> = ({ data }) => {
  const count = data.count || 0;

  return (
    <div className="rounded-sm border border-slate-300 bg-white px-3 py-2 shadow-sm relative">
      {/* Provide handles on all sides with stable ids for edge anchoring */}
      <Handle
        id="left"
        type="target"
        position={Position.Left}
        isConnectable={false}
        style={{ opacity: 0 }}
      />
      <Handle
        id="right"
        type="source"
        position={Position.Right}
        isConnectable={false}
        style={{ opacity: 0 }}
      />
      <Handle
        id="top"
        type="target"
        position={Position.Top}
        isConnectable={false}
        style={{ opacity: 0 }}
      />
      <Handle
        id="bottom"
        type="source"
        position={Position.Bottom}
        isConnectable={false}
        style={{ opacity: 0 }}
      />
      <div className="text-[10px] tracking-wide text-primary-200">{data.type}</div>
      <div className="text-sm text-nuances-black text-lg font-semibold flex items-center gap-2">
        <span className="truncate max-w-[200px]" title={data.name}>
          {data.name}
        </span>
      </div>
      {count > 1 && (
        <span
          className="absolute -top-2 -right-2 inline-flex items-center justify-center h-5 min-w-[20px] px-1 rounded-full text-primary-100 bg-nuances-black text-[10px] font-semibold shadow"
          title={`${count} instances`}
          aria-label={`${count} instances`}
        >
          {count}
        </span>
      )}
    </div>
  );
};

export default ResourceNode;
