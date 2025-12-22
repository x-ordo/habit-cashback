import React from "react";
import { Button } from "@toss-design-system/mobile";

export default function BottomCTA({
  label,
  onClick,
  disabled,
}: {
  label: string;
  onClick: () => void;
  disabled?: boolean;
}) {
  return (
    <div
      style={{
        position: "fixed",
        left: 0,
        right: 0,
        bottom: 0,
        padding: 16,
        background: "rgba(242,244,246,0.92)",
        backdropFilter: "blur(10px)",
      }}
    >
      <Button size="large" style={{ width: "100%" }} disabled={disabled} onClick={onClick}>
        {label}
      </Button>
    </div>
  );
}
