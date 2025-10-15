import React, { useEffect, useMemo } from 'react';
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  MarkerType,
  useEdgesState,
  useNodesState,
  type Edge
} from 'reactflow';
import 'reactflow/dist/style.css';
import ResourceNode from './ResourceNode';
import { ReactFlowGraph } from '@/utils/stateGraph';

const nodeTypes = { resource: ResourceNode };

type ReactFlowViewProps = {
  rf: ReactFlowGraph;
  onNodeClick?: (id: string) => void;
  variant?: 'light' | 'dark';
};

const ReactFlowView: React.FC<ReactFlowViewProps> = ({
  rf,
  onNodeClick,
  variant = 'light'
}) => {
  const [nodes, setNodes, onNodesChange] = useNodesState(rf.nodes || []);
  const [edges, setEdges, onEdgesChange] = useEdgesState(rf.edges || []);

  // keep local RF state in sync when rf prop changes (e.g., new graph loaded)
  useEffect(() => {
    setNodes(rf.nodes || []);
  }, [rf.nodes, setNodes]);
  useEffect(() => {
    setEdges(rf.edges || []);
  }, [rf.edges, setEdges]);

  const edgesWithMarkers = useMemo(
    () =>
      (edges || []).map((e) => ({
        ...e,
        animated: true,
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: variant === 'dark' ? '#64748b' : '#94a3b8',
          width: 16,
          height: 16
        },
        style: { stroke: variant === 'dark' ? '#64748b' : '#94a3b8' }
      })) as Edge[],
    [edges, variant]
  );

  const fitViewOptions = useMemo(() => ({ padding: 0.2 }), []);

  return (
    <ReactFlow
      nodes={nodes}
      edges={edgesWithMarkers}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      nodeTypes={nodeTypes}
      nodesDraggable={true}
      fitView
      fitViewOptions={fitViewOptions}
      onNodeClick={(_, n) => onNodeClick && onNodeClick(n.id)}
      className='bg-nuances-white rounded-2xl outline-1 outline-primary-500'
    >
      <Background
        gap={32}
        color={variant === 'dark' ? '#a7a7a7ff' : '#767676ff'}
      />
      <Controls showInteractive={false} />
    </ReactFlow>
  );
};

export default ReactFlowView;
