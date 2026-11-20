interface TunnelChainNode {
  nodeId: number;
}

interface TunnelFormInput {
  name: string;
  type: number;
  inNodeId: TunnelChainNode[];
  outNodeId?: TunnelChainNode[];
  chainNodes?: TunnelChainNode[][];
  trafficRatio: number;
  forwardMode?: "agent" | "nftables";
  probeTargetHost?: string;
  probeTargetPort?: number;
}

export const FIXED_TUNNEL_STRATEGY = "fifo";

export const isLegacyMultiNodeTunnelPath = (
  nodes: TunnelChainNode[] | undefined,
) => (nodes || []).filter((node) => node.nodeId > 0).length > 1;

interface TunnelNodeInput {
  id: number;
  status: number;
  forwardMode?: "agent" | "nftables";
}

const isValidProbeIPv4 = (host: string) => {
  const parts = host.split(".");

  return (
    parts.length === 4 &&
    parts.every((part) => {
      if (!/^\d+$/.test(part)) {
        return false;
      }
      if (part.length > 1 && part.startsWith("0")) {
        return false;
      }

      const value = Number(part);

      return value >= 0 && value <= 255;
    })
  );
};

const isIPv4LikeProbeHost = (host: string) =>
  /^[0-9.]+$/.test(host) && host.includes(".");

const isSchemeLikeProbeHost = (host: string) => {
  const colonIndex = host.indexOf(":");

  if (colonIndex <= 0) {
    return false;
  }

  return /^[A-Za-z][A-Za-z0-9+.-]*$/.test(host.slice(0, colonIndex));
};

const isValidProbeDomain = (host: string) => {
  if (!host || host.length > 253) {
    return false;
  }

  return host.split(".").every((label) => {
    if (
      !label ||
      label.length > 63 ||
      label.startsWith("-") ||
      label.endsWith("-")
    ) {
      return false;
    }

    return /^[A-Za-z0-9-]+$/.test(label);
  });
};

const isValidProbeTargetHost = (host: string) => {
  if (isValidProbeIPv4(host)) {
    return true;
  }
  if (host.includes(":") || isIPv4LikeProbeHost(host)) {
    return false;
  }

  return isValidProbeDomain(host);
};

export const createTunnelFormDefaults = () => {
  return {
    name: "",
    type: 1,
    inNodeId: [],
    outNodeId: [],
    chainNodes: [],
    flow: 1,
    trafficRatio: 1.0,
    forwardMode: "agent",
    inIp: "",
    ipPreference: "v4",
    probeTargetHost: "",
    probeTargetPort: 0,
    status: 1,
  };
};

export const validateTunnelForm = (
  form: TunnelFormInput,
  nodes: TunnelNodeInput[],
  isEdit = false,
): Record<string, string> => {
  const errors: Record<string, string> = {};

  if (!form.name.trim()) {
    errors.name = "请输入隧道名称";
  } else if (form.name.length < 2 || form.name.length > 50) {
    errors.name = "隧道名称长度应在2-50个字符之间";
  }

  if (!form.inNodeId || form.inNodeId.length === 0) {
    errors.inNodeId = "请至少选择一个入口节点";
  } else if (!isEdit) {
    // Only enforce online check for new tunnels. During edit the backend
    // allows existing offline nodes (user may be removing them).
    const offlineInNodes = form.inNodeId.filter((item) => {
      const node = nodes.find((n) => n.id === item.nodeId);

      return node && node.status !== 1;
    });

    if (offlineInNodes.length > 0) {
      errors.inNodeId = "所有入口节点必须在线";
    }
  }

  if (form.trafficRatio <= 0 || form.trafficRatio > 100.0) {
    errors.trafficRatio = "流量倍率须大于0，支持小数（如 0.5）";
  }

  const rawProbeHost = form.probeTargetHost || "";
  const probeHost = rawProbeHost.trim();
  const probePortInput = form.probeTargetPort;
  const probePort = Number(probePortInput ?? 0);
  const hasProbeHostInput = rawProbeHost.length > 0;
  const hasProbePort = probePortInput != null && probePortInput !== 0;

  if (hasProbeHostInput || hasProbePort) {
    if (!probeHost) {
      errors.probeTargetHost = "请输入测试目标 Host";
    } else if (
      probeHost.includes("://") ||
      /[\s/?#]/.test(rawProbeHost) ||
      isSchemeLikeProbeHost(probeHost)
    ) {
      errors.probeTargetHost = "Host 不能包含协议、端口、空格或路径";
    } else if (!isValidProbeTargetHost(probeHost)) {
      errors.probeTargetHost = "测试目标 Host 格式无效";
    }

    if (!Number.isInteger(probePort) || probePort < 1 || probePort > 65535) {
      errors.probeTargetPort = "端口必须是 1-65535";
    }
  }

  if (form.type === 2) {
    if (form.forwardMode === "nftables") {
      errors.type = "nftables 节点仅支持端口转发";
    }

    if (!form.outNodeId || form.outNodeId.length === 0) {
      errors.outNodeId = "请至少选择一个出口节点";
    } else {
      if (!isEdit && isLegacyMultiNodeTunnelPath(form.outNodeId)) {
        errors.outNodeId = "新隧道只能选择一个固定出口节点";
      }

      if (!isEdit) {
        const offlineOutNodes = form.outNodeId.filter((item) => {
          const node = nodes.find((n) => n.id === item.nodeId);

          return node && node.status !== 1;
        });

        if (offlineOutNodes.length > 0) {
          errors.outNodeId = "所有出口节点必须在线";
        }
      }

      const inNodeIds = form.inNodeId.map((item) => item.nodeId);
      const outNodeIds = form.outNodeId.map((item) => item.nodeId);
      const overlap = inNodeIds.filter((id) => outNodeIds.includes(id));

      if (overlap.length > 0) {
        errors.outNodeId = "隧道转发模式下，入口和出口不能有相同节点";
      }
    }

    if (
      !isEdit &&
      (form.chainNodes || []).some((group) =>
        isLegacyMultiNodeTunnelPath(group),
      )
    ) {
      errors.chainNodes = "新隧道的每一跳只能选择一个固定节点";
    }
  }

  if (form.forwardMode === "nftables") {
    const nftEntryNodes = (form.inNodeId || []).filter((item) => {
      const node = nodes.find((n) => n.id === item.nodeId);

      return node?.forwardMode === "nftables";
    });

    if (nftEntryNodes.length > 0) {
      if (form.inNodeId.length !== 1) {
        errors.inNodeId = "nftables 节点仅支持单入口隧道";
      }
      if ((form.outNodeId || []).length > 0) {
        errors.outNodeId = "nftables 节点不支持出口节点配置";
      }
    }
  }

  return errors;
};

export const getTunnelTypeDisplay = (type: number) => {
  switch (type) {
    case 1:
      return { text: "端口转发", color: "primary" };
    case 2:
      return { text: "隧道转发", color: "secondary" };
    default:
      return { text: "未知", color: "default" };
  }
};

export const getTunnelFlowDisplay = (flow: number) => {
  switch (flow) {
    case 1:
      return "单向计算";
    case 2:
      return "双向计算";
    default:
      return "未知";
  }
};
