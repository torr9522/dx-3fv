import { useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";
import {
  CirclePause,
  CirclePlay,
  Activity,
  ChevronDown,
  Pencil,
  Plus,
  RefreshCw,
  Route as RouteIcon,
  RotateCw,
  Settings2,
  Trash2,
} from "lucide-react";

import {
  createIXPathProfile,
  createIXRoute,
  diagnoseIXRoute,
  deleteIXPathProfile,
  deleteIXRoute,
  getIXPathProfiles,
  getIXPortPools,
  getIXRoutes,
  getNodeList,
  pauseIXRoute,
  reconcileIXRoute,
  resumeIXRoute,
  updateIXPathProfile,
  updateIXRoute,
  type IXPathProfileApiItem,
  type IXPortPoolApiItem,
  type IXRouteApiItem,
  type IXRouteDiagnosisApiItem,
} from "@/api";
import { AnimatedPage } from "@/components/animated-page";
import {
  DiagnosisResultView,
  type DiagnosisDisplayItem,
} from "@/components/diagnosis-result-view";
import { PageLoadingState } from "@/components/page-state";
import { Button } from "@/shadcn-bridge/heroui/button";
import { Card, CardBody, CardHeader } from "@/shadcn-bridge/heroui/card";
import { Chip } from "@/shadcn-bridge/heroui/chip";
import { Input } from "@/shadcn-bridge/heroui/input";
import {
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader,
} from "@/shadcn-bridge/heroui/modal";
import { Select, SelectItem } from "@/shadcn-bridge/heroui/select";
import { Switch } from "@/shadcn-bridge/heroui/switch";
import { formatTraffic } from "@/utils/traffic";
import { useLocalStorageState } from "@/hooks/use-local-storage-state";
import {
  isPathCollapsed,
  togglePathCollapsed,
  type CollapsedPathMap,
} from "@/pages/multiport-collapse";
import { isValidIPv4OnlyLandingHost } from "@/pages/forward/landing-target";

interface NodeOption {
  id: number;
  name: string;
}

const emptyPathForm = {
  id: 0,
  name: "",
  entryNodeId: 0,
  ixNodeId: 0,
  entryPoolId: 0,
  ixPoolId: 0,
  status: 1,
};

const emptyRouteForm = {
  id: 0,
  pathProfileId: 0,
  name: "",
  entryPort: "",
  ixPort: "",
  landingHost: "",
  landingPort: "",
  protocolMode: "tcp_udp" as IXRouteApiItem["protocolMode"],
};

const protocolLabel: Record<IXRouteApiItem["protocolMode"], string> = {
  tcp: "TCP",
  udp: "UDP",
  tcp_udp: "TCP+UDP",
};

const statusColor = (state: string) => {
  if (state === "active") return "success";
  if (state === "paused") return "warning";
  if (state === "pending" || state === "cleanup_pending") return "primary";

  return "danger";
};

const validOptionalPort = (value: string) => {
  if (!value.trim()) return true;
  const port = Number(value);

  return Number.isInteger(port) && port >= 1 && port <= 65535;
};

export default function MultiportPage() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [busyRouteId, setBusyRouteId] = useState(0);
  const [nodes, setNodes] = useState<NodeOption[]>([]);
  const [pools, setPools] = useState<IXPortPoolApiItem[]>([]);
  const [paths, setPaths] = useState<IXPathProfileApiItem[]>([]);
  const [routes, setRoutes] = useState<IXRouteApiItem[]>([]);
  const [pathModalOpen, setPathModalOpen] = useState(false);
  const [routeModalOpen, setRouteModalOpen] = useState(false);
  const [deletePath, setDeletePath] = useState<IXPathProfileApiItem | null>(
    null,
  );
  const [deleteRoute, setDeleteRoute] = useState<IXRouteApiItem | null>(null);
  const [diagnosisRoute, setDiagnosisRoute] = useState<IXRouteApiItem | null>(
    null,
  );
  const [diagnosis, setDiagnosis] = useState<IXRouteDiagnosisApiItem | null>(
    null,
  );
  const [payloadEcho, setPayloadEcho] = useState(false);
  const [pathForm, setPathForm] = useState(emptyPathForm);
  const [routeForm, setRouteForm] = useState(emptyRouteForm);
  const [collapsedPaths, setCollapsedPaths] =
    useLocalStorageState<CollapsedPathMap>(
      "dx3fv-multiport-collapsed-paths",
      {},
    );

  const nodeById = useMemo(
    () => new Map(nodes.map((node) => [node.id, node])),
    [nodes],
  );
  const routesByPath = useMemo(() => {
    const grouped = new Map<number, IXRouteApiItem[]>();

    for (const route of routes) {
      const items = grouped.get(route.pathProfileId) || [];

      items.push(route);
      grouped.set(route.pathProfileId, items);
    }

    return grouped;
  }, [routes]);

  const loadAll = async () => {
    setLoading(true);
    try {
      const [pathResponse, routeResponse, poolResponse, nodeResponse] =
        await Promise.all([
          getIXPathProfiles(),
          getIXRoutes(),
          getIXPortPools(),
          getNodeList(),
        ]);

      for (const response of [
        pathResponse,
        routeResponse,
        poolResponse,
        nodeResponse,
      ]) {
        if (response.code !== 0) throw new Error(response.msg);
      }
      setPaths(pathResponse.data || []);
      setRoutes(routeResponse.data || []);
      setPools(poolResponse.data || []);
      setNodes((nodeResponse.data || []) as NodeOption[]);
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : "加载多端口业务失败",
      );
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAll();
  }, []);

  const openNewPath = () => {
    setPathForm(emptyPathForm);
    setPathModalOpen(true);
  };

  const openEditPath = (path: IXPathProfileApiItem) => {
    setPathForm({
      id: path.id,
      name: path.name,
      entryNodeId: path.entryNodeId,
      ixNodeId: path.ixNodeId,
      entryPoolId: path.entryPoolId,
      ixPoolId: path.ixPoolId,
      status: path.status,
    });
    setPathModalOpen(true);
  };

  const submitPath = async () => {
    if (!pathForm.name.trim() || !pathForm.entryNodeId || !pathForm.ixNodeId) {
      toast.error("请填写路径名称并选择前置入口与 IX 节点");

      return;
    }
    if (!pathForm.entryPoolId || !pathForm.ixPoolId) {
      toast.error("请选择前置入口与 IX 的端口资源");

      return;
    }
    if (pathForm.entryNodeId === pathForm.ixNodeId) {
      toast.error("前置入口与 IX 必须是两个独立节点身份");

      return;
    }
    setSaving(true);
    const response = pathForm.id
      ? await updateIXPathProfile(pathForm)
      : await createIXPathProfile({
          name: pathForm.name,
          entryNodeId: pathForm.entryNodeId,
          ixNodeId: pathForm.ixNodeId,
          entryPoolId: pathForm.entryPoolId,
          ixPoolId: pathForm.ixPoolId,
          status: pathForm.status,
        });

    setSaving(false);
    if (response.code !== 0) {
      toast.error(response.msg || "保存路径失败");

      return;
    }
    toast.success(pathForm.id ? "路径已更新" : "路径已创建");
    setPathModalOpen(false);
    await loadAll();
  };

  const openNewRoute = (pathId: number) => {
    setRouteForm({ ...emptyRouteForm, pathProfileId: pathId });
    setRouteModalOpen(true);
  };

  const openEditRoute = (route: IXRouteApiItem) => {
    setRouteForm({
      id: route.id,
      pathProfileId: route.pathProfileId,
      name: route.name,
      entryPort: String(route.entryPort),
      ixPort: String(route.ixPublicPort),
      landingHost: route.landingHost,
      landingPort: String(route.landingPort),
      protocolMode: route.protocolMode,
    });
    setRouteModalOpen(true);
  };

  const submitRoute = async () => {
    const landingPort = Number(routeForm.landingPort);

    if (!routeForm.pathProfileId || !routeForm.name.trim()) {
      toast.error("请选择路径并填写业务名称");

      return;
    }
    if (!routeForm.landingHost.trim()) {
      toast.error("请填写 Landing IP 或域名");

      return;
    }
    if (!isValidIPv4OnlyLandingHost(routeForm.landingHost)) {
      toast.error("Landing 仅支持 IPv4 地址或具有 A 记录的域名");

      return;
    }
    if (
      !Number.isInteger(landingPort) ||
      landingPort < 1 ||
      landingPort > 65535
    ) {
      toast.error("Landing 端口必须是 1-65535");

      return;
    }
    if (!validOptionalPort(routeForm.entryPort)) {
      toast.error("前置入口端口必须留空或填写 1-65535");

      return;
    }
    if (!validOptionalPort(routeForm.ixPort)) {
      toast.error("IX 中间端口必须留空或填写 1-65535");

      return;
    }
    setSaving(true);
    const payload = {
      pathProfileId: routeForm.pathProfileId,
      name: routeForm.name.trim(),
      entryPort: routeForm.entryPort ? Number(routeForm.entryPort) : 0,
      ixPort: routeForm.ixPort ? Number(routeForm.ixPort) : 0,
      landingHost: routeForm.landingHost.trim(),
      landingPort,
      protocolMode: routeForm.protocolMode,
    };
    const response = routeForm.id
      ? await updateIXRoute({ ...payload, id: routeForm.id })
      : await createIXRoute(payload);

    setSaving(false);
    if (response.code !== 0) {
      toast.error(
        response.msg || (routeForm.id ? "更新业务失败" : "创建业务失败"),
      );

      return;
    }
    toast.success(
      routeForm.id
        ? "业务已更新，旧端口已在新 Listener 确认后释放"
        : response.data?.observedState === "active"
          ? "业务已创建并生效"
          : "业务已保存，等待 Agent 收敛",
    );
    setRouteModalOpen(false);
    await loadAll();
  };

  const mutateRoute = async (
    route: IXRouteApiItem,
    action: "pause" | "resume" | "reconcile",
  ) => {
    setBusyRouteId(route.id);
    const response =
      action === "pause"
        ? await pauseIXRoute(route.id)
        : action === "resume"
          ? await resumeIXRoute(route.id)
          : await reconcileIXRoute(route.id);

    setBusyRouteId(0);
    if (response.code !== 0) {
      toast.error(response.msg || "操作失败");

      return;
    }
    toast.success(
      action === "pause"
        ? "业务已暂停"
        : action === "resume"
          ? "业务已恢复"
          : "已触发状态收敛",
    );
    await loadAll();
  };

  const confirmDeleteRoute = async () => {
    if (!deleteRoute) return;
    setBusyRouteId(deleteRoute.id);
    const response = await deleteIXRoute(deleteRoute.id);

    setBusyRouteId(0);
    if (response.code !== 0) {
      toast.error(response.msg || "删除业务失败");

      return;
    }
    toast.success(
      response.data?.observedState === "deleted"
        ? "业务与 Listener 已删除"
        : "删除已登记，Lease 将在 Listener 清理确认后释放",
    );
    setDeleteRoute(null);
    await loadAll();
  };

  const confirmDeletePath = async () => {
    if (!deletePath) return;
    setSaving(true);
    const response = await deleteIXPathProfile(deletePath.id);

    setSaving(false);
    if (response.code !== 0) {
      toast.error(response.msg || "删除路径失败");

      return;
    }
    toast.success("路径已删除");
    setDeletePath(null);
    await loadAll();
  };

  const openDiagnosis = (route: IXRouteApiItem) => {
    setDiagnosisRoute(route);
    setDiagnosis(null);
    setPayloadEcho(false);
  };

  const runDiagnosis = async () => {
    if (!diagnosisRoute) return;
    setBusyRouteId(diagnosisRoute.id);
    const response = await diagnoseIXRoute(diagnosisRoute.id, payloadEcho);

    setBusyRouteId(0);
    if (response.code !== 0 || !response.data) {
      toast.error(response.msg || "链路探测失败");

      return;
    }
    setDiagnosis(response.data);
  };

  const diagnosisItems = useMemo<DiagnosisDisplayItem[]>(() => {
    if (!diagnosis) return [];
    const sectionLabels = {
      gz_to_ix: "入口测试",
      ix_to_landing: "出口测试",
      end_to_end: "端到端业务",
    } as const;

    return diagnosis.legs.map((leg, index) => ({
      id: `${leg.protocol}-${leg.leg}-${index}`,
      section: sectionLabels[leg.leg],
      path: `${leg.from} → ${leg.to}`,
      protocol: leg.protocol,
      method: leg.method,
      status: leg.status,
      latencyMs: leg.status === "PASS" ? leg.latencyMs : undefined,
      packetLoss: leg.status === "PASS" ? leg.loss : undefined,
      message: leg.error,
    }));
  }, [diagnosis]);

  if (loading) return <PageLoadingState message="正在加载多端口业务" />;

  return (
    <AnimatedPage className="px-3 py-6 space-y-5 lg:px-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-xl font-semibold">
            <RouteIcon className="w-5 h-5" />
            多端口转发
          </h1>
          <p className="mt-1 text-sm text-default-500">
            前置入口到 IX 路径与独立 Landing 业务
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            isIconOnly
            title="刷新"
            variant="flat"
            onPress={() => void loadAll()}
          >
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button
            color="primary"
            startContent={<Plus className="w-4 h-4" />}
            onPress={openNewPath}
          >
            新增路径
          </Button>
        </div>
      </div>

      {paths.length === 0 ? (
        <Card>
          <CardBody className="py-14 text-center text-default-500">
            暂无前置入口到 IX 路径
          </CardBody>
        </Card>
      ) : (
        <div className="space-y-4">
          {paths.map((path) => {
            const pathRoutes = routesByPath.get(path.id) || [];
            const entryNode = nodeById.get(path.entryNodeId);
            const ixNode = nodeById.get(path.ixNodeId);
            const collapsed = isPathCollapsed(collapsedPaths, path.id);

            return (
              <Card key={path.id}>
                <CardHeader className="flex flex-col items-stretch gap-3 md:flex-row md:items-center md:justify-between">
                  <div className="flex min-w-0 items-start gap-2">
                    <Button
                      isIconOnly
                      aria-label={
                        collapsed
                          ? `展开路径 ${path.name}`
                          : `折叠路径 ${path.name}`
                      }
                      className="h-8 w-8 min-w-8"
                      size="sm"
                      title={collapsed ? "展开路径" : "折叠路径"}
                      variant="light"
                      onPress={() =>
                        setCollapsedPaths((state) =>
                          togglePathCollapsed(state, path.id),
                        )
                      }
                    >
                      <ChevronDown
                        className={`h-4 w-4 transition-transform ${collapsed ? "-rotate-90" : "rotate-0"}`}
                      />
                    </Button>
                    <div className="min-w-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <h2 className="font-semibold break-words">
                          {path.name}
                        </h2>
                        <Chip
                          color={path.status === 1 ? "success" : "default"}
                          size="sm"
                          variant="flat"
                        >
                          {path.status === 1 ? "启用" : "停用"}
                        </Chip>
                        <Chip size="sm" variant="bordered">
                          {pathRoutes.length} 条业务
                        </Chip>
                      </div>
                      <p className="mt-1 text-sm text-default-500 break-words">
                        {entryNode?.name || `Node #${path.entryNodeId}`} →{" "}
                        {ixNode?.name || `Node #${path.ixNodeId}`}
                      </p>
                    </div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button
                      isIconOnly
                      title="编辑路径"
                      variant="flat"
                      onPress={() => openEditPath(path)}
                    >
                      <Settings2 className="w-4 h-4" />
                    </Button>
                    <Button
                      isIconOnly
                      color="danger"
                      title="删除路径"
                      variant="flat"
                      onPress={() => setDeletePath(path)}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                    <Button
                      color="primary"
                      isDisabled={path.status !== 1}
                      startContent={<Plus className="w-4 h-4" />}
                      onPress={() => openNewRoute(path.id)}
                    >
                      新增业务
                    </Button>
                  </div>
                </CardHeader>
                {!collapsed && (
                  <CardBody className="pt-0">
                    {pathRoutes.length === 0 ? (
                      <div className="border-t border-divider py-8 text-center text-sm text-default-500">
                        当前路径还没有业务
                      </div>
                    ) : (
                      <div className="divide-y divide-divider border-t border-divider">
                        {pathRoutes.map((route) => (
                          <div
                            key={route.id}
                            className="grid gap-3 py-4 md:grid-cols-[minmax(150px,1.2fr)_minmax(180px,1.4fr)_minmax(170px,1fr)_auto] md:items-center"
                          >
                            <div className="min-w-0">
                              <div className="font-medium break-words">
                                {route.name}
                              </div>
                              <div className="mt-1 flex flex-wrap gap-1">
                                <Chip size="sm" variant="flat">
                                  {protocolLabel[route.protocolMode]}
                                </Chip>
                                <Chip
                                  color={statusColor(route.observedState)}
                                  size="sm"
                                  variant="flat"
                                >
                                  {route.observedState}
                                </Chip>
                              </div>
                            </div>
                            <div className="text-sm break-all">
                              <div>前置入口 :{route.entryPort}</div>
                              <div className="text-default-500">
                                IX :{route.ixPublicPort} → {route.landingHost}:
                                {route.landingPort}
                              </div>
                            </div>
                            <div className="text-sm">
                              <div>累计流量</div>
                              <div className="text-default-500">
                                {formatTraffic(route.inFlow + route.outFlow)}
                              </div>
                              {route.lastError && (
                                <div className="mt-1 text-danger break-words">
                                  {route.lastError}
                                </div>
                              )}
                            </div>
                            <div className="flex gap-2 md:justify-end">
                              <Button
                                isIconOnly
                                isDisabled={route.observedState !== "active"}
                                title="编辑业务"
                                variant="flat"
                                onPress={() => openEditRoute(route)}
                              >
                                <Pencil className="w-4 h-4" />
                              </Button>
                              <Button
                                isIconOnly
                                isDisabled={route.observedState !== "active"}
                                isLoading={busyRouteId === route.id}
                                title="链路探测"
                                variant="flat"
                                onPress={() => openDiagnosis(route)}
                              >
                                <Activity className="w-4 h-4" />
                              </Button>
                              {route.desiredState === "paused" ? (
                                <Button
                                  isIconOnly
                                  color="success"
                                  isLoading={busyRouteId === route.id}
                                  title="恢复"
                                  variant="flat"
                                  onPress={() =>
                                    void mutateRoute(route, "resume")
                                  }
                                >
                                  <CirclePlay className="w-4 h-4" />
                                </Button>
                              ) : (
                                <Button
                                  isIconOnly
                                  color="warning"
                                  isLoading={busyRouteId === route.id}
                                  title="暂停"
                                  variant="flat"
                                  onPress={() =>
                                    void mutateRoute(route, "pause")
                                  }
                                >
                                  <CirclePause className="w-4 h-4" />
                                </Button>
                              )}
                              <Button
                                isIconOnly
                                isLoading={busyRouteId === route.id}
                                title="重新收敛"
                                variant="flat"
                                onPress={() =>
                                  void mutateRoute(route, "reconcile")
                                }
                              >
                                <RotateCw className="w-4 h-4" />
                              </Button>
                              <Button
                                isIconOnly
                                color="danger"
                                title="删除"
                                variant="flat"
                                onPress={() => setDeleteRoute(route)}
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </CardBody>
                )}
              </Card>
            );
          })}
        </div>
      )}

      <Modal
        isOpen={pathModalOpen}
        scrollBehavior="inside"
        onOpenChange={setPathModalOpen}
      >
        <ModalContent>
          <ModalHeader>{pathForm.id ? "编辑路径" : "新增路径"}</ModalHeader>
          <ModalBody className="space-y-4">
            <Input
              label="路径名称"
              value={pathForm.name}
              onChange={(event) =>
                setPathForm((form) => ({ ...form, name: event.target.value }))
              }
            />
            <Select
              label="前置入口"
              selectedKeys={
                pathForm.entryNodeId ? [String(pathForm.entryNodeId)] : []
              }
              onSelectionChange={(keys) =>
                setPathForm((form) => ({
                  ...form,
                  entryNodeId: Number(Array.from(keys)[0]),
                  entryPoolId: 0,
                }))
              }
            >
              {nodes.map((node) => (
                <SelectItem key={String(node.id)} textValue={node.name}>
                  {node.name}
                </SelectItem>
              ))}
            </Select>
            <Select
              label="前置入口端口策略"
              selectedKeys={
                pathForm.entryPoolId ? [String(pathForm.entryPoolId)] : []
              }
              onSelectionChange={(keys) =>
                setPathForm((form) => ({
                  ...form,
                  entryPoolId: Number(Array.from(keys)[0]),
                }))
              }
            >
              {pools
                .filter(
                  (pool) =>
                    pool.nodeId === pathForm.entryNodeId &&
                    pool.kind === "full",
                )
                .map((pool) => (
                  <SelectItem key={String(pool.id)} textValue={pool.name}>
                    {pool.name}
                  </SelectItem>
                ))}
            </Select>
            <Select
              label="IX 节点"
              selectedKeys={
                pathForm.ixNodeId ? [String(pathForm.ixNodeId)] : []
              }
              onSelectionChange={(keys) =>
                setPathForm((form) => ({
                  ...form,
                  ixNodeId: Number(Array.from(keys)[0]),
                  ixPoolId: 0,
                }))
              }
            >
              {nodes.map((node) => (
                <SelectItem key={String(node.id)} textValue={node.name}>
                  {node.name}
                </SelectItem>
              ))}
            </Select>
            <Select
              label="IX 端口资源"
              selectedKeys={
                pathForm.ixPoolId ? [String(pathForm.ixPoolId)] : []
              }
              onSelectionChange={(keys) =>
                setPathForm((form) => ({
                  ...form,
                  ixPoolId: Number(Array.from(keys)[0]),
                }))
              }
            >
              {pools
                .filter((pool) => pool.nodeId === pathForm.ixNodeId)
                .map((pool) => (
                  <SelectItem key={String(pool.id)} textValue={pool.name}>
                    {pool.name}
                  </SelectItem>
                ))}
            </Select>
            {pathForm.id > 0 && (
              <Select
                label="路径状态"
                selectedKeys={[String(pathForm.status)]}
                onSelectionChange={(keys) =>
                  setPathForm((form) => ({
                    ...form,
                    status: Number(Array.from(keys)[0]),
                  }))
                }
              >
                <SelectItem key="1">启用</SelectItem>
                <SelectItem key="0">停用</SelectItem>
              </Select>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setPathModalOpen(false)}>
              取消
            </Button>
            <Button
              color="primary"
              isLoading={saving}
              onPress={() => void submitPath()}
            >
              保存
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal
        isOpen={routeModalOpen}
        scrollBehavior="inside"
        onOpenChange={setRouteModalOpen}
      >
        <ModalContent>
          <ModalHeader>
            {routeForm.id ? "编辑多端口业务" : "新增多端口业务"}
          </ModalHeader>
          <ModalBody className="space-y-4">
            <Select
              isDisabled={routeForm.id > 0}
              label="前置入口 → IX 路径"
              selectedKeys={
                routeForm.pathProfileId ? [String(routeForm.pathProfileId)] : []
              }
              onSelectionChange={(keys) =>
                setRouteForm((form) => ({
                  ...form,
                  pathProfileId: Number(Array.from(keys)[0]),
                }))
              }
            >
              {paths
                .filter((path) => path.status === 1)
                .map((path) => (
                  <SelectItem key={String(path.id)} textValue={path.name}>
                    {path.name}
                  </SelectItem>
                ))}
            </Select>
            <Input
              label="业务名称"
              value={routeForm.name}
              onChange={(event) =>
                setRouteForm((form) => ({ ...form, name: event.target.value }))
              }
            />
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <Input
                description={
                  routeForm.id
                    ? "保留原值不迁移；清空后重新自动分配；SSH 22 永久排除"
                    : "留空自动分配；SSH 22 永久排除"
                }
                label="前置入口端口"
                type="number"
                value={routeForm.entryPort}
                onChange={(event) =>
                  setRouteForm((form) => ({
                    ...form,
                    entryPort: event.target.value,
                  }))
                }
              />
              <Input
                description={
                  routeForm.id
                    ? "保留原值不迁移；清空后从已核验 NAT 资源重新分配"
                    : "留空从已核验 NAT 资源分配"
                }
                label="IX 中间端口"
                type="number"
                value={routeForm.ixPort}
                onChange={(event) =>
                  setRouteForm((form) => ({
                    ...form,
                    ixPort: event.target.value,
                  }))
                }
              />
              <Input
                description="IPv4 地址或具有 A 记录的域名"
                label="Landing IP / 域名"
                placeholder="192.0.2.10 或 example.com"
                value={routeForm.landingHost}
                onChange={(event) =>
                  setRouteForm((form) => ({
                    ...form,
                    landingHost: event.target.value,
                  }))
                }
              />
              <Input
                label="Landing 端口"
                type="number"
                value={routeForm.landingPort}
                onChange={(event) =>
                  setRouteForm((form) => ({
                    ...form,
                    landingPort: event.target.value,
                  }))
                }
              />
            </div>
            <Select
              label="协议"
              selectedKeys={[routeForm.protocolMode]}
              onSelectionChange={(keys) =>
                setRouteForm((form) => ({
                  ...form,
                  protocolMode: Array.from(
                    keys,
                  )[0] as IXRouteApiItem["protocolMode"],
                }))
              }
            >
              <SelectItem key="tcp">TCP</SelectItem>
              <SelectItem key="udp">UDP</SelectItem>
              <SelectItem key="tcp_udp">TCP+UDP</SelectItem>
            </Select>
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setRouteModalOpen(false)}>
              取消
            </Button>
            <Button
              color="primary"
              isLoading={saving}
              onPress={() => void submitRoute()}
            >
              {routeForm.id ? "保存" : "创建"}
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal
        isOpen={Boolean(diagnosisRoute)}
        scrollBehavior="inside"
        size="2xl"
        onOpenChange={(open) => {
          if (!open) {
            setDiagnosisRoute(null);
            setDiagnosis(null);
          }
        }}
      >
        <ModalContent>
          <ModalHeader>链路探测 · {diagnosisRoute?.name}</ModalHeader>
          <ModalBody className="space-y-4">
            <div className="flex items-start justify-between gap-4 rounded-md border border-divider p-3">
              <div className="min-w-0">
                <div className="text-sm font-medium">受控 Echo payload</div>
                <div className="mt-1 text-xs text-default-500">
                  仅在 Landing 是受控 TCP/UDP Echo 服务时启用；否则 UDP 保持
                  UNVERIFIED
                </div>
              </div>
              <Switch isSelected={payloadEcho} onValueChange={setPayloadEcho} />
            </div>
            <DiagnosisResultView
              businessType="多端口转发"
              items={diagnosisItems}
              measuredAt={diagnosis?.completedTime}
              overallStatus={diagnosis?.overallStatus}
              ruleName={diagnosisRoute?.name || "-"}
            />
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setDiagnosisRoute(null)}>
              关闭
            </Button>
            <Button
              color="primary"
              isLoading={Boolean(
                diagnosisRoute && busyRouteId === diagnosisRoute.id,
              )}
              startContent={<Activity className="w-4 h-4" />}
              onPress={() => void runDiagnosis()}
            >
              开始探测
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal
        isOpen={Boolean(deleteRoute)}
        onOpenChange={(open) => !open && setDeleteRoute(null)}
      >
        <ModalContent>
          <ModalHeader>删除多端口业务</ModalHeader>
          <ModalBody>
            确认删除“{deleteRoute?.name}”？端口只会在两段 Listener
            清理确认后释放。
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setDeleteRoute(null)}>
              取消
            </Button>
            <Button
              color="danger"
              isLoading={Boolean(deleteRoute && busyRouteId === deleteRoute.id)}
              onPress={() => void confirmDeleteRoute()}
            >
              删除
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal
        isOpen={Boolean(deletePath)}
        onOpenChange={(open) => !open && setDeletePath(null)}
      >
        <ModalContent>
          <ModalHeader>删除路径</ModalHeader>
          <ModalBody>
            确认删除“{deletePath?.name}”？存在活动业务时系统会拒绝删除。
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setDeletePath(null)}>
              取消
            </Button>
            <Button
              color="danger"
              isLoading={saving}
              onPress={() => void confirmDeletePath()}
            >
              删除
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </AnimatedPage>
  );
}
