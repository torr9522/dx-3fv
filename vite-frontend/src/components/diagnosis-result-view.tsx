import { useMemo } from "react";

import { Alert } from "@/shadcn-bridge/heroui/alert";
import { Chip } from "@/shadcn-bridge/heroui/chip";
import { Spinner } from "@/shadcn-bridge/heroui/spinner";

export type DiagnosisDisplayStatus =
  | "PASS"
  | "FAIL"
  | "UNVERIFIED"
  | "NOT_PROBED"
  | "RUNNING";

export interface DiagnosisDisplayItem {
  id: string;
  section: string;
  path: string;
  target?: string;
  protocol?: string;
  method?: string;
  status: DiagnosisDisplayStatus;
  latencyMs?: number;
  packetLoss?: number;
  message?: string;
}

interface DiagnosisResultViewProps {
  businessType: string;
  items: DiagnosisDisplayItem[];
  loading?: boolean;
  measuredAt?: number;
  overallStatus?: string;
  ruleName: string;
  timedOut?: boolean;
}

const statusColor = (status: DiagnosisDisplayStatus) => {
  if (status === "PASS") return "success" as const;
  if (status === "FAIL") return "danger" as const;
  if (status === "RUNNING") return "primary" as const;
  return "warning" as const;
};

const quality = (latency?: number, loss?: number) => {
  if (latency === undefined || loss === undefined) return null;
  if (latency < 50 && loss === 0) return { text: "优秀", color: "success" as const };
  if (latency < 100 && loss < 1) return { text: "良好", color: "primary" as const };
  if (latency < 200 && loss < 5) return { text: "一般", color: "warning" as const };
  return { text: "较差", color: "danger" as const };
};

export function DiagnosisResultView({
  businessType,
  items,
  loading = false,
  measuredAt,
  overallStatus,
  ruleName,
  timedOut = false,
}: DiagnosisResultViewProps) {
  const groups = useMemo(() => {
    const result = new Map<string, DiagnosisDisplayItem[]>();
    for (const item of items) {
      const group = result.get(item.section) || [];
      group.push(item);
      result.set(item.section, group);
    }
    return Array.from(result.entries());
  }, [items]);
  const success = items.filter((item) => item.status === "PASS").length;
  const failed = items.filter((item) => item.status === "FAIL").length;
  const pending = items.length - success - failed;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <Chip color="primary" size="sm" variant="flat">{businessType}</Chip>
        <span className="min-w-0 flex-1 truncate text-sm font-medium">{ruleName}</span>
        {overallStatus && (
          <Chip color={overallStatus === "PASS" ? "success" : overallStatus === "FAIL" ? "danger" : "warning"} size="sm" variant="flat">
            {overallStatus}
          </Chip>
        )}
        {measuredAt ? (
          <span className="text-xs text-default-500">{new Date(measuredAt).toLocaleString()}</span>
        ) : null}
      </div>

      {loading && (
        <div className="flex items-center gap-2 rounded-md border border-primary/20 bg-primary/5 px-3 py-2 text-sm text-primary">
          <Spinner size="sm" />
          <span>正在诊断</span>
        </div>
      )}
      {timedOut && (
        <Alert color="warning" description="部分测试未在时限内返回，以下保留已完成证据。" title="诊断超时" variant="flat" />
      )}

      <div className="grid grid-cols-3 gap-2 sm:gap-3">
        <div className="rounded-md border border-divider bg-default-50 p-3 text-center">
          <div className="text-xl font-semibold">{items.length}</div>
          <div className="mt-1 text-xs text-default-500">总测试数</div>
        </div>
        <div className="rounded-md border border-success-200 bg-success-50 p-3 text-center dark:bg-success-900/20">
          <div className="text-xl font-semibold text-success-600">{success}</div>
          <div className="mt-1 text-xs text-success-600">成功</div>
        </div>
        <div className="rounded-md border border-danger-200 bg-danger-50 p-3 text-center dark:bg-danger-900/20">
          <div className="text-xl font-semibold text-danger-600">{failed}</div>
          <div className="mt-1 text-xs text-danger-600">失败</div>
        </div>
      </div>
      {pending > 0 && <div className="text-xs text-default-500">另有 {pending} 项未验证、未执行或进行中</div>}

      {items.length === 0 ? (
        <div className="py-12 text-center text-sm text-default-500">暂无诊断数据</div>
      ) : (
        <div className="space-y-3">
          {groups.map(([section, sectionItems]) => (
            <section key={section} className="overflow-hidden rounded-md border border-divider">
              <h3 className="border-b border-divider bg-default-50 px-3 py-2 text-sm font-semibold">{section}</h3>
              <div className="divide-y divide-divider">
                {sectionItems.map((item) => {
                  const grade = item.status === "PASS" ? quality(item.latencyMs, item.packetLoss) : null;
                  return (
                    <div key={item.id} className="grid gap-3 p-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          {item.protocol && <Chip size="sm" variant="bordered">{item.protocol.toUpperCase()}</Chip>}
                          <Chip color={statusColor(item.status)} size="sm" variant="flat">{item.status}</Chip>
                          <span className="break-words text-sm font-medium">{item.path}</span>
                        </div>
                        <div className="mt-1 break-all text-xs text-default-500">
                          {[item.target, item.method, item.message].filter(Boolean).join(" · ")}
                        </div>
                      </div>
                      <div className="flex min-w-0 items-center gap-3 text-xs md:justify-end">
                        <span>{item.status === "PASS" && item.latencyMs !== undefined ? `${item.latencyMs.toFixed(1)} ms` : "-"}</span>
                        <span>{item.status === "PASS" && item.packetLoss !== undefined ? `${item.packetLoss.toFixed(1)}%` : "-"}</span>
                        {grade ? <Chip color={grade.color} size="sm" variant="flat">{grade.text}</Chip> : null}
                      </div>
                    </div>
                  );
                })}
              </div>
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
