import { StateGraph, StateGraphEdge, StateGraphNode, StateGraphResourceInstance } from '@/clients/layers/types';

export type PlanAction = 'create' | 'delete' | 'update' | 'replace';

export type PlanChange = {
  action: PlanAction;
  before: unknown;
  after: unknown;
  addr: string;
  base: string;
  index: string | null;
};

export type AggregatedPlanChange = {
  action: PlanAction | null;
  instances: Record<string, PlanChange>;
  single?: PlanChange;
};

export type ParsedTerraformPlan = {
  byAddr: Map<string, PlanChange>;
  byBase: Map<string, AggregatedPlanChange>;
};

type RawPlanChange = {
  address?: string;
  addr?: string;
  resource_address?: string;
  change?: {
    actions?: string[];
    before?: unknown;
    after?: unknown;
  };
};

type RawTerraformPlan = {
  resource_changes?: RawPlanChange[];
};

const PLAN_ACTION_PRIORITY: Record<PlanAction, number> = {
  replace: 4,
  delete: 3,
  update: 2,
  create: 1
};

function normalizeAction(actions: string[] | undefined): PlanAction | null {
  if (!Array.isArray(actions) || actions.length === 0) {
    return null;
  }
  if (actions.includes('create') && actions.includes('delete')) {
    return 'replace';
  }
  if (actions.includes('create')) {
    return 'create';
  }
  if (actions.includes('delete')) {
    return 'delete';
  }
  if (actions.includes('update')) {
    return 'update';
  }
  return null;
}

function mergeAction(existing: PlanAction | null, next: PlanAction | null): PlanAction | null {
  if (!existing) return next;
  if (!next) return existing;
  return PLAN_ACTION_PRIORITY[existing] >= PLAN_ACTION_PRIORITY[next] ? existing : next;
}

function baseAddr(addr: string): string {
  const idx = addr.lastIndexOf('[');
  if (idx > -1 && addr.endsWith(']')) {
    return addr.slice(0, idx);
  }
  return addr;
}

function extractIndex(addr: string): string | null {
  const match = addr.match(/\[(.*)\]$/);
  return match ? match[0] : null;
}

function toPlanChange(rawAddr: string, action: PlanAction, change: NonNullable<RawPlanChange['change']>): PlanChange {
  const base = baseAddr(rawAddr);
  const index = extractIndex(rawAddr);
  return {
    action,
    before: change.before ?? null,
    after: change.after ?? null,
    addr: rawAddr,
    base,
    index
  };
}

export function parseTerraformPlan(plan: unknown): ParsedTerraformPlan {
  const byAddr = new Map<string, PlanChange>();
  const byBase = new Map<string, AggregatedPlanChange>();

  const resourceChanges = (plan as RawTerraformPlan)?.resource_changes;
  if (!Array.isArray(resourceChanges)) {
    return { byAddr, byBase };
  }

  for (const rc of resourceChanges) {
    const addr = rc.address || rc.addr || rc.resource_address;
    if (!addr) {
      continue;
    }
    const action = normalizeAction(rc.change?.actions);
    if (!action || !rc.change) {
      continue;
    }

    const change = toPlanChange(addr, action, rc.change);
    byAddr.set(addr, change);

    const aggregate = byBase.get(change.base) ?? { action: null, instances: {} };
    aggregate.action = mergeAction(aggregate.action, change.action);
    if (change.index) {
      aggregate.instances[change.index] = change;
    } else {
      aggregate.single = change;
    }
    byBase.set(change.base, aggregate);
  }

  return { byAddr, byBase };
}

type ParsedAddressInfo = {
  module?: string;
  type: string;
  name: string;
  mode: 'managed' | 'data';
};

function parseBaseAddress(addr: string): ParsedAddressInfo {
  const segments = addr.split('.');
  const moduleParts: string[] = [];
  let index = 0;

  while (index + 1 < segments.length && segments[index] === 'module') {
    moduleParts.push(`module.${segments[index + 1]}`);
    index += 2;
  }

  let mode: 'managed' | 'data' = 'managed';
  if (segments[index] === 'data') {
    mode = 'data';
    index += 1;
  }

  const type = segments[index] ?? '';
  const name = segments[index + 1] ?? '';

  return {
    module: moduleParts.length > 0 ? moduleParts.join('.') : undefined,
    type,
    name,
    mode
  };
}

function planChangesForAggregate(aggregate: AggregatedPlanChange): PlanChange[] {
  const out: PlanChange[] = [];
  if (aggregate.single) {
    out.push(aggregate.single);
  }
  for (const change of Object.values(aggregate.instances)) {
    out.push(change);
  }
  return out;
}

function planInstances(changes: PlanChange[]): StateGraphResourceInstance[] | undefined {
  if (changes.length === 0) {
    return undefined;
  }
  return changes.map((change) => {
    const rawAttrs = (change.after ?? change.before) ?? undefined;
    // Ensure the attributes conform to Record<string, unknown> | undefined
    const attributes = (rawAttrs && typeof rawAttrs === 'object')
      ? (rawAttrs as Record<string, unknown>)
      : undefined;
    return {
      addr: change.addr,
      attributes
    };
  });
}

function cloneInstance(instance: StateGraphResourceInstance): StateGraphResourceInstance {
  return {
    addr: instance.addr,
    dependencies: instance.dependencies ? [...instance.dependencies] : undefined,
    attributes: instance.attributes ? { ...instance.attributes } : undefined,
    created_at: instance.created_at
  };
}

function cloneNode(node: StateGraphNode): StateGraphNode {
  return {
    ...node,
    instances: node.instances ? node.instances.map(cloneInstance) : undefined
  };
}

function cloneEdge(edge: StateGraphEdge): StateGraphEdge {
  return { ...edge };
}

export function augmentStateGraphWithPlan(
  baseGraph: StateGraph | null | undefined,
  plan: ParsedTerraformPlan | null | undefined
): StateGraph {
  const mergedNodes = new Map<string, StateGraphNode>();

  if (baseGraph?.nodes) {
    for (const node of baseGraph.nodes) {
      mergedNodes.set(node.id, cloneNode(node));
    }
  }

  if (plan) {
    for (const [base, aggregate] of plan.byBase.entries()) {
      const changes = planChangesForAggregate(aggregate);
      if (changes.length === 0) {
        continue;
      }
      const planInst = planInstances(changes);

      const existing = mergedNodes.get(base);
      if (existing) {
        if ((!existing.instances || existing.instances.length === 0) && planInst) {
          existing.instances = planInst.map(cloneInstance);
        }
        if (existing.instances_count === 0) {
          existing.instances_count = planInst?.length ?? changes.length;
        }
        continue;
      }

      const parsed = parseBaseAddress(base);
      const instances = planInst ? planInst.map(cloneInstance) : undefined;
      mergedNodes.set(base, {
        id: base,
        addr: base,
        mode: parsed.mode,
        type: parsed.type,
        name: parsed.name,
        module: parsed.module,
        provider: '',
        instances_count: instances?.length ?? changes.length,
        instances
      });
    }
  }

  const nodes = Array.from(mergedNodes.values()).sort((a, b) =>
    a.id < b.id ? -1 : a.id > b.id ? 1 : 0
  );
  const edges = baseGraph?.edges ? baseGraph.edges.map(cloneEdge) : [];

  return {
    nodes,
    edges
  };
}
