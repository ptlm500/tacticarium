"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface GridFloorProps extends React.HTMLAttributes<HTMLDivElement> {
  perspective?: boolean
  animated?: boolean
}

export function GridFloor({
  children,
  className,
  perspective = true,
  animated = true,
  ...props
}: GridFloorProps) {
  return (
    <div className={cn("relative min-h-screen overflow-hidden", className)} {...props}>
      {/* Grid background */}
      <div
        className={cn(
          "pointer-events-none absolute inset-0",
          perspective && "perspective-grid"
        )}
      >
        <div
          className={cn(
            "absolute inset-0 grid-floor opacity-20",
            animated && "animate-grid-scroll"
          )}
        />
      </div>

      {/* Gradient overlay for depth */}
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-background via-transparent to-background" />

      {/* Content */}
      <div className="relative z-10">{children}</div>

    </div>
  )
}
