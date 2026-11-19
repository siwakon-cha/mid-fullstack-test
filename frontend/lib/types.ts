export type User = {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  phone: string | null;
  role: 1 | 2;
};

export type EventItem = {
  id: number;
  title: string;
  description: string;
  location: string;
  starts_at: string;
  total_seats: number;
  available_seats: number;
};

export type EventListResponse = {
  items: EventItem[];
  page: number;
  page_size: number;
  total: number;
};

export type BookingItem = {
  id: number;
  event_id: number;
  quantity: number;
  status: 1 | 2;
  cancelled_at: string | null;
  event_title: string;
  event_starts_at: string;
  event_location: string;
};

export type APIError = { error: { code: string; message: string } };
