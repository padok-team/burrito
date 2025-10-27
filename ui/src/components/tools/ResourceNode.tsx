import React from 'react';
import { Handle, Position } from 'reactflow';
import { twMerge } from 'tailwind-merge';

type ResourceNodeData = {
  id: string;
  type: string;
  name: string;
  count: number;
  provider: string;
  module: string;
  change: 'create' | 'delete' | 'update' | 'replace' | null;
  variant?: 'light' | 'dark';
};

type ResourceNodeProps = {
  data: ResourceNodeData;
};

const ResourceNode: React.FC<ResourceNodeProps> = ({ data }) => {
  const count = data.count || 0;
  const change = data.change || null;
  const variant = data.variant ?? 'light';

  const planColorMap: Record<'create' | 'delete' | 'update' | 'replace', string> = {
    create: '#10b981',
    delete: '#ef4444',
    update: '#f59e0b',
    replace: '#8b5cf6'
  };

  const planSymbolMap: Record<'create' | 'delete' | 'update' | 'replace', string> = {
    create: '+',
    delete: '-',
    update: '~',
    replace: '↻'
  };

  const accentColor = change ? planColorMap[change] : undefined;
  const changeSymbol = change ? planSymbolMap[change] : undefined;
  const containerClass = twMerge(
    'rounded-sm border px-3 py-2 shadow-sm relative transition-colors',
    variant === 'light'
      ? 'bg-nuances-white border-slate-300 text-nuances-black'
      : 'bg-nuances-400 border-nuances-200 text-nuances-50'
  );
  const typeClass =
    variant === 'light' ? 'text-primary-200' : 'text-nuances-200';
  const nameClass =
    variant === 'light' ? 'text-nuances-black' : 'text-nuances-50';
  const countClass = twMerge(
    'absolute -top-2 -right-2 inline-flex items-center justify-center h-5 min-w-[20px] px-1 rounded-full text-[10px] font-semibold shadow',
    variant === 'light' ? 'text-primary-100 bg-nuances-black' : 'text-nuances-black bg-nuances-50'
  );

  return (
    <div
      className={containerClass}
      style={accentColor ? { boxShadow: `inset 0 0 0 2px ${accentColor}33` } : undefined}
    >
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
      {accentColor && (
        <div
          className="absolute inset-y-0 left-0 w-1 rounded-l-sm"
          style={{ backgroundColor: accentColor }}
          aria-hidden
        />
      )}
      <div className={twMerge('text-[10px] tracking-wide uppercase', typeClass)}>
        {data.type}
      </div>
      <div className={twMerge('text-lg font-semibold flex items-center gap-2', nameClass)}>
        <span className="truncate max-w-[200px]" title={data.name}>
          {data.name}
        </span>
        {change && (
          <span
            className="inline-flex items-center justify-center h-5 min-w-[20px] px-1 rounded-full text-white text-[10px] font-semibold"
            style={{ backgroundColor: accentColor }}
            title={`Planned: ${change}`}
          >
            {changeSymbol}
          </span>
        )}
      </div>
      {count > 1 && (
        <span
          className={countClass}
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
