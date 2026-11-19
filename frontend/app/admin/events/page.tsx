import AdminEvents from "@/components/AdminEvents";

export default function AdminEventsPage() {
  return (
    <main className="shell page admin-page">
      <header className="admin-page-header">
        <div>
          <span className="eyebrow" style={{ color: "var(--brand)" }}>Admin console</span>
          <h1>จัดการอีเวนต์</h1>
        </div>
        <p>สร้าง แก้ไข และติดตามจำนวนที่นั่งของทุกอีเวนต์</p>
      </header>
      <AdminEvents />
    </main>
  );
}
