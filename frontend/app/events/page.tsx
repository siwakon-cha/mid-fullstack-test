import EventCard from "@/components/EventCard";
import EventListControls, { EventPagination } from "@/components/EventListControls";
import { apiJSON } from "@/lib/api";
import type { EventListResponse } from "@/lib/types";

type Props = { searchParams: Promise<{ page?: string; search?: string }> };

export default async function EventsPage({ searchParams }: Props) {
  const query = await searchParams;
  const page = Math.max(1, Number(query.page) || 1);
  const search = (query.search ?? "").trim();
  const data = await apiJSON<EventListResponse>(`/api/events?page=${page}&pageSize=6&search=${encodeURIComponent(search)}`);
  const pages = Math.max(1, Math.ceil(data.total / data.page_size));

  return (
    <main className="shell page stack" style={{ gap: 34 }}>
      <section className="hero">
        <span className="eyebrow">Book with confidence</span>
        <h1>ระบบจองตั๋วอีเวนต์<br />(Event Ticket Booking)</h1>
      </section>
      <section>
        <div style={{ display: "flex", justifyContent: "space-between", gap: 16, alignItems: "end", flexWrap: "wrap" }}><div><span className="eyebrow" style={{ color: "var(--brand)" }}>Upcoming</span><h2>อีเวนต์ที่กำลังจะมาถึง</h2></div><span className="meta">ทั้งหมด {data.total} รายการ</span></div>
        <EventListControls search={search} />
        {data.items.length ? <div className="grid">{data.items.map((event) => <EventCard key={event.id} event={event} />)}</div> : <div className="empty"><h3>ไม่พบอีเวนต์</h3><p>ลองเปลี่ยนคำค้นหรือกลับมาดูใหม่ภายหลัง</p></div>}
        <EventPagination page={page} pages={pages} search={search} />
      </section>
    </main>
  );
}
