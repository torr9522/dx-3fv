import { useNavigate } from "react-router-dom";
import { ArrowRight, Network, Route, Signpost } from "lucide-react";

import { AnimatedPage } from "@/components/animated-page";
import { Button } from "@/shadcn-bridge/heroui/button";

const entries = [
  { path: "/single-forward", label: "单端口转发", description: "前置入口直接连接落地", icon: Signpost },
  { path: "/forward", label: "隧道转发", description: "使用 Tunnel 的 Forward 业务规则", icon: Network },
  { path: "/multiport", label: "多端口转发", description: "前置入口经 IX 转发到落地", icon: Route },
];

export default function ForwardHubPage() {
  const navigate = useNavigate();
  return (
    <AnimatedPage>
      <div className="px-3 py-8 lg:px-6">
        <h1 className="mb-5 text-xl font-semibold">转发</h1>
        <div className="space-y-2">
          {entries.map(({ path, label, description, icon: Icon }) => (
            <Button key={path} className="h-auto w-full justify-between px-4 py-4" variant="flat" onPress={() => navigate(path)}>
              <span className="flex min-w-0 items-center gap-3 text-left">
                <Icon className="h-5 w-5 shrink-0" />
                <span><span className="block font-medium">{label}</span><span className="block text-xs text-default-500">{description}</span></span>
              </span>
              <ArrowRight className="h-4 w-4 shrink-0" />
            </Button>
          ))}
        </div>
      </div>
    </AnimatedPage>
  );
}
