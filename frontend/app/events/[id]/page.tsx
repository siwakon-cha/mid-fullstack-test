import { notFound } from "next/navigation";
import BookingForm from "@/components/BookingForm";
import { formatDate } from "@/components/EventCard";
import { apiFetch } from "@/lib/api";
import type { EventItem } from "@/lib/types";

type Props = { params: Promise<{ id: string }> };

export default async function EventDetailPage({ params }: Props) {
  const { id } = await params;
  const response = await apiFetch(`/api/events/${id}`);
  if (response.status === 404) notFound();
  if (!response.ok) throw new Error(`event API returned ${response.status}`);
  const { event } = (await response.json()) as { event: EventItem };
  return (
    <main className="shell page detail-grid">
      <article className="card stack" style={{ padding: "clamp(24px, 5vw, 52px)" }}>
        <div><span className="seats">เหลือ {event.available_seats} จาก {event.total_seats} ที่</span><h1 style={{ fontSize: "clamp(32px, 6vw, 54px)" }}>{event.title}</h1></div>
        <p className="lead" style={{ color: "var(--muted)" }}>{event.description || "ยังไม่มีรายละเอียดเพิ่มเติม"}</p>
        <div className="form-grid"><div><span className="eyebrow" style={{ color: "var(--brand)" }}>วันและเวลา</span><p>{formatDate(event.starts_at)}</p></div><div><span className="eyebrow" style={{ color: "var(--brand)" }}>สถานที่</span><p>{event.location || "รอประกาศ"}</p></div></div>
      </article>
      <BookingForm eventId={event.id} available={event.available_seats} />
    </main>
  );
}
