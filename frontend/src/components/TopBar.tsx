import React from "react";
import { useNavigate } from "react-router-dom";
import { Button, Text } from "@toss-design-system/mobile";

export default function TopBar({
  title,
  right,
  backTo,
}: {
  title: string;
  right?: React.ReactNode;
  backTo?: string;
}) {
  const nav = useNavigate();
  const onBack = () => {
    if (backTo) nav(backTo);
    else nav(-1);
  };

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        padding: "16px",
        gap: 12,
        borderBottom: "1px solid #e5e8eb",
        background: "white",
        position: "sticky",
        top: 0,
        zIndex: 10,
      }}
    >
      <Button size="small" onClick={onBack} style={{ minWidth: 44 }}>
        ←
      </Button>

      <div style={{ flex: 1 }}>
        <Text typography="t4" fontWeight="bold">
          {title}
        </Text>
      </div>

      <div>{right}</div>
    </div>
  );
}
