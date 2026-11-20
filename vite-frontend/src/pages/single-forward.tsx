import { useCallback, useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";
import { Activity, Pencil, Play, Plus, Square, Trash2 } from "lucide-react";

import {
  createForward,
  deleteForward,
  diagnoseForward,
  getForwardList,
  getNodeList,
  pauseForwardService,
  resumeForwardService,
  updateForward,
} from "@/api";
import type { ForwardApiItem, NodeApiItem } from "@/api/types";
import { AnimatedPage } from "@/components/animated-page";
import { LandingTargetFields } from "@/components/landing-target-fields";
import {
  DiagnosisResultView,
  type DiagnosisDisplayItem,
} from "@/components/diagnosis-result-view";
import { PageLoadingState } from "@/components/page-state";
import type { ForwardDiagnosisResult } from "@/pages/forward/diagnosis";
import {
  emptyLandingTarget,
  formatLandingTargets,
  parseLandingTargets,
  type LandingTarget,
  validateLandingTargets,
} from "@/pages/forward/landing-target";
import { formatTraffic } from "@/utils/traffic";
import { Button } from "@/shadcn-bridge/heroui/button";
import { Card, CardBody } from "@/shadcn-bridge/heroui/card";
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

type ForwardProtocol = "tcp" | "udp" | "tcp+udp";

interface DirectForward extends ForwardApiItem {
  entryNodeId: number;
  protocol: ForwardProtocol;
}

interface FormState {
  id?: number;
  name: string;
  entryNodeId: number | null;
  inPort: number | null;
  landingTargets: LandingTarget[];
  protocol: ForwardProtocol;
}

const emptyForm = (): FormState => ({
  name: "",
  entryNodeId: null,
  inPort: null,
  landingTargets: [emptyLandingTarget()],
  protocol: "tcp",
});

const normalizeDirectForward = (item: ForwardApiItem): DirectForward => ({
  ...item,
  entryNodeId: Number(item.entryNodeId || 0),
  protocol:
    item.protocol === "udp" || item.protocol === "tcp+udp"
      ? item.protocol
      : "tcp",
});

export default function SingleForwardPage() {
  const [items, setItems] = useState<DirectForward[]>([]);
  const [nodes, setNodes] = useState<NodeApiItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [landingErrors, setLandingErrors] = useState<Array<{ host?: string; port?: string }>>([]);
  const [diagnosisItem, setDiagnosisItem] = useState<DirectForward | null>(null);
  const [diagnosis, setDiagnosis] = useState<ForwardDiagnosisResult | null>(null);
  const [diagnosisLoading, setDiagnosisLoading] = useState(false);

  const nodeById = useMemo(
    () => new Map(nodes.map((node) => [node.id, node])),
    [nodes],
  );
  const diagnosisItems = useMemo<DiagnosisDisplayItem[]>(() =>
    (diagnosis?.results || []).map((result, index) => {
      const status = result.diagnosing
        ? "RUNNING"
        : result.status || (result.success ? "PASS" : "FAIL");
      return {
        id: `${result.nodeId}-${result.targetIp}-${result.targetPort || 0}-${index}`,
        section: "入口到落地",
        path: result.description,
        target: `${result.targetIp}:${result.targetPort || "-"}`,
        protocol: result.protocol || "tcp",
        method: result.method || "tcp_connect",
        status,
        latencyMs: status === "PASS" ? result.averageTime : undefined,
        packetLoss: status === "PASS" ? result.packetLoss : undefined,
        message: result.message,
      };
    }), [diagnosis]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [forwardResult, nodeResult] = await Promise.all([
        getForwardList(),
        getNodeList(),
      ]);
      if (forwardResult.code !== 0) {
        throw new Error(forwardResult.msg || "获取单端口转发失败");
      }
      if (nodeResult.code !== 0) {
        throw new Error(nodeResult.msg || "获取节点失败");
      }
      setItems(
        (forwardResult.data || [])
          .filter((item) => Number(item.tunnelId || 0) === 0)
          .map(normalizeDirectForward),
      );
      setNodes((nodeResult.data || []).filter((node) => node.forwardMode !== "nftables"));
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "加载失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const openCreate = () => {
    setForm(emptyForm());
    setErrors({});
    setLandingErrors([]);
    setModalOpen(true);
  };

  const openEdit = (item: DirectForward) => {
    setForm({
      id: item.id,
      name: item.name,
      entryNodeId: item.entryNodeId,
      inPort: item.inPort || null,
      landingTargets: parseLandingTargets(item.remoteAddr || ""),
      protocol: item.protocol,
    });
    setErrors({});
    setLandingErrors([]);
    setModalOpen(true);
  };

  const validate = () => {
    const next: Record<string, string> = {};
    if (form.name.trim().length < 2 || form.name.trim().length > 50) {
      next.name = "规则名称长度应为 2-50 个字符";
    }
    if (!form.entryNodeId) next.entryNodeId = "请选择前置入口";
    if (form.inPort !== null && (form.inPort < 1 || form.inPort > 65535)) {
      next.inPort = "入口端口必须为 1-65535";
    }
    const targetValidation = validateLandingTargets(form.landingTargets);
    setLandingErrors(targetValidation.fieldErrors);
    if (!targetValidation.valid) next.landingTargets = "落地目标格式错误";
    setErrors(next);
    return Object.keys(next).length === 0;
  };

  const submit = async () => {
    if (!validate()) return;
    setSaving(true);
    try {
      const payload = {
        ...(form.id ? { id: form.id } : {}),
        name: form.name.trim(),
        tunnelId: 0,
        entryNodeId: form.entryNodeId || undefined,
        inPort: form.inPort,
        remoteAddr: formatLandingTargets(form.landingTargets),
        protocol: form.protocol,
        strategy: "fifo",
      };
      const result = form.id
        ? await updateForward(payload)
        : await createForward(payload);
      if (result.code !== 0) throw new Error(result.msg || "保存失败");
      toast.success(form.id ? "单端口转发已更新" : "单端口转发已创建");
      setModalOpen(false);
      await load();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  const toggle = async (item: DirectForward) => {
    const result = item.status === 1
      ? await pauseForwardService(item.id)
      : await resumeForwardService(item.id);
    if (result.code !== 0) {
      toast.error(result.msg || "状态修改失败");
      return;
    }
    toast.success(item.status === 1 ? "已停用" : "已启用");
    await load();
  };

  const remove = async (item: DirectForward) => {
    if (!window.confirm(`确认删除单端口转发“${item.name}”？`)) return;
    const result = await deleteForward(item.id);
    if (result.code !== 0) {
      toast.error(result.msg || "删除失败");
      return;
    }
    toast.success("删除成功");
    await load();
  };

  const runDiagnosis = async (item: DirectForward) => {
    setDiagnosisItem(item);
    setDiagnosisLoading(true);
    const result = await diagnoseForward(item.id);
    setDiagnosisLoading(false);
    if (result.code !== 0 || !result.data) {
      toast.error(result.msg || "诊断失败");
      return;
    }
    setDiagnosis(result.data as ForwardDiagnosisResult);
  };

  if (loading) return <PageLoadingState message="正在加载单端口转发..." />;

  return (
    <AnimatedPage>
      <div className="px-3 py-6 lg:px-6">
        <div className="mb-5 flex items-center justify-between gap-3">
          <div>
            <h1 className="text-xl font-semibold">单端口转发</h1>
            <p className="mt-1 text-sm text-default-500">前置入口直接转发到一个落地目标</p>
          </div>
          <Button color="primary" size="sm" startContent={<Plus className="h-4 w-4" />} onPress={openCreate}>
            新增
          </Button>
        </div>

        {items.length === 0 ? (
          <div className="py-16 text-center text-sm text-default-500">暂无单端口转发</div>
        ) : (
          <div className="grid gap-3 xl:grid-cols-2">
            {items.map((item) => (
              <Card key={item.id} className="border border-divider bg-content1 shadow-sm">
                <CardBody className="gap-3 p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <div className="truncate font-medium">{item.name}</div>
                      <div className="mt-1 text-xs text-default-500">
                        {nodeById.get(item.entryNodeId)?.name || `Node ${item.entryNodeId}`}:{item.inPort || "-"}
                        <span className="px-2">→</span>{item.remoteAddr}
                      </div>
                    </div>
                    <Chip color={item.status === 1 ? "success" : "default"} size="sm" variant="flat">
                      {item.status === 1 ? "运行中" : "已停用"}
                    </Chip>
                  </div>
                  <div className="flex flex-wrap items-center gap-2 text-xs text-default-500">
                    <Chip size="sm" variant="bordered">{item.protocol.toUpperCase()}</Chip>
                    <span>累计 {formatTraffic(Number(item.inFlow || 0) + Number(item.outFlow || 0))}</span>
                  </div>
                  <div className="flex justify-end gap-1">
                    <Button isIconOnly aria-label="诊断" size="sm" variant="light" onPress={() => void runDiagnosis(item)}><Activity className="h-4 w-4" /></Button>
                    <Button isIconOnly aria-label="编辑" size="sm" variant="light" onPress={() => openEdit(item)}><Pencil className="h-4 w-4" /></Button>
                    <Button isIconOnly aria-label={item.status === 1 ? "停用" : "启用"} size="sm" variant="light" onPress={() => void toggle(item)}>
                      {item.status === 1 ? <Square className="h-4 w-4" /> : <Play className="h-4 w-4" />}
                    </Button>
                    <Button isIconOnly aria-label="删除" color="danger" size="sm" variant="light" onPress={() => void remove(item)}><Trash2 className="h-4 w-4" /></Button>
                  </div>
                </CardBody>
              </Card>
            ))}
          </div>
        )}
      </div>

      <Modal isOpen={modalOpen} placement="center" scrollBehavior="inside" size="lg" onOpenChange={setModalOpen}>
        <ModalContent>
          <ModalHeader>{form.id ? "编辑单端口转发" : "新增单端口转发"}</ModalHeader>
          <ModalBody className="gap-4">
            <Input errorMessage={errors.name} isInvalid={Boolean(errors.name)} label="规则名称" value={form.name} variant="bordered" onChange={(event) => setForm((current) => ({ ...current, name: event.target.value }))} />
            <Select errorMessage={errors.entryNodeId} isInvalid={Boolean(errors.entryNodeId)} label="前置入口" selectedKeys={form.entryNodeId ? [String(form.entryNodeId)] : []} variant="bordered" onSelectionChange={(keys) => setForm((current) => ({ ...current, entryNodeId: Number(Array.from(keys)[0] || 0) || null }))}>
              {nodes.map((node) => <SelectItem key={String(node.id)} textValue={node.name}>{node.name} ({node.status === 1 ? "在线" : "离线"})</SelectItem>)}
            </Select>
            <Input description="留空自动分配；SSH 和实际占用端口不会被抢占" errorMessage={errors.inPort} isInvalid={Boolean(errors.inPort)} label="前置入口端口" placeholder="自动分配" type="number" value={form.inPort ? String(form.inPort) : ""} variant="bordered" onChange={(event) => setForm((current) => ({ ...current, inPort: event.target.value ? Number(event.target.value) : null }))} />
            <LandingTargetFields
              errors={landingErrors}
              targets={form.landingTargets}
              onChange={(landingTargets) => setForm((current) => ({ ...current, landingTargets }))}
            />
            <Select label="协议" selectedKeys={[form.protocol]} variant="bordered" onSelectionChange={(keys) => setForm((current) => ({ ...current, protocol: String(Array.from(keys)[0]) as ForwardProtocol }))}>
              <SelectItem key="tcp">TCP</SelectItem>
              <SelectItem key="udp">UDP</SelectItem>
              <SelectItem key="tcp+udp">TCP+UDP</SelectItem>
            </Select>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={() => setModalOpen(false)}>取消</Button>
            <Button color="primary" isLoading={saving} onPress={() => void submit()}>{form.id ? "保存修改" : "创建规则"}</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      <Modal isOpen={Boolean(diagnosisItem)} placement="center" scrollBehavior="inside" size="2xl" onOpenChange={(open) => {
        if (!open) {
          setDiagnosisItem(null);
          setDiagnosis(null);
        }
      }}>
        <ModalContent>
          <ModalHeader>规则诊断结果</ModalHeader>
          <ModalBody>
            <DiagnosisResultView
              businessType="单端口转发"
              items={diagnosisItems}
              loading={diagnosisLoading}
              measuredAt={diagnosis?.timestamp}
              ruleName={diagnosisItem?.name || "-"}
            />
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={() => setDiagnosisItem(null)}>关闭</Button>
            {diagnosisItem && <Button color="primary" isLoading={diagnosisLoading} onPress={() => void runDiagnosis(diagnosisItem)}>重新诊断</Button>}
          </ModalFooter>
        </ModalContent>
      </Modal>
    </AnimatedPage>
  );
}
