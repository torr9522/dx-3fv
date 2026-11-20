import { useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";
import { DatabaseZap, Plus, RefreshCw } from "lucide-react";

import {
  createIXPortPool,
  getIXPortLeases,
  getIXPortPools,
  getIXPortResources,
  getNodeList,
  upsertIXPortResource,
  type IXPortLeaseApiItem,
  type IXPortPoolApiItem,
  type IXPortResourceApiItem,
} from "@/api";
import { AnimatedPage } from "@/components/animated-page";
import { PageLoadingState } from "@/components/page-state";
import { Button } from "@/shadcn-bridge/heroui/button";
import { Card, CardBody } from "@/shadcn-bridge/heroui/card";
import { Checkbox } from "@/shadcn-bridge/heroui/checkbox";
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
import {
  Table,
  TableBody,
  TableCell,
  TableColumn,
  TableHeader,
  TableRow,
} from "@/shadcn-bridge/heroui/table";

interface NodeOption {
  id: number;
  name: string;
}

const emptyResource = {
  publicIp: "",
  publicPort: 0,
  internalIp: "",
  internalPort: 0,
  tcpMapped: true,
  udpMapped: false,
  verificationStatus: "unverified" as IXPortResourceApiItem["verificationStatus"],
  excluded: false,
  reservedReason: "",
};

export default function IXPortsPage() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [nodes, setNodes] = useState<NodeOption[]>([]);
  const [pools, setPools] = useState<IXPortPoolApiItem[]>([]);
  const [resources, setResources] = useState<IXPortResourceApiItem[]>([]);
  const [leases, setLeases] = useState<IXPortLeaseApiItem[]>([]);
  const [selectedPoolId, setSelectedPoolId] = useState<number>(0);
  const [poolModalOpen, setPoolModalOpen] = useState(false);
  const [resourceModalOpen, setResourceModalOpen] = useState(false);
  const [poolForm, setPoolForm] = useState({
    nodeId: 0,
    name: "",
    kind: "nat" as IXPortPoolApiItem["kind"],
    rangeStart: 1,
    rangeEnd: 65535,
    status: 1,
  });
  const [resourceForm, setResourceForm] = useState(emptyResource);

  const selectedPool = useMemo(
    () => pools.find((pool) => pool.id === selectedPoolId),
    [pools, selectedPoolId],
  );
  const activeLeases = useMemo(
    () => leases.filter((lease) => lease.state !== "released"),
    [leases],
  );

  const loadResources = async (poolId: number) => {
    if (!poolId) {
      setResources([]);
      return;
    }
    const response = await getIXPortResources(poolId);
    if (response.code !== 0) throw new Error(response.msg);
    setResources(response.data || []);
  };

  const loadAll = async () => {
    setLoading(true);
    try {
      const [poolResponse, nodeResponse, leaseResponse] = await Promise.all([
        getIXPortPools(),
        getNodeList(),
        getIXPortLeases(),
      ]);
      if (poolResponse.code !== 0) throw new Error(poolResponse.msg);
      if (nodeResponse.code !== 0) throw new Error(nodeResponse.msg);
      if (leaseResponse.code !== 0) throw new Error(leaseResponse.msg);
      const nextPools = poolResponse.data || [];
      const nextPoolId = nextPools.some((pool) => pool.id === selectedPoolId)
        ? selectedPoolId
        : nextPools[0]?.id || 0;
      setPools(nextPools);
      setNodes((nodeResponse.data || []) as NodeOption[]);
      setLeases(leaseResponse.data || []);
      setSelectedPoolId(nextPoolId);
      await loadResources(nextPoolId);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载 IX 端口失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadAll();
  }, []);

  const changePool = async (poolId: number) => {
    setSelectedPoolId(poolId);
    try {
      await loadResources(poolId);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载端口映射失败");
    }
  };

  const submitPool = async () => {
    if (!poolForm.nodeId || !poolForm.name.trim()) {
      toast.error("请选择节点并填写端口池名称");
      return;
    }
    if (
      poolForm.kind === "nat" &&
      (poolForm.rangeStart < 1 ||
        poolForm.rangeEnd > 65535 ||
        poolForm.rangeStart > poolForm.rangeEnd)
    ) {
      toast.error("请输入有效的 NAT 端口范围");
      return;
    }
    setSaving(true);
    const response = await createIXPortPool({
      ...poolForm,
      rangeStart: poolForm.kind === "full" ? 1 : poolForm.rangeStart,
      rangeEnd: poolForm.kind === "full" ? 65535 : poolForm.rangeEnd,
    });
    setSaving(false);
    if (response.code !== 0) {
      toast.error(response.msg || "创建端口池失败");
      return;
    }
    toast.success("端口池已创建");
    setPoolModalOpen(false);
    await loadAll();
  };

  const submitResource = async () => {
    if (!selectedPool) return;
    if (!resourceForm.internalPort || !resourceForm.publicPort) {
      toast.error("公网端口和内部监听端口不能为空");
      return;
    }
    if (!resourceForm.tcpMapped && !resourceForm.udpMapped && !resourceForm.excluded) {
      toast.error("至少登记一种协议，或将端口标记为排除");
      return;
    }
    setSaving(true);
    const response = await upsertIXPortResource({
      ...resourceForm,
      poolId: selectedPool.id,
      nodeId: selectedPool.nodeId,
    });
    setSaving(false);
    if (response.code !== 0) {
      toast.error(response.msg || "保存端口资源失败");
      return;
    }
    toast.success("端口资源已保存");
    setResourceModalOpen(false);
    setResourceForm(emptyResource);
    await loadAll();
  };

  if (loading) return <PageLoadingState message="正在加载 IX 端口资源" />;

  return (
    <AnimatedPage className="px-3 lg:px-6 py-6 space-y-5">
      <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-xl font-semibold flex items-center gap-2">
            <DatabaseZap className="w-5 h-5" />
            IX 端口
          </h1>
          <p className="text-sm text-default-500 mt-1">
            NAT 映射、保留端口、协议能力与持久租约
          </p>
        </div>
        <div className="flex gap-2">
          <Button isIconOnly title="刷新" variant="flat" onPress={() => void loadAll()}>
            <RefreshCw className="w-4 h-4" />
          </Button>
          <Button color="primary" startContent={<Plus className="w-4 h-4" />} onPress={() => setPoolModalOpen(true)}>
            新增端口池
          </Button>
        </div>
      </div>

      {pools.length === 0 ? (
        <Card>
          <CardBody className="py-14 text-center text-default-500">
            暂无 IX 端口池
          </CardBody>
        </Card>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-[minmax(240px,360px)_1fr] gap-4 items-end">
            <Select
              label="节点端口池"
              selectedKeys={selectedPoolId ? [String(selectedPoolId)] : []}
              variant="bordered"
              onSelectionChange={(keys) => void changePool(Number(Array.from(keys)[0]))}
            >
              {pools.map((pool) => (
                <SelectItem key={String(pool.id)} textValue={pool.name}>
                  {pool.name}
                </SelectItem>
              ))}
            </Select>
            <div className="flex flex-wrap gap-2 pb-1">
              <Chip color="primary" variant="flat">
                {selectedPool?.kind === "full" ? "全端口策略" : `${selectedPool?.rangeStart}-${selectedPool?.rangeEnd}`}
              </Chip>
              <Chip variant="flat">资源 {resources.length}</Chip>
              <Chip variant="flat">
                使用中 {activeLeases.filter((lease) => lease.poolId === selectedPoolId).length}
              </Chip>
            </div>
          </div>

          <div className="flex justify-end">
            <Button
              color="primary"
              startContent={<Plus className="w-4 h-4" />}
              variant="flat"
              onPress={() => setResourceModalOpen(true)}
            >
              登记端口映射
            </Button>
          </div>

          <Table aria-label="IX 端口资源">
            <TableHeader>
              <TableColumn>公网端点</TableColumn>
              <TableColumn>内部监听</TableColumn>
              <TableColumn>协议</TableColumn>
              <TableColumn>核验</TableColumn>
              <TableColumn>租约</TableColumn>
            </TableHeader>
            <TableBody emptyContent="当前端口池还没有已登记资源">
              {resources.map((resource) => {
                const owned = activeLeases.filter((lease) => lease.resourceId === resource.id);
                return (
                  <TableRow key={resource.id}>
                    <TableCell>{resource.publicIp || "-"}:{resource.publicPort}</TableCell>
                    <TableCell>{resource.internalIp || "*"}:{resource.internalPort}</TableCell>
                    <TableCell>
                      <div className="flex gap-1">
                        {resource.tcpMapped && <Chip size="sm">TCP</Chip>}
                        {resource.udpMapped && <Chip size="sm">UDP</Chip>}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Chip color={resource.excluded ? "warning" : resource.verificationStatus === "verified" || resource.verificationStatus === "local" ? "success" : "default"} size="sm" variant="flat">
                        {resource.excluded ? resource.reservedReason || "已排除" : resource.verificationStatus}
                      </Chip>
                    </TableCell>
                    <TableCell>
                      {owned.length === 0 ? "空闲" : owned.map((lease) => `${lease.ownerType} #${lease.ownerId} ${lease.protocol.toUpperCase()}`).join(" / ")}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </>
      )}

      <Modal isOpen={poolModalOpen} onOpenChange={setPoolModalOpen}>
        <ModalContent>
          <ModalHeader>新增端口池</ModalHeader>
          <ModalBody className="space-y-4">
            <Select
              label="节点"
              selectedKeys={poolForm.nodeId ? [String(poolForm.nodeId)] : []}
              onSelectionChange={(keys) => setPoolForm((form) => ({ ...form, nodeId: Number(Array.from(keys)[0]) }))}
            >
              {nodes.map((node) => <SelectItem key={String(node.id)} textValue={node.name}>{node.name}</SelectItem>)}
            </Select>
            <Input label="名称" value={poolForm.name} onChange={(event) => setPoolForm((form) => ({ ...form, name: event.target.value }))} />
            <Select
              label="资源类型"
              selectedKeys={[poolForm.kind]}
              onSelectionChange={(keys) => setPoolForm((form) => ({ ...form, kind: Array.from(keys)[0] as IXPortPoolApiItem["kind"] }))}
            >
              <SelectItem key="full">直连节点全端口</SelectItem>
              <SelectItem key="nat">NAT 映射资源</SelectItem>
            </Select>
            {poolForm.kind === "nat" && (
              <div className="grid grid-cols-2 gap-3">
                <Input label="范围起点" type="number" value={String(poolForm.rangeStart)} onChange={(event) => setPoolForm((form) => ({ ...form, rangeStart: Number(event.target.value) }))} />
                <Input label="范围终点" type="number" value={String(poolForm.rangeEnd)} onChange={(event) => setPoolForm((form) => ({ ...form, rangeEnd: Number(event.target.value) }))} />
              </div>
            )}
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setPoolModalOpen(false)}>取消</Button>
            <Button color="primary" isLoading={saving} onPress={() => void submitPool()}>创建</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal isOpen={resourceModalOpen} scrollBehavior="inside" onOpenChange={setResourceModalOpen}>
        <ModalContent>
          <ModalHeader>登记端口映射</ModalHeader>
          <ModalBody className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <Input label="公网 IP" value={resourceForm.publicIp} onChange={(event) => setResourceForm((form) => ({ ...form, publicIp: event.target.value }))} />
              <Input label="公网端口" type="number" value={resourceForm.publicPort ? String(resourceForm.publicPort) : ""} onChange={(event) => setResourceForm((form) => ({ ...form, publicPort: Number(event.target.value) }))} />
              <Input label="内部 IP" value={resourceForm.internalIp} onChange={(event) => setResourceForm((form) => ({ ...form, internalIp: event.target.value }))} />
              <Input label="内部监听端口" type="number" value={resourceForm.internalPort ? String(resourceForm.internalPort) : ""} onChange={(event) => setResourceForm((form) => ({ ...form, internalPort: Number(event.target.value) }))} />
            </div>
            <div className="flex gap-5">
              <Checkbox isSelected={resourceForm.tcpMapped} onValueChange={(value) => setResourceForm((form) => ({ ...form, tcpMapped: value }))}>TCP 映射</Checkbox>
              <Checkbox isSelected={resourceForm.udpMapped} onValueChange={(value) => setResourceForm((form) => ({ ...form, udpMapped: value }))}>UDP 映射</Checkbox>
            </div>
            <Select
              label="NAT 核验状态"
              selectedKeys={[resourceForm.verificationStatus]}
              onSelectionChange={(keys) => setResourceForm((form) => ({ ...form, verificationStatus: Array.from(keys)[0] as IXPortResourceApiItem["verificationStatus"] }))}
            >
              <SelectItem key="unverified">未核验</SelectItem>
              <SelectItem key="verified">已核验</SelectItem>
              <SelectItem key="unavailable">不可用</SelectItem>
            </Select>
            <Checkbox isSelected={resourceForm.excluded} onValueChange={(value) => setResourceForm((form) => ({ ...form, excluded: value }))}>保留或排除此端口</Checkbox>
            {resourceForm.excluded && <Input label="保留原因" value={resourceForm.reservedReason} onChange={(event) => setResourceForm((form) => ({ ...form, reservedReason: event.target.value }))} />}
          </ModalBody>
          <ModalFooter>
            <Button variant="flat" onPress={() => setResourceModalOpen(false)}>取消</Button>
            <Button color="primary" isLoading={saving} onPress={() => void submitResource()}>保存</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </AnimatedPage>
  );
}
