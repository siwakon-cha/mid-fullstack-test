"use client";

import { useEffect, useMemo, useState, type FormEvent, type MouseEvent } from "react";
import type { APIError, EventItem, EventListResponse, User } from "@/lib/types";

const PAGE_SIZE = 10;
const emptyForm = { id: 0, title: "", description: "", location: "", starts_at: "", total_seats: 50 };

function formatEventDate(value: string) {
  return new Date(value).toLocaleString("th-TH", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function AdminEvents() {
  const [user, setUser] = useState<User | null>(null);
  const [events, setEvents] = useState<EventItem[]>([]);
  const [form, setForm] = useState(emptyForm);
  const [query, setQuery] = useState("");
  const [page, setPage] = useState(1);
  const [modalOpen, setModalOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function load() {
    const [meResponse, eventResponse] = await Promise.all([
      fetch("/api/auth/me", { cache: "no-store" }),
      fetch("/api/events?page=1&pageSize=50", { cache: "no-store" }),
    ]);
    if (meResponse.ok) setUser(((await meResponse.json()) as { user: User }).user);
    if (eventResponse.ok) setEvents(((await eventResponse.json()) as EventListResponse).items);
  }

  useEffect(() => {
    fetch("/api/auth/me", { cache: "no-store" })
      .then((response) => response.ok ? response.json() : null)
      .then((data: { user: User } | null) => setUser(data?.user ?? null));
    fetch("/api/events?page=1&pageSize=50", { cache: "no-store" })
      .then((response) => response.ok ? response.json() : null)
      .then((data: EventListResponse | null) => setEvents(data?.items ?? []));
  }, []);

  useEffect(() => {
    if (!modalOpen) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !pending) setModalOpen(false);
    };
    document.addEventListener("keydown", closeOnEscape);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [modalOpen, pending]);

  const totals = useMemo(() => events.reduce(
    (result, item) => ({
      seats: result.seats + item.total_seats,
      available: result.available + item.available_seats,
    }),
    { seats: 0, available: 0 },
  ), [events]);

  const filteredEvents = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase("th-TH");
    if (!normalizedQuery) return events;
    return events.filter((item) =>
      `${item.title} ${item.location}`.toLocaleLowerCase("th-TH").includes(normalizedQuery),
    );
  }, [events, query]);

  const totalPages = Math.max(1, Math.ceil(filteredEvents.length / PAGE_SIZE));
  const currentPage = Math.min(page, totalPages);
  const pageStart = (currentPage - 1) * PAGE_SIZE;
  const pageEvents = filteredEvents.slice(pageStart, pageStart + PAGE_SIZE);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    setMessage("");

    const method = form.id ? "PATCH" : "POST";
    const path = form.id ? `/api/admin/events/${form.id}` : "/api/admin/events";
    const response = await fetch(path, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ...form, starts_at: new Date(form.starts_at).toISOString() }),
    });

    if (!response.ok) {
      const data = (await response.json()) as APIError;
      setError(data.error.message);
      setPending(false);
      return;
    }

    setMessage(form.id ? "แก้ไขอีเวนต์แล้ว" : "สร้างอีเวนต์แล้ว");
    setModalOpen(false);
    setForm(emptyForm);
    setPending(false);
    await load();
  }

  function openCreateModal() {
    setForm(emptyForm);
    setError("");
    setMessage("");
    setModalOpen(true);
  }

  function openEditModal(item: EventItem) {
    setForm({
      id: item.id,
      title: item.title,
      description: item.description,
      location: item.location,
      starts_at: item.starts_at.slice(0, 16),
      total_seats: item.total_seats,
    });
    setError("");
    setMessage("");
    setModalOpen(true);
  }

  function closeModal() {
    if (!pending) setModalOpen(false);
  }

  function closeFromBackdrop(event: MouseEvent<HTMLDivElement>) {
    if (event.currentTarget === event.target) closeModal();
  }

  if (!user) return <div className="empty">กำลังตรวจสอบสิทธิ์…</div>;
  if (user.role !== 1) return <div className="empty"><h2>ไม่มีสิทธิ์เข้าถึง</h2><p>หน้านี้สำหรับผู้ดูแลระบบเท่านั้น</p></div>;

  return (
    <div className="admin-dashboard">
      <section className="admin-stats" aria-label="ภาพรวมอีเวนต์">
        <article className="admin-stat">
          <span className="admin-stat-label">อีเวนต์ทั้งหมด</span>
          <strong>{events.length}</strong>
          <span className="meta">รายการที่กำลังเปิดจอง</span>
        </article>
        <article className="admin-stat">
          <span className="admin-stat-label">ที่นั่งคงเหลือ</span>
          <strong>{totals.available.toLocaleString("th-TH")}</strong>
          <span className="meta">จาก {totals.seats.toLocaleString("th-TH")} ที่นั่ง</span>
        </article>
        <article className="admin-stat">
          <span className="admin-stat-label">จองแล้ว</span>
          <strong>{(totals.seats - totals.available).toLocaleString("th-TH")}</strong>
          <span className="meta">รวมทุกอีเวนต์</span>
        </article>
      </section>

      <section className="admin-event-section">
        <div className="admin-list-header">
          <div>
            <span className="eyebrow" style={{ color: "var(--brand)" }}>Event inventory</span>
            <h2>รายการอีเวนต์</h2>
            <p className="meta">
              {filteredEvents.length
                ? `แสดง ${pageStart + 1}–${Math.min(pageStart + PAGE_SIZE, filteredEvents.length)} จาก ${filteredEvents.length} รายการ`
                : "ไม่พบรายการ"}
            </p>
          </div>
          <div className="admin-list-actions">
            <div className="field admin-search">
              <label htmlFor="admin-event-search">ค้นหาอีเวนต์หรือสถานที่</label>
              <input
                className="input"
                id="admin-event-search"
                type="search"
                value={query}
                onChange={(event) => {
                  setQuery(event.target.value);
                  setPage(1);
                }}
                placeholder="ค้นหารายการ…"
              />
            </div>
            <button className="btn admin-add-button" type="button" onClick={openCreateModal}>+ เพิ่มอีเวนต์</button>
          </div>
        </div>

        {message && <p className="message" role="status">{message}</p>}

        {pageEvents.length ? (
          <>
            <div className="admin-event-table">
              <div className="admin-table-head" aria-hidden="true">
                <span>อีเวนต์</span>
                <span>วันและเวลา</span>
                <span>จำนวนที่นั่ง</span>
                <span>จัดการ</span>
              </div>
              {pageEvents.map((item) => {
                const availablePercent = Math.max(0, Math.min(100, (item.available_seats / item.total_seats) * 100));
                return (
                  <article className="admin-event-row" key={item.id}>
                    <div className="admin-event-name">
                      <h3>{item.title}</h3>
                      <span className="meta">{item.location || "ยังไม่ระบุสถานที่"}</span>
                    </div>
                    <time className="meta" dateTime={item.starts_at}>{formatEventDate(item.starts_at)}</time>
                    <div className="admin-capacity">
                      <div className="admin-capacity-label">
                        <span className="seats">เหลือ {item.available_seats}</span>
                        <span className="meta">/ {item.total_seats}</span>
                      </div>
                      <div className="admin-capacity-track" aria-hidden="true">
                        <span style={{ width: `${availablePercent}%` }} />
                      </div>
                    </div>
                    <button className="btn secondary admin-edit-button" type="button" onClick={() => openEditModal(item)}>แก้ไข</button>
                  </article>
                );
              })}
            </div>

            {totalPages > 1 && (
              <nav className="admin-pagination" aria-label="หน้ารายการอีเวนต์">
                <button className="btn secondary" type="button" disabled={currentPage === 1} onClick={() => setPage(currentPage - 1)}>ก่อนหน้า</button>
                <div className="admin-page-numbers">
                  {Array.from({ length: totalPages }, (_, index) => index + 1).map((pageNumber) => (
                    <button
                      className={`admin-page-button${pageNumber === currentPage ? " active" : ""}`}
                      type="button"
                      aria-label={`หน้า ${pageNumber}`}
                      aria-current={pageNumber === currentPage ? "page" : undefined}
                      key={pageNumber}
                      onClick={() => setPage(pageNumber)}
                    >
                      {pageNumber}
                    </button>
                  ))}
                </div>
                <button className="btn secondary" type="button" disabled={currentPage === totalPages} onClick={() => setPage(currentPage + 1)}>ถัดไป</button>
              </nav>
            )}
          </>
        ) : (
          <div className="empty"><h3>ไม่พบอีเวนต์</h3><p>ลองค้นหาด้วยชื่อหรือสถานที่อื่น</p></div>
        )}
      </section>

      {modalOpen && (
        <div className="admin-modal-backdrop" role="presentation" onMouseDown={closeFromBackdrop}>
          <form
            className="admin-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="event-modal-title"
            onSubmit={submit}
          >
            <header className="admin-modal-header">
              <div>
                <span className="eyebrow" style={{ color: "var(--brand)" }}>{form.id ? "Edit event" : "New event"}</span>
                <h2 id="event-modal-title">{form.id ? "แก้ไขอีเวนต์" : "เพิ่มอีเวนต์ใหม่"}</h2>
              </div>
              <button className="admin-modal-close" type="button" aria-label="ปิดหน้าต่าง" disabled={pending} onClick={closeModal}>×</button>
            </header>

            <div className="admin-modal-body">
              {error && <p className="message error" role="alert">{error}</p>}
              <div className="admin-form-layout">
                <div className="field admin-span-6">
                  <label htmlFor="title">ชื่ออีเวนต์</label>
                  <input className="input" id="title" autoFocus value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} required />
                </div>
                <div className="field admin-span-6">
                  <label htmlFor="location">สถานที่</label>
                  <input className="input" id="location" value={form.location} onChange={(event) => setForm({ ...form, location: event.target.value })} />
                </div>
                <div className="field admin-span-12">
                  <label htmlFor="description">รายละเอียด</label>
                  <textarea className="input" id="description" value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} />
                </div>
                <div className="field admin-span-6">
                  <label htmlFor="starts_at">วันและเวลา</label>
                  <input className="input" id="starts_at" type="datetime-local" value={form.starts_at} onChange={(event) => setForm({ ...form, starts_at: event.target.value })} required />
                </div>
                <div className="field admin-span-6">
                  <label htmlFor="total_seats">ที่นั่งทั้งหมด</label>
                  <input className="input" id="total_seats" type="number" min={1} value={form.total_seats} onChange={(event) => setForm({ ...form, total_seats: Number(event.target.value) })} required />
                </div>
              </div>
            </div>

            <footer className="admin-modal-footer">
              <button className="btn secondary" type="button" disabled={pending} onClick={closeModal}>ยกเลิก</button>
              <button className="btn" disabled={pending}>{pending ? "กำลังบันทึก…" : form.id ? "บันทึกการแก้ไข" : "เพิ่มอีเวนต์"}</button>
            </footer>
          </form>
        </div>
      )}
    </div>
  );
}
