import type { MouseEvent } from "react";

import { Copy } from "lucide-react";
import toast from "react-hot-toast";

interface CopyableEntryValueProps {
  label: string;
  value: string | number | null | undefined;
}

const copyWithFallback = async (text: string): Promise<boolean> => {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text);

    return true;
  }

  const textarea = document.createElement("textarea");

  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.appendChild(textarea);
  textarea.select();

  try {
    return document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
};

export function CopyableEntryValue({ label, value }: CopyableEntryValueProps) {
  const text =
    value === null || value === undefined ? "" : String(value).trim();

  if (!text) {
    return <span className="font-mono text-default-400">未配置</span>;
  }

  const copy = async (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    event.stopPropagation();

    try {
      if (!(await copyWithFallback(text))) {
        throw new Error("copy command was rejected");
      }
      toast.success(`${label}已复制`);
    } catch {
      toast.error(`${label}复制失败，请手动选择`);
    }
  };

  return (
    <button
      aria-label={`复制${label} ${text}`}
      className="inline-flex min-w-0 items-center gap-1 rounded px-1 py-0.5 font-mono text-foreground hover:bg-default-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
      title={`复制${label}`}
      type="button"
      onClick={copy}
    >
      <span className="break-all">{text}</span>
      <Copy aria-hidden="true" className="h-3 w-3 shrink-0 text-default-400" />
    </button>
  );
}
