import Link from "next/link";

export default function NotFound() {
  return <main className="shell page"><div className="empty"><h2>ไม่พบอีเวนต์</h2><p>อีเวนต์นี้อาจถูกลบหรือ URL ไม่ถูกต้อง</p><Link className="btn" href="/events">กลับไปหน้ารายการ</Link></div></main>;
}
