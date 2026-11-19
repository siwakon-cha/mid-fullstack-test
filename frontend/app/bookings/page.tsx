import { cookies } from "next/headers";
import Link from "next/link";
import CancelButton from "@/components/CancelButton";
import { formatDate } from "@/components/EventCard";
import { apiFetch } from "@/lib/api";
import type { BookingItem } from "@/lib/types";

export default async function BookingsPage() {
  const cookie = (await cookies()).toString();
  const response = await apiFetch("/api/bookings/me", cookie);
  if (response.status === 401) {
    return <main className="shell page"><div className="empty"><h2>เข้าสู่ระบบเพื่อดูการจอง</h2><p>รายการจองของคุณจะถูกแสดงในหน้านี้</p><Link className="btn" href="/login">เข้าสู่ระบบ</Link></div></main>;
  }
  if (!response.ok) throw new Error(`bookings API returned ${response.status}`);
  const { items } = (await response.json()) as { items: BookingItem[] };
  return (
    <main className="shell page stack" style={{ gap: 28 }}>
      <div><span className="eyebrow" style={{ color: "var(--brand)" }}>Your tickets</span><h1 style={{ fontSize: "clamp(34px, 6vw, 52px)" }}>การจองของฉัน</h1></div>
      {items.length === 0 ? <div className="empty"><h3>ยังไม่มีการจอง</h3><p>เลือกอีเวนต์ที่สนใจแล้วเริ่มจองได้เลย</p><Link className="btn" href="/events">ดูอีเวนต์</Link></div> : <div className="stack">{items.map((booking) => (
        <article className="card" key={booking.id} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 20, flexWrap: "wrap" }}>
          <div><span className={booking.status === 1 ? "status-confirmed" : "status-cancelled"}>{booking.status === 1 ? "ยืนยันแล้ว" : "ยกเลิกแล้ว"}</span><h3>{booking.event_title}</h3><p className="meta">{formatDate(booking.event_starts_at)} · {booking.event_location}<br />จำนวน {booking.quantity} ที่นั่ง</p></div>
          {booking.status === 1 ? <CancelButton bookingId={booking.id} /> : <Link className="btn secondary" href={`/events/${booking.event_id}`}>จองใหม่</Link>}
        </article>
      ))}</div>}
    </main>
  );
}
