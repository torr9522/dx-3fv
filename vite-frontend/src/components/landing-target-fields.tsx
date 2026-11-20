import { Plus, Trash2 } from "lucide-react";

import { Button } from "@/shadcn-bridge/heroui/button";
import { Input } from "@/shadcn-bridge/heroui/input";
import {
  emptyLandingTarget,
  type LandingTarget,
} from "@/pages/forward/landing-target";

interface LandingTargetFieldsProps {
  allowMultiple?: boolean;
  errors?: Array<{ host?: string; port?: string }>;
  targets: LandingTarget[];
  onChange: (targets: LandingTarget[]) => void;
}

export function LandingTargetFields({
  allowMultiple = false,
  errors = [],
  targets,
  onChange,
}: LandingTargetFieldsProps) {
  const update = (index: number, values: Partial<LandingTarget>) => {
    onChange(targets.map((target, targetIndex) =>
      targetIndex === index ? { ...target, ...values } : target,
    ));
  };

  return (
    <div className="space-y-3">
      {targets.map((target, index) => (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_10rem_auto]" key={index}>
          <Input
            errorMessage={errors[index]?.host}
            isInvalid={Boolean(errors[index]?.host)}
            label={targets.length > 1 ? `落地 IP / 域名 ${index + 1}` : "落地 IP / 域名"}
            placeholder="192.0.2.10 或 example.com"
            value={target.host}
            variant="bordered"
            onChange={(event) => update(index, { host: event.target.value })}
          />
          <Input
            errorMessage={errors[index]?.port}
            isInvalid={Boolean(errors[index]?.port)}
            label="落地端口"
            max="65535"
            min="1"
            placeholder="443"
            type="number"
            value={target.port}
            variant="bordered"
            onChange={(event) => update(index, { port: event.target.value })}
          />
          {allowMultiple && targets.length > 1 ? (
            <Button
              isIconOnly
              aria-label={`删除落地目标 ${index + 1}`}
              className="self-center"
              color="danger"
              variant="light"
              onPress={() => onChange(targets.filter((_, targetIndex) => targetIndex !== index))}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          ) : null}
        </div>
      ))}
      {allowMultiple ? (
        <Button
          size="sm"
          startContent={<Plus className="h-4 w-4" />}
          variant="flat"
          onPress={() => onChange([...targets, emptyLandingTarget()])}
        >
          添加落地目标
        </Button>
      ) : null}
    </div>
  );
}
