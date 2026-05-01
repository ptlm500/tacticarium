import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn } from "@/lib/utils";

interface Props {
  avatarUrl?: string;
  username: string;
  size?: "xs" | "sm" | "md";
  className?: string;
}

const SIZES: Record<NonNullable<Props["size"]>, { wrap: string; fallback: string }> = {
  xs: { wrap: "size-4", fallback: "text-[8px]" },
  sm: { wrap: "size-6", fallback: "text-[10px]" },
  md: { wrap: "size-8", fallback: "text-xs" },
};

export function PlayerAvatar({ avatarUrl, username, size = "sm", className }: Props) {
  const s = SIZES[size];
  const label = `${username}'s avatar`;
  return (
    <Avatar aria-label={label} className={cn(s.wrap, className)}>
      {avatarUrl && <AvatarImage src={avatarUrl} alt={label} loading="lazy" decoding="async" />}
      <AvatarFallback className={cn("font-mono uppercase", s.fallback)}>
        {username ? username.charAt(0).toUpperCase() : "?"}
      </AvatarFallback>
    </Avatar>
  );
}
