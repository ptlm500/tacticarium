import DOMPurify from "dompurify";
import { useMemo } from "react";
import { cn } from "@/lib/utils";

const ALLOWED_TAGS = ["b", "strong", "i", "em", "u", "br", "p", "ul", "ol", "li", "span"];
const ALLOWED_ATTR = ["class"];

interface Props {
  html: string;
  className?: string;
}

export function RulesText({ html, className }: Props) {
  const clean = useMemo(() => DOMPurify.sanitize(html, { ALLOWED_TAGS, ALLOWED_ATTR }), [html]);
  return (
    <div className={cn("rules-text", className)} dangerouslySetInnerHTML={{ __html: clean }} />
  );
}
