"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState, type FormEvent } from "react";
import type { APIError, BookingItem } from "@/lib/types";

export default function BookingForm({ eventId, available }: { eventId: number; available: number }) {
  const router = useRouter();
  const [bookedQuantity, setBookedQuantity] = useState(0);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  useEffect(() => {
    fetch("/api/bookings/me", { cache: "no-store" })
      .then((response) => response.ok ? response.json() : null)
      .then((data: { items: BookingItem[] } | null) => {
        const current = data?.items.find((booking) => booking.event_id === eventId && booking.status === 1);
        setBookedQuantity(current?.quantity ?? 0);
      });
  }, [eventId]);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true); setError(""); setMessage("");
    const form = new FormData(event.currentTarget);
    const response = await fetch(`/api/events/${eventId}/bookings`, {
      method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ quantity: Number(form.get("quantity")) }),
    });
    if (!response.ok) {
      const data = (await response.json()) as APIError;
      setError(response.status === 401 ? "กรุณาเข้าสู่ระบบก่อนจอง" : data.error.message);
      setPending(false);
      return;
    }
    const data = (await response.json()) as { booking: { quantity: number } };
    setBookedQuantity(data.booking.quantity);
    setMessage(`จองสำเร็จแล้ว ขณะนี้คุณจองรวม ${data.booking.quantity} ที่นั่ง`);
    setPending(false);
    router.refresh();
  }

  const remainingQuota = Math.max(0, 4 - bookedQuantity);
  const maxAdditional = Math.min(remainingQuota, available);

  return (
    <form className="card stack" onSubmit={submit}>
      <div><span className="eyebrow" style={{ color: "var(--brand)" }}>Reserve</span><h2 style={{ fontSize: 27 }}>จองที่นั่ง</h2></div>
      <p className="meta">จองเพิ่มได้หลายครั้ง โดยยอดรวมต่อผู้ใช้ต่ออีเวนต์ต้องไม่เกิน 4 ที่นั่ง</p>
      {bookedQuantity > 0 && <p className="seats">จองแล้ว {bookedQuantity} ที่ · เพิ่มได้อีก {remainingQuota} ที่</p>}
      {message && <p className="message">{message} · <Link href="/bookings" style={{ fontWeight: 800 }}>เปิดรายการจอง</Link></p>}
      {error && <p className="message error" role="alert">{error}</p>}
      <div className="field"><label htmlFor="quantity">จำนวนที่ต้องการจองเพิ่ม</label><input className="input" id="quantity" name="quantity" type="number" min={1} max={Math.max(1, maxAdditional)} defaultValue={1} disabled={maxAdditional === 0} required /></div>
      <button className="btn" disabled={pending || maxAdditional === 0}>{available === 0 ? "ที่นั่งเต็ม" : remainingQuota === 0 ? "จองครบ 4 ที่แล้ว" : pending ? "กำลังจอง…" : bookedQuantity > 0 ? "เพิ่มการจอง" : "ยืนยันการจอง"}</button>
    </form>
  );
}
