import ELK from 'elkjs/lib/elk.bundled.js';
import { StateGraph, StateGraphNode, StateGraphEdge } from '@/clients/layers/types.ts';
import { Node, Edge, Position } from 'reactflow';

export type ReactFlowNode = Node<{
  id: string;
  type: string;
  name: string;
  count: number;
  provider: string;
  module: string;
  change: 'create' | 'delete' | 'update' | 'replace' | null;
  future?: unknown;
}>;

export type ReactFlowEdge = Edge;

export type ReactFlowGraph = {
  nodes: ReactFlowNode[];
  edges: ReactFlowEdge[];
};

// Estimate a good node size from label text lines
function sizeForLabel(label: string): { w: number; h: number } {
  const lines = String(label || '').split('\n');
  const maxChars = Math.max(0, ...lines.map((l) => l.length));
  const w = Math.max(100, Math.min(260, maxChars * 8 + 40));
  const h = Math.max(28, 22 + lines.length * 16);
  return { w, h };
}

// Async builder using ELK with fixed settings
export async function buildReactFlow(graph: StateGraph): Promise<ReactFlowGraph> {
  const positioned = await layoutWithElk(graph);
  const isTB = false; // hard-coded RIGHT direction
  const nodes: ReactFlowNode[] = (graph.nodes || []).map((n: StateGraphNode) => {
    const label = `${(n.type || '').toUpperCase()}\n${n.name || ''}`;
    const { w, h } = sizeForLabel(label);
    const p = positioned.get(n.id) || { x: 0, y: 0 };
    const pos = { x: p.x - w / 2, y: p.y - h / 2 };
    return {
      id: n.id,
      type: 'resource',
      position: pos,
      sourcePosition: isTB ? Position.Bottom : Position.Right,
      targetPosition: isTB ? Position.Top : Position.Left,
      data: {
        id: n.id,
        type: (n.type || '').toUpperCase(),
        name: n.name || '',
        count: n.instances_count || 0,
        provider: n.provider || '',
        module: n.module || '',
        change: null
      }
    };
  });
  const edges: ReactFlowEdge[] = (graph.edges || []).map((e: StateGraphEdge) => ({
    id: `${e.from}->${e.to}`,
    source: e.from,
    target: e.to,
    type: 'step',
    sourceHandle: isTB ? 'bottom' : 'right',
    targetHandle: isTB ? 'top' : 'left'
  }));
  return { nodes, edges };
}

async function layoutWithElk(
  graph: StateGraph
): Promise<Map<string, { x: number; y: number }>> {
  const elk = new ELK();

  const elkGraph = {
    id: 'root',
    layoutOptions: {
      'elk.algorithm': 'layered',
      'elk.direction': 'RIGHT',
      'elk.layered.nodePlacement.strategy': 'BRANDES_KOEPF',
      'elk.layered.crossingMinimization.strategy': 'LAYER_SWEEP',
      'elk.spacing.nodeNode': '70',
      'elk.spacing.edgeNode': '70',
      'elk.spacing.edgeEdge': '100',
      'elk.layered.spacing.nodeNodeBetweenLayers': '70',
      'elk.layered.spacing.edgeNodeBetweenLayers': '70',
      'elk.layered.spacing.edgeEdgeBetweenLayers': '100',
      'elk.layered.layering.strategy': 'MIN_WIDTH',
      'elk.layered.thoroughness': '10',
      'elk.edgeRouting': 'POLYLINE',
      'elk.edgeRoutingMode': 'AVOID_OVERLAP'
    },
    children: (graph.nodes || []).map((n: StateGraphNode) => {
      const label = `${(n.type || '').toUpperCase()}\n${n.name || ''}`;
      const { w, h } = sizeForLabel(label);
      return { id: n.id, width: w, height: h };
    }),
    edges: (graph.edges || [])
      .filter((e: StateGraphEdge) => e.from && e.to)
      .map((e: StateGraphEdge) => ({
        id: `${e.from}->${e.to}`,
        sources: [e.from],
        targets: [e.to]
      }))
  };

  const res = await elk.layout(elkGraph);
  const pos = new Map<string, { x: number; y: number }>();
  for (const child of res.children || []) {
    // ELK returns top-left coordinates; convert to center
    pos.set(child.id, {
      x: (child.x ?? 0) + (child.width ?? 0) / 2,
      y: (child.y ?? 0) + (child.height ?? 0) / 2
    });
  }
  return pos;
}
