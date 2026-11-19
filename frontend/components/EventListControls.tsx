"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import type { FormEvent } from "react";

type Props = {
  search: string;
};

export default function EventListControls({ search }: Props) {
  const router = useRouter();

  function handleSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const nextSearch = String(formData.get("search") ?? "").trim();

    router.push(`/events?page=1&search=${encodeURIComponent(nextSearch)}`, {
      scroll: false,
    });
  }

  return (
    <form className="toolbar" onSubmit={handleSearch}>
      <div className="field">
        <label htmlFor="search">ค้นหาจากคำขึ้นต้นของชื่อ</label>
        <input
          className="input"
          id="search"
          name="search"
          defaultValue={search}
          placeholder="เช่น Bangkok"
          maxLength={80}
        />
      </div>
      <button className="btn" type="submit">ค้นหา</button>
      {search && (
        <Link className="btn secondary" href="/events" scroll={false}>
          ล้างคำค้น
        </Link>
      )}
    </form>
  );
}

type PaginationProps = Props & {
  page: number;
  pages: number;
};

export function EventPagination({ page, pages, search }: PaginationProps) {
  const pageLink = (nextPage: number) =>
    `/events?page=${nextPage}&search=${encodeURIComponent(search)}`;

  if (pages <= 1) return null;

  return (
    <nav className="pagination" aria-label="เปลี่ยนหน้า">
      {page > 1 ? (
        <Link className="btn secondary" href={pageLink(page - 1)} scroll={false}>
          ก่อนหน้า
        </Link>
      ) : (
        <span />
      )}
      <span className="meta">หน้า {page} / {pages}</span>
      {page < pages ? (
        <Link className="btn secondary" href={pageLink(page + 1)} scroll={false}>
          ถัดไป
        </Link>
      ) : (
        <span />
      )}
    </nav>
  );
}
