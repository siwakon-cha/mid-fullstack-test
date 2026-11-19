"use client";

import { useState } from "react";
import type { APIError } from "@/lib/types";

export default function CancelButton({ bookingId }: { bookingId: number }) {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  async function cancel() {
    if (!window.confirm("ยืนยันการยกเลิกการจอง? ที่นั่งจะถูกคืนกลับทันที")) return;
    setPending(true); setError("");
    const response = await fetch(`/api/bookings/${bookingId}`, { method: "DELETE" });
    if (!response.ok) {
      const data = (await response.json()) as APIError;
      setError(data.error.message); setPending(false); return;
    }
    window.location.reload();
  }
  return <div className="stack" style={{ gap: 6 }}><button className="btn danger" type="button" onClick={cancel} disabled={pending}>{pending ? "กำลังยกเลิก…" : "ยกเลิกการจอง"}</button>{error && <small style={{ color: "var(--danger)" }}>{error}</small>}</div>;
}
