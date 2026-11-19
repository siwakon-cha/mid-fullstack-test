import Link from "next/link";
import type { EventItem } from "@/lib/types";

export function formatDate(value: string): string {
  return new Intl.DateTimeFormat("th-TH", { dateStyle: "medium", timeStyle: "short", timeZone: "Asia/Bangkok" }).format(new Date(value));
}

export default function EventCard({ event }: { event: EventItem }) {
  return (
    <article className="card">
      <span className="seats">เหลือ {event.available_seats} ที่</span>
      <h3>{event.title}</h3>
      <p className="meta">{formatDate(event.starts_at)}<br />{event.location || "รอประกาศสถานที่"}</p>
      <Link className="btn secondary" href={`/events/${event.id}`}>ดูรายละเอียด</Link>
    </article>
  );
}
